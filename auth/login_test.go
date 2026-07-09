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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/openserverless-cli/config"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

// setupMockServer sets up a new mock HTTP server with the given test data and expected response
func setupMockServer(t *testing.T, expectedLogin, expectedPass, expectedRes string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		login, ok := requestBody["login"]
		require.True(t, ok, "expected login field in request body")
		require.Equal(t, expectedLogin, login, "expected login %s, got %s", expectedLogin, login)

		password, ok := requestBody["password"]
		require.True(t, ok, "expected password field in request body")
		require.Equal(t, expectedPass, password, "expected password %s, got %s", expectedPass, password)

		_, _ = w.Write([]byte(expectedRes))
	}))

	return server
}
func TestLoginCmd(t *testing.T) {
	homeDir, _ := homedir.Expand("~/.ops")
	os.MkdirAll(homeDir, 0755)
	os.Setenv("OPS_HOME", homeDir)

	keyring.MockInit()

	t.Run("error: returns error when empty password", func(t *testing.T) {
		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "",
			ReturnError: false,
		}
		os.Args = []string{"login", "fakeApiHost", "fakeUser"}
		res, err := LoginCmd()
		pwdReader = oldPwdReader
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "password is empty" {
			t.Errorf("Expected error to be 'password is empty', got %s", err.Error())
		}
		if res != nil {
			t.Errorf("Expected response to be nil, got %v", res)
		}
	})

	t.Run("with only apihost, add received credentials", func(t *testing.T) {
		mockServer := setupMockServer(t, "nuvolaris", "a password", "{\"AUTH\": \"test\"}")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}

		os.Args = []string{"login", mockServer.URL}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader

		require.NoError(t, err)
		require.NotNil(t, loginResult)

		// cred, err := keyring.Get(opsSecretServiceName, "AUTH")
		opsHome, err := homedir.Expand("~/.ops")
		require.NoError(t, err)

		configMap, err := config.NewConfigMapBuilder().
			WithConfigJson(filepath.Join(opsHome, "config.json")).
			Build()
		require.NoError(t, err)

		v, err := configMap.Get("AUTH")
		require.NoError(t, err)
		require.Equal(t, "test", v)
	})

	t.Run("with apihost and user adds received credentials to secret store", func(t *testing.T) {
		mockServer := setupMockServer(t, "a user", "a password", "{ \"AUTH\": \"testauth\", \"fakeCred\": \"test\"}")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}

		os.Args = []string{"login", mockServer.URL, "a user"}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader
		require.NoError(t, err)
		require.NotNil(t, loginResult)

		// cred, err := keyring.Get(opsSecretServiceName, "fakeCred")
		// require.NoError(t, err)
		// require.Equal(t, "test", cred)

		opsHome, err := homedir.Expand("~/.ops")
		require.NoError(t, err)

		configMap, err := config.NewConfigMapBuilder().
			WithConfigJson(filepath.Join(opsHome, "config.json")).
			Build()
		require.NoError(t, err)

		v, err := configMap.Get("AUTH")
		require.NoError(t, err)
		require.Equal(t, "testauth", v)

		v, err = configMap.Get("FAKECRED")
		require.NoError(t, err)
		require.Equal(t, "test", v)
	})

	t.Run("error when response body is invalid", func(t *testing.T) {
		mockServer := setupMockServer(t, "a user", "a password", "invalid json")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}
		os.Args = []string{"login", mockServer.URL, "a user"}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader
		require.Error(t, err)
		require.Equal(t, "failed to decode response from login request", err.Error())
		require.Nil(t, loginResult)
	})

	t.Run("error when response body is missing AUTH token", func(t *testing.T) {
		mockServer := setupMockServer(t, "a user", "a password", "{\"fakeCred\": \"test\"}")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}
		os.Args = []string{"login", mockServer.URL, "a user"}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader
		require.Error(t, err)
		require.Equal(t, "missing AUTH token from login response", err.Error())
		require.Nil(t, loginResult)
	})

	t.Run("SSO enabled starts OIDC device flow", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("OPS_HOME", tmpDir)
		t.Setenv("SSO_ENABLED", "true")
		t.Setenv("SSO_OIDC_AUDIENCE", "openserverless-admin-api")
		t.Setenv("OPS_SSO_DISABLE_BROWSER", "true")
		t.Setenv("OPS_PASSWORD", "")
		t.Setenv("OPS_USER", "")
		t.Setenv("OPS_APIHOST", "")

		var mockServer *httptest.Server
		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/realms/lab/.well-known/openid-configuration":
				_, _ = w.Write([]byte(fmt.Sprintf(`{
					"device_authorization_endpoint": "%s/realms/lab/protocol/openid-connect/auth/device",
					"token_endpoint": "%s/realms/lab/protocol/openid-connect/token"
				}`, mockServer.URL, mockServer.URL)))
			case "/realms/lab/protocol/openid-connect/auth/device":
				require.NoError(t, r.ParseForm())
				require.Equal(t, "openserverless-admin-api", r.Form.Get("client_id"))
				require.Equal(t, "S256", r.Form.Get("code_challenge_method"))
				require.NotEmpty(t, r.Form.Get("code_challenge"))
				_, _ = w.Write([]byte(`{
					"device_code": "device-code",
					"user_code": "ABCD-EFGH",
					"verification_uri": "http://localhost/device",
					"verification_uri_complete": "http://localhost/device?user_code=ABCD-EFGH",
					"expires_in": 10,
					"interval": 1
				}`))
			case "/realms/lab/protocol/openid-connect/token":
				require.NoError(t, r.ParseForm())
				require.Equal(t, "urn:ietf:params:oauth:grant-type:device_code", r.Form.Get("grant_type"))
				require.Equal(t, "openserverless-admin-api", r.Form.Get("client_id"))
				require.Equal(t, "device-code", r.Form.Get("device_code"))
				require.NotEmpty(t, r.Form.Get("code_verifier"))
				_, _ = w.Write([]byte(`{"access_token":"device-access-token"}`))
			case "/system/api/v1/auth/oidc":
				require.Equal(t, "Bearer device-access-token", r.Header.Get("Authorization"))
				_, _ = w.Write([]byte(`{"AUTH":"oidc-auth","NAMESPACE":"michelem"}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer mockServer.Close()

		t.Setenv("SSO_OIDC_ISSUER_URL", mockServer.URL+"/realms/lab")

		os.Args = []string{"login", mockServer.URL}
		loginResult, err := LoginCmd()
		require.NoError(t, err)
		require.NotNil(t, loginResult)
		require.Equal(t, "michelem", loginResult.Login)
		require.Equal(t, "oidc-auth", loginResult.Auth)
	})

	t.Run("SSO confidential client uses backend managed device flow", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("OPS_HOME", tmpDir)
		t.Setenv("SSO_ENABLED", "true")
		t.Setenv("SSO_CLIENT_MODE", "confidential")
		t.Setenv("OPS_SSO_DISABLE_BROWSER", "true")
		t.Setenv("OPS_PASSWORD", "")
		t.Setenv("OPS_USER", "")
		t.Setenv("OPS_APIHOST", "")

		pollCount := 0
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/system/api/v1/auth/oidc/device/start":
				require.Equal(t, http.MethodPost, r.Method)
				var payload map[string]string
				require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
				require.Equal(t, "michelem", payload["namespace"])
				_, _ = w.Write([]byte(`{
					"flow_id": "flow-1",
					"user_code": "ABCD-EFGH",
					"verification_uri_complete": "http://localhost/device?user_code=ABCD-EFGH",
					"expires_in": 10,
					"interval": 1
				}`))
			case "/system/api/v1/auth/oidc/device/poll":
				require.Equal(t, http.MethodPost, r.Method)
				var payload map[string]string
				require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
				require.Equal(t, "flow-1", payload["flow_id"])
				pollCount++
				_, _ = w.Write([]byte(`{"AUTH":"oidc-auth","NAMESPACE":"michelem"}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer mockServer.Close()

		os.Args = []string{"login", mockServer.URL, "michelem"}
		loginResult, err := LoginCmd()
		require.NoError(t, err)
		require.NotNil(t, loginResult)
		require.Equal(t, "michelem", loginResult.Login)
		require.Equal(t, "oidc-auth", loginResult.Auth)
		require.Equal(t, 1, pollCount)
	})

	t.Run("SSO enabled fails without issuer", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("OPS_HOME", tmpDir)
		t.Setenv("SSO_ENABLED", "true")
		t.Setenv("SSO_OIDC_ISSUER_URL", "")
		t.Setenv("SSO_OIDC_AUDIENCE", "openserverless-admin-api")
		t.Setenv("OPS_PASSWORD", "")
		t.Setenv("OPS_USER", "")
		t.Setenv("OPS_APIHOST", "")

		os.Args = []string{"login", "http://localhost:5000"}
		loginResult, err := LoginCmd()
		require.Error(t, err)
		require.Nil(t, loginResult)
		require.Contains(t, err.Error(), "SSO_OIDC_ISSUER_URL")
	})
}

func Test_doLogin(t *testing.T) {
	mockServer := setupMockServer(t, "a user", "a password", "{\"fakeCred\": \"test\"}")
	defer mockServer.Close()

	cred, err := doLogin(mockServer.URL, "a user", "a password")
	require.NoError(t, err)
	require.NotNil(t, cred)
	require.Equal(t, "test", cred["fakeCred"])
}

func Test_doOIDCLogin(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/system/api/v1/auth/oidc", r.URL.Path)
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var requestBody map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestBody))
		require.Equal(t, "test-token", requestBody["access_token"])

		_, _ = w.Write([]byte(`{"AUTH":"test-auth","NAMESPACE":"devel"}`))
	}))
	defer mockServer.Close()

	cred, err := doOIDCLogin(mockServer.URL+"/system/api/v1/auth/oidc", "test-token")
	require.NoError(t, err)
	require.NotNil(t, cred)
	require.Equal(t, "test-auth", cred["AUTH"])
	require.Equal(t, "devel", cred["NAMESPACE"])
}

func Test_storeCredentials(t *testing.T) {
	keyring.MockInit()

	fakeCreds := make(map[string]string)
	fakeCreds["AUTH"] = "fakeValue"
	fakeCreds["REDIS_URL"] = "fakeValue"
	fakeCreds["MONGODB"] = "fakeValue"

	err := storeCredentials(fakeCreds)
	require.NoError(t, err)
	require.NotNil(t, fakeCreds)

	for k := range fakeCreds {
		cred, err := keyring.Get(opsSecretServiceName, k)
		require.NoError(t, err)
		require.Equal(t, fakeCreds[k], cred)
	}
}
