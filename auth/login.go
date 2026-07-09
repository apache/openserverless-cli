// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apache/openserverless-cli/config"
	"github.com/pkg/browser"
	"github.com/zalando/go-keyring"
)

type LoginResult struct {
	Login   string
	Auth    string
	ApiHost string
}

type oidcDiscovery struct {
	TokenEndpoint               string `json:"token_endpoint"`
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
}

type deviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	Error                   string `json:"error"`
	ErrorDescription        string `json:"error_description"`
}

type oidcTokenResponse struct {
	AccessToken      string `json:"access_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type backendDeviceStartResponse struct {
	FlowID                  string `json:"flow_id"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type backendDevicePollResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Interval int    `json:"interval"`
}

const usage = `Usage:
ops -login <apihost> [<user>]

Login to an OpenServerless instance. If no user is specified, the default user "nuvolaris" is used.
You can set the environment variables OPS_APIHOST and OPS_USER to avoid specifying them on the command line.
You can set OPS_PASSWORD to avoid entering the password interactively.

Options:
  -h, --help   Show usage`

const whiskLoginPath = "/api/v1/web/whisk-system/nuv/login"
const oidcLoginPath = "/system/api/v1/auth/oidc"
const oidcDeviceStartPath = "/system/api/v1/auth/oidc/device/start"
const oidcDevicePollPath = "/system/api/v1/auth/oidc/device/poll"
const defaultUser = "nuvolaris"
const opsSecretServiceName = "nuvolaris"

func LoginCmd() (*LoginResult, error) {

	// enable log output if requested
	if os.Getenv("DEBUG")+os.Getenv("TRACE") != "" {
		log.SetOutput(os.Stdout)
	}

	flag := flag.NewFlagSet("-login", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Println(usage)
	}

	var helpFlag bool
	flag.BoolVar(&helpFlag, "h", false, "Show usage")
	flag.BoolVar(&helpFlag, "help", false, "Show usage")
	err := flag.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if helpFlag {
		flag.Usage()
		return nil, nil
	}

	args := flag.Args()

	if len(args) == 0 && os.Getenv("OPS_APIHOST") == "" {
		flag.Usage()
		return nil, errors.New("missing apihost")
	}

	apihost := os.Getenv("OPS_APIHOST")
	if apihost == "" {
		apihost = args[0]
	}
	apihost = ensureSchema(apihost)
	passwordLoginURL := apihost + whiskLoginPath
	oidcLoginURL := apihost + oidcLoginPath

	// try to get the user from the environment
	user := os.Getenv("OPS_USER")
	if user == "" {
		// if env var not set, try to get it from the command line
		if os.Getenv("OPS_APIHOST") != "" {
			// if apihost env var was set, treat the first arg as the user
			if len(args) > 0 {
				user = args[0]
			}
		} else {
			// if apihost env var was not set, treat the second arg as the user
			if len(args) > 1 {
				user = args[1]
			}
		}
	}

	var creds map[string]string
	ssoEnabled := isTruthy(os.Getenv("SSO_ENABLED"))
	var oidcToken string
	if ssoEnabled {
		if useBackendManagedOIDCDeviceFlow() {
			fmt.Println("Logging in", apihost, "with backend-managed OIDC")
			creds, err = backendManagedOIDCDeviceLogin(apihost, user)
			if err != nil {
				return nil, err
			}
			user = loginFromCredentials(creds, user)
		} else {
			oidcToken, err = oidcDeviceAccessToken()
			if err != nil {
				return nil, err
			}
		}
	}
	if creds == nil && oidcToken != "" {
		fmt.Println("Logging in", apihost, "with OIDC")
		creds, err = doOIDCLogin(oidcLoginURL, oidcToken)
		if err != nil {
			return nil, err
		}
		user = loginFromCredentials(creds, user)
	} else if creds == nil {
		if ssoEnabled {
			return nil, errors.New("SSO is enabled but OIDC login did not return an access token")
		}
		// if still not set, use the default user
		if user == "" {
			fmt.Println("Using the default user:", defaultUser)
			user = defaultUser
		}

		fmt.Println("Logging in", apihost, "as", user)

		password := os.Getenv("OPS_PASSWORD")
		if password == "" {
			fmt.Print("Enter Password: ")
			pwd, err := AskPassword()
			if err != nil {
				return nil, err
			}
			password = pwd
			fmt.Println()
		}

		creds, err = doLogin(passwordLoginURL, user, password)
		if err != nil {
			return nil, err
		}
	}

	if _, ok := creds["AUTH"]; !ok {
		return nil, errors.New("missing AUTH token from login response")
	}

	opsHome := os.Getenv("OPS_HOME")
	if opsHome == "" {
		return nil, fmt.Errorf("OPS_HOME not defined")
	}

	configMap, err := config.NewConfigMapBuilder().
		WithConfigJson(filepath.Join(opsHome, "config.json")).
		Build()

	if err != nil {
		return nil, err
	}

	for k, v := range creds {
		if err := configMap.Insert(k, v); err != nil {
			return nil, err
		}
	}

	if err := configMap.Insert("STATUS_LOGGED_USER", user); err != nil {
		log.Println("[Warning] Failed to insert STATUS_LOGGED_USER")
	}

	err = configMap.SaveConfig()
	if err != nil {
		return nil, err
	}

	// if err := storeCredentials(creds); err != nil {
	// 	return nil, err
	// }

	// auth, err := keyring.Get(opsSecretServiceName, "AUTH")
	// if err != nil {
	// 	return nil, err
	// }

	return &LoginResult{
		Login:   user,
		Auth:    creds["AUTH"],
		ApiHost: apihost,
	}, nil
}

func oidcDeviceAccessToken() (string, error) {
	issuer := strings.TrimRight(firstNonEmpty(os.Getenv("SSO_OIDC_ISSUER_URL"), os.Getenv("OIDC_ISSUER_URL")), "/")
	clientID := firstNonEmpty(os.Getenv("SSO_OIDC_AUDIENCE"), os.Getenv("OIDC_AUDIENCE"))
	if issuer == "" {
		return "", errors.New("SSO is enabled but SSO_OIDC_ISSUER_URL is not configured")
	}
	if clientID == "" {
		return "", errors.New("SSO is enabled but SSO_OIDC_AUDIENCE is not configured")
	}

	discovery, err := fetchOIDCDiscovery(issuer)
	if err != nil {
		return "", err
	}
	if discovery.DeviceAuthorizationEndpoint == "" {
		return "", errors.New("OIDC provider does not expose device_authorization_endpoint")
	}
	if discovery.TokenEndpoint == "" {
		return "", errors.New("OIDC provider does not expose token_endpoint")
	}

	codeVerifier, codeChallenge, err := pkceChallenge()
	if err != nil {
		return "", err
	}

	device, err := startOIDCDeviceAuthorization(discovery.DeviceAuthorizationEndpoint, clientID, codeChallenge)
	if err != nil {
		return "", err
	}

	verificationURL := device.VerificationURIComplete
	if verificationURL == "" {
		verificationURL = device.VerificationURI
	}
	fmt.Println()
	fmt.Println("SSO is enabled for this cluster.")
	fmt.Println("Open this URL in your browser to login:")
	fmt.Println(verificationURL)
	if device.UserCode != "" {
		fmt.Println("Code:", device.UserCode)
	}
	fmt.Println("Waiting for authentication...")

	if verificationURL != "" && !isTruthy(os.Getenv("OPS_SSO_DISABLE_BROWSER")) {
		_ = browser.OpenURL(verificationURL)
	}

	return pollOIDCDeviceToken(discovery.TokenEndpoint, clientID, device, codeVerifier)
}

func useBackendManagedOIDCDeviceFlow() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("SSO_CLIENT_MODE")), "confidential") ||
		isTruthy(firstNonEmpty(os.Getenv("SSO_OIDC_CLIENT_SECRET_CONFIGURED"), os.Getenv("OIDC_CLIENT_SECRET_CONFIGURED")))
}

func backendManagedOIDCDeviceLogin(apihost, requestedNamespace string) (map[string]string, error) {
	startURL := strings.TrimRight(apihost, "/") + oidcDeviceStartPath
	pollURL := strings.TrimRight(apihost, "/") + oidcDevicePollPath

	start, err := startBackendManagedOIDCDeviceFlow(startURL, requestedNamespace)
	if err != nil {
		return nil, err
	}
	if start.FlowID == "" {
		return nil, errors.New("SSO device login response missing flow_id")
	}
	if start.Interval <= 0 {
		start.Interval = 5
	}
	if start.ExpiresIn <= 0 {
		start.ExpiresIn = 600
	}

	verificationURL := start.VerificationURIComplete
	if verificationURL == "" {
		verificationURL = start.VerificationURI
	}
	fmt.Println()
	fmt.Println("SSO is enabled for this cluster.")
	fmt.Println("Open this URL in your browser to login:")
	fmt.Println(verificationURL)
	if start.UserCode != "" {
		fmt.Println("Code:", start.UserCode)
	}
	fmt.Println("Waiting for authentication...")

	if verificationURL != "" && !isTruthy(os.Getenv("OPS_SSO_DISABLE_BROWSER")) {
		_ = browser.OpenURL(verificationURL)
	}

	return pollBackendManagedOIDCDeviceFlow(pollURL, start, requestedNamespace)
}

func startBackendManagedOIDCDeviceFlow(url, requestedNamespace string) (*backendDeviceStartResponse, error) {
	payload := map[string]string{}
	if strings.TrimSpace(requestedNamespace) != "" {
		payload["namespace"] = strings.TrimSpace(requestedNamespace)
	}
	startJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(startJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC device authorization failed (%d): %s", resp.StatusCode, string(body))
	}

	var start backendDeviceStartResponse
	if err := json.Unmarshal(body, &start); err != nil {
		return nil, errors.New("failed to decode OIDC device authorization response")
	}
	return &start, nil
}

func pollBackendManagedOIDCDeviceFlow(url string, start *backendDeviceStartResponse, requestedNamespace string) (map[string]string, error) {
	deadline := time.Now().Add(time.Duration(start.ExpiresIn) * time.Second)
	interval := time.Duration(start.Interval) * time.Second

	for {
		if time.Now().After(deadline) {
			return nil, errors.New("OIDC device login expired")
		}
		time.Sleep(interval)

		payload := map[string]string{"flow_id": start.FlowID}
		if strings.TrimSpace(requestedNamespace) != "" {
			payload["namespace"] = strings.TrimSpace(requestedNamespace)
		}
		loginJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(loginJSON))
		if err != nil {
			return nil, err
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}

		if resp.StatusCode == http.StatusAccepted {
			var pending backendDevicePollResponse
			if err := json.Unmarshal(body, &pending); err == nil && pending.Interval > 0 {
				interval = time.Duration(pending.Interval) * time.Second
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("OIDC token polling failed (%d): %s", resp.StatusCode, string(body))
		}

		var creds map[string]string
		if err := json.Unmarshal(body, &creds); err != nil {
			return nil, errors.New("failed to decode response from OIDC device login request")
		}
		return creds, nil
	}
}

func fetchOIDCDiscovery(issuer string) (*oidcDiscovery, error) {
	resp, err := http.Get(issuer + "/.well-known/openid-configuration")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OIDC discovery failed (%d): %s", resp.StatusCode, string(body))
	}

	var discovery oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, errors.New("failed to decode OIDC discovery response")
	}
	return &discovery, nil
}

func startOIDCDeviceAuthorization(endpoint, clientID, codeChallenge string) (*deviceAuthorizationResponse, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("scope", "openid email profile")
	form.Set("code_challenge", codeChallenge)
	form.Set("code_challenge_method", "S256")

	resp, err := http.PostForm(endpoint, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var device deviceAuthorizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return nil, errors.New("failed to decode OIDC device authorization response")
	}
	if resp.StatusCode != http.StatusOK {
		if device.Error != "" {
			return nil, fmt.Errorf("OIDC device authorization failed (%d): %s: %s", resp.StatusCode, device.Error, device.ErrorDescription)
		}
		return nil, fmt.Errorf("OIDC device authorization failed with status code %d", resp.StatusCode)
	}
	if device.DeviceCode == "" {
		return nil, errors.New("OIDC device authorization response missing device_code")
	}
	if device.Interval <= 0 {
		device.Interval = 5
	}
	if device.ExpiresIn <= 0 {
		device.ExpiresIn = 600
	}
	return &device, nil
}

func pollOIDCDeviceToken(tokenEndpoint, clientID string, device *deviceAuthorizationResponse, codeVerifier string) (string, error) {
	deadline := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second)
	interval := time.Duration(device.Interval) * time.Second

	for {
		if time.Now().After(deadline) {
			return "", errors.New("OIDC device login expired")
		}
		time.Sleep(interval)

		form := url.Values{}
		form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		form.Set("client_id", clientID)
		form.Set("device_code", device.DeviceCode)
		form.Set("code_verifier", codeVerifier)

		resp, err := http.PostForm(tokenEndpoint, form)
		if err != nil {
			return "", err
		}

		var token oidcTokenResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&token)
		resp.Body.Close()
		if decodeErr != nil {
			return "", errors.New("failed to decode OIDC token response")
		}

		if resp.StatusCode == http.StatusOK && token.AccessToken != "" {
			return token.AccessToken, nil
		}

		switch token.Error {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5 * time.Second
			continue
		case "access_denied":
			return "", errors.New("OIDC device login denied")
		case "expired_token":
			return "", errors.New("OIDC device login expired")
		default:
			if token.Error != "" {
				return "", fmt.Errorf("OIDC token polling failed: %s: %s", token.Error, token.ErrorDescription)
			}
			return "", fmt.Errorf("OIDC token polling failed with status code %d", resp.StatusCode)
		}
	}
}

func pkceChallenge() (string, string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(random)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func loginFromCredentials(creds map[string]string, fallback string) string {
	for _, key := range []string{"NAMESPACE", "LOGIN", "USER", "USERNAME"} {
		if value := strings.TrimSpace(creds[key]); value != "" {
			return value
		}
	}
	return fallback
}

func ensureSchema(apihost string) string {
	if !strings.HasPrefix(apihost, "http://") && !strings.HasPrefix(apihost, "https://") {
		if apihost == "localhost" {
			apihost = "http://" + apihost
		} else {
			apihost = "https://" + apihost
		}
	}
	return apihost
}

func doLogin(url, user, password string) (map[string]string, error) {
	data := map[string]string{
		"login":    user,
		"password": password,
	}
	loginJson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(loginJson))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("login failed with status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("login failed (%d): %s", resp.StatusCode, string(body))
	}

	var creds map[string]string
	err = json.NewDecoder(resp.Body).Decode(&creds)
	if err != nil {
		return nil, errors.New("failed to decode response from login request")
	}

	return creds, nil
}

func doOIDCLogin(url, accessToken string) (map[string]string, error) {
	token := strings.TrimSpace(accessToken)
	if token == "" {
		return nil, errors.New("missing OIDC access token")
	}

	data := map[string]string{
		"access_token": strings.TrimPrefix(token, "Bearer "),
	}
	loginJson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(loginJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.HasPrefix(token, "Bearer ") {
		req.Header.Set("Authorization", token)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("OIDC login failed with status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("OIDC login failed (%d): %s", resp.StatusCode, string(body))
	}

	var creds map[string]string
	err = json.NewDecoder(resp.Body).Decode(&creds)
	if err != nil {
		return nil, errors.New("failed to decode response from OIDC login request")
	}

	return creds, nil
}

func storeCredentials(creds map[string]string) error {
	for k, v := range creds {
		err := keyring.Set(opsSecretServiceName, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
