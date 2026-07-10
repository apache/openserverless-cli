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

package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const (
	defaultSSONamespace = "nuvolaris"
	defaultSSOConfigMap = "openserverless-sso-config"
	defaultSSOSecret    = "openserverless-sso-secret"
	defaultSSOWorkload  = "nuvolaris-system-api"
	defaultSSOContainer = "nuvolaris-system-api"
)

type commandRunner func(name string, args []string, stdin []byte) ([]byte, error)

var runSSOCommand commandRunner = realCommandRunner

type ssoOptions struct {
	IssuerURL              string
	JWKSURL                string
	Audience               string
	ClientID               string
	ClientSecret           string
	RequiredGroup          string
	UsernameClaim          string
	GroupsClaim            string
	AutoProvision          bool
	AutoProvisionTimeout   string
	AutoProvisionPoll      string
	AutoProvisionServices  string
	NamespacePreserveValid bool
	NamespaceHashLength    string
	NamespaceMaxLength     string
	Namespace              string
	ConfigMapName          string
	SecretName             string
	WorkloadName           string
	ContainerName          string
	NoRollout              bool
}

func printSSOUsage() {
	fmt.Print(`Usage:
ops config sso keycloak --enable --issuer-url URL --jwks-url URL (--audience AUDIENCE|--client-id CLIENT_ID) --required-group GROUP [options]
ops config sso show
ops config sso disable [options]

Legacy embedded form:
ops -config sso keycloak --enable --issuer-url URL --jwks-url URL (--audience AUDIENCE|--client-id CLIENT_ID) --required-group GROUP [options]
ops -config sso show
ops -config sso disable [options]

Configure OpenServerless SSO/OIDC integration for admin-api.

Options:
  --username-claim CLAIM   OIDC username claim. Default: preferred_username
  --groups-claim CLAIM     OIDC groups claim. Default: groups
  --client-id CLIENT_ID    OIDC client id. Defaults to --audience when omitted
  --client-secret SECRET   OIDC confidential client secret stored only in Kubernetes Secret
  --namespace NS           Kubernetes namespace. Default: nuvolaris
  --configmap NAME         Kubernetes ConfigMap name. Default: openserverless-sso-config
  --secret NAME            Kubernetes Secret name. Default: openserverless-sso-secret
  --statefulset NAME       admin-api StatefulSet name. Default: nuvolaris-system-api
  --container NAME         admin-api container name. Default: nuvolaris-system-api
  --no-rollout             Do not restart or wait for admin-api rollout
`)
}

func ConfigSSOTool(configMap ConfigMap, args []string) error {
	if len(args) == 0 {
		printSSOUsage()
		return nil
	}

	switch args[0] {
	case "keycloak":
		return configureKeycloakSSO(configMap, args[1:])
	case "show":
		printSSOConfig(configMap)
		return nil
	case "disable":
		return disableSSO(configMap, args[1:])
	case "-h", "--help", "help":
		printSSOUsage()
		return nil
	default:
		return fmt.Errorf("unknown sso command: %s", args[0])
	}
}

func configureKeycloakSSO(configMap ConfigMap, args []string) error {
	opts, err := parseKeycloakSSOArgs(args)
	if err != nil {
		return err
	}

	if err := saveSSOConfig(configMap, opts); err != nil {
		return err
	}

	if err := applySSOConfigMap(opts); err != nil {
		return err
	}

	if opts.ClientSecret != "" {
		if err := applySSOSecret(opts); err != nil {
			return err
		}
	}

	if err := patchSSOEnvFrom(opts); err != nil {
		return err
	}

	if !opts.NoRollout {
		if err := rolloutSSOWorkload(opts); err != nil {
			return err
		}
	}

	fmt.Println("SSO configuration applied to admin-api.")
	fmt.Printf("ConfigMap: %s/%s\n", opts.Namespace, opts.ConfigMapName)
	if opts.ClientSecret != "" {
		fmt.Printf("Secret: %s/%s\n", opts.Namespace, opts.SecretName)
	}
	return nil
}

func parseKeycloakSSOArgs(args []string) (ssoOptions, error) {
	opts := ssoOptions{
		UsernameClaim:          "preferred_username",
		GroupsClaim:            "groups",
		AutoProvision:          true,
		AutoProvisionTimeout:   "120",
		AutoProvisionPoll:      "2",
		AutoProvisionServices:  "all",
		NamespacePreserveValid: true,
		NamespaceHashLength:    "8",
		NamespaceMaxLength:     "61",
		Namespace:              defaultSSONamespace,
		ConfigMapName:          defaultSSOConfigMap,
		SecretName:             defaultSSOSecret,
		WorkloadName:           defaultSSOWorkload,
		ContainerName:          defaultSSOContainer,
	}

	flags := flag.NewFlagSet("sso keycloak", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	enable := flags.Bool("enable", false, "enable SSO")
	flags.StringVar(&opts.IssuerURL, "issuer-url", "", "OIDC issuer URL")
	flags.StringVar(&opts.JWKSURL, "jwks-url", "", "OIDC JWKS URL")
	flags.StringVar(&opts.Audience, "audience", "", "OIDC audience")
	flags.StringVar(&opts.ClientID, "client-id", "", "OIDC client id")
	flags.StringVar(&opts.ClientSecret, "client-secret", "", "OIDC confidential client secret")
	flags.StringVar(&opts.RequiredGroup, "required-group", "", "required OIDC group")
	flags.StringVar(&opts.UsernameClaim, "username-claim", opts.UsernameClaim, "OIDC username claim")
	flags.StringVar(&opts.GroupsClaim, "groups-claim", opts.GroupsClaim, "OIDC groups claim")
	flags.StringVar(&opts.Namespace, "namespace", opts.Namespace, "Kubernetes namespace")
	flags.StringVar(&opts.ConfigMapName, "configmap", opts.ConfigMapName, "Kubernetes ConfigMap name")
	flags.StringVar(&opts.SecretName, "secret", opts.SecretName, "Kubernetes Secret name")
	flags.StringVar(&opts.WorkloadName, "statefulset", opts.WorkloadName, "admin-api StatefulSet name")
	flags.StringVar(&opts.ContainerName, "container", opts.ContainerName, "admin-api container name")
	flags.BoolVar(&opts.NoRollout, "no-rollout", false, "skip rollout restart/status")

	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if !*enable {
		return opts, fmt.Errorf("missing --enable")
	}
	if flags.NArg() > 0 {
		return opts, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if opts.IssuerURL == "" {
		return opts, fmt.Errorf("missing --issuer-url")
	}
	if opts.JWKSURL == "" {
		return opts, fmt.Errorf("missing --jwks-url")
	}
	if opts.ClientID == "" {
		opts.ClientID = opts.Audience
	}
	if opts.Audience == "" {
		opts.Audience = opts.ClientID
	}
	if opts.Audience == "" {
		return opts, fmt.Errorf("missing --audience or --client-id")
	}
	if opts.RequiredGroup == "" {
		return opts, fmt.Errorf("missing --required-group")
	}
	if opts.UsernameClaim == "" {
		return opts, fmt.Errorf("missing --username-claim")
	}
	if opts.GroupsClaim == "" {
		return opts, fmt.Errorf("missing --groups-claim")
	}
	if opts.SecretName == "" {
		return opts, fmt.Errorf("missing --secret")
	}
	return opts, nil
}

func parseDisableSSOArgs(args []string) (ssoOptions, error) {
	opts := ssoOptions{
		Namespace:     defaultSSONamespace,
		ConfigMapName: defaultSSOConfigMap,
		SecretName:    defaultSSOSecret,
		WorkloadName:  defaultSSOWorkload,
		ContainerName: defaultSSOContainer,
	}

	flags := flag.NewFlagSet("sso disable", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&opts.Namespace, "namespace", opts.Namespace, "Kubernetes namespace")
	flags.StringVar(&opts.ConfigMapName, "configmap", opts.ConfigMapName, "Kubernetes ConfigMap name")
	flags.StringVar(&opts.SecretName, "secret", opts.SecretName, "Kubernetes Secret name")
	flags.StringVar(&opts.WorkloadName, "statefulset", opts.WorkloadName, "admin-api StatefulSet name")
	flags.StringVar(&opts.ContainerName, "container", opts.ContainerName, "admin-api container name")
	flags.BoolVar(&opts.NoRollout, "no-rollout", false, "skip rollout restart/status")

	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if flags.NArg() > 0 {
		return opts, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	return opts, nil
}

func saveSSOConfig(configMap ConfigMap, opts ssoOptions) error {
	values := map[string]string{
		"SSO_ENABLED":                        "true",
		"SSO_PROVIDER":                       "keycloak",
		"SSO_OIDC_ISSUER_URL":                opts.IssuerURL,
		"SSO_OIDC_JWKS_URL":                  opts.JWKSURL,
		"SSO_OIDC_AUDIENCE":                  opts.Audience,
		"SSO_OIDC_CLIENT_ID":                 opts.ClientID,
		"SSO_OIDC_REQUIRED_GROUP":            opts.RequiredGroup,
		"SSO_OIDC_USERNAME_CLAIM":            opts.UsernameClaim,
		"SSO_OIDC_GROUPS_CLAIM":              opts.GroupsClaim,
		"SSO_OIDC_CLIENT_SECRET_CONFIGURED":  fmt.Sprintf("%t", opts.ClientSecret != ""),
		"SSO_CLIENT_MODE":                    ssoClientMode(opts),
		"SSO_AUTOPROVISION_ON_LOGIN":         fmt.Sprintf("%t", opts.AutoProvision),
		"SSO_AUTOPROVISION_TIMEOUT_SECONDS":  opts.AutoProvisionTimeout,
		"SSO_AUTOPROVISION_POLL_SECONDS":     opts.AutoProvisionPoll,
		"SSO_AUTOPROVISION_DEFAULT_SERVICES": opts.AutoProvisionServices,
		"SSO_NAMESPACE_PRESERVE_VALID":       fmt.Sprintf("%t", opts.NamespacePreserveValid),
		"SSO_NAMESPACE_HASH_LENGTH":          opts.NamespaceHashLength,
		"SSO_NAMESPACE_MAX_LENGTH":           opts.NamespaceMaxLength,
		"SSO_KUBE_NAMESPACE":                 opts.Namespace,
		"SSO_KUBE_CONFIGMAP":                 opts.ConfigMapName,
		"SSO_KUBE_SECRET":                    opts.SecretName,
		"SSO_KUBE_STATEFULSET":               opts.WorkloadName,
		"SSO_KUBE_CONTAINER":                 opts.ContainerName,
	}
	for key, value := range values {
		if err := configMap.Insert(key, value); err != nil {
			return err
		}
	}
	return configMap.SaveConfig()
}

func printSSOConfig(configMap ConfigMap) {
	values := configMap.Flatten()
	keys := make([]string, 0)
	for key := range values {
		if strings.HasPrefix(key, "SSO_") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s=%s\n", key, printableSSOValue(key, values[key]))
	}
}

func disableSSO(configMap ConfigMap, args []string) error {
	opts, err := parseDisableSSOArgs(args)
	if err != nil {
		return err
	}

	if err := removeLocalSSOConfig(configMap); err != nil {
		return err
	}
	if err := removeSSOEnvFrom(opts); err != nil {
		return err
	}
	if err := deleteSSOConfigMap(opts); err != nil {
		return err
	}
	if err := deleteSSOSecret(opts); err != nil {
		return err
	}
	if !opts.NoRollout {
		if err := rolloutSSOWorkload(opts); err != nil {
			return err
		}
	}

	fmt.Println("SSO configuration disabled for admin-api.")
	return nil
}

func removeLocalSSOConfig(configMap ConfigMap) error {
	for key := range configMap.Flatten() {
		if !strings.HasPrefix(key, "SSO_") {
			continue
		}
		if err := configMap.Delete(key); err != nil && !strings.Contains(err.Error(), "does not exist in config.json") {
			return err
		}
	}
	return configMap.SaveConfig()
}

func applySSOConfigMap(opts ssoOptions) error {
	data := map[string]string{
		"OIDC_ISSUER_URL":                    opts.IssuerURL,
		"OIDC_JWKS_URL":                      opts.JWKSURL,
		"OIDC_AUDIENCE":                      opts.Audience,
		"OIDC_CLIENT_ID":                     opts.ClientID,
		"OIDC_REQUIRED_GROUP":                opts.RequiredGroup,
		"OIDC_USERNAME_CLAIM":                opts.UsernameClaim,
		"OIDC_GROUPS_CLAIM":                  opts.GroupsClaim,
		"SSO_AUTOPROVISION_ON_LOGIN":         fmt.Sprintf("%t", opts.AutoProvision),
		"SSO_AUTOPROVISION_TIMEOUT_SECONDS":  opts.AutoProvisionTimeout,
		"SSO_AUTOPROVISION_POLL_SECONDS":     opts.AutoProvisionPoll,
		"SSO_AUTOPROVISION_DEFAULT_SERVICES": opts.AutoProvisionServices,
		"SSO_NAMESPACE_PRESERVE_VALID":       fmt.Sprintf("%t", opts.NamespacePreserveValid),
		"SSO_NAMESPACE_HASH_LENGTH":          opts.NamespaceHashLength,
		"SSO_NAMESPACE_MAX_LENGTH":           opts.NamespaceMaxLength,
	}
	obj := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]string{
			"name":      opts.ConfigMapName,
			"namespace": opts.Namespace,
		},
		"data": data,
	}
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = runSSOCommand("kubectl", []string{"apply", "-f", "-"}, payload)
	return err
}

func applySSOSecret(opts ssoOptions) error {
	obj := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]string{
			"name":      opts.SecretName,
			"namespace": opts.Namespace,
		},
		"type": "Opaque",
		"stringData": map[string]string{
			"OIDC_CLIENT_SECRET": opts.ClientSecret,
		},
	}
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = runSSOCommand("kubectl", []string{"apply", "-f", "-"}, payload)
	return err
}

func patchSSOEnvFrom(opts ssoOptions) error {
	envFrom := []map[string]interface{}{
		{
			"configMapRef": map[string]string{
				"name": opts.ConfigMapName,
			},
		},
	}
	if opts.ClientSecret != "" {
		envFrom = append(envFrom, map[string]interface{}{
			"secretRef": map[string]string{
				"name": opts.SecretName,
			},
		})
	}
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":    opts.ContainerName,
							"envFrom": envFrom,
						},
					},
				},
			},
		},
	}
	payload, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	_, err = runSSOCommand("kubectl", []string{"-n", opts.Namespace, "patch", "statefulset", opts.WorkloadName, "--type=strategic", "-p", string(payload)}, nil)
	return err
}

func removeSSOEnvFrom(opts ssoOptions) error {
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":    opts.ContainerName,
							"envFrom": nil,
						},
					},
				},
			},
		},
	}
	payload, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	_, err = runSSOCommand("kubectl", []string{"-n", opts.Namespace, "patch", "statefulset", opts.WorkloadName, "--type=strategic", "-p", string(payload)}, nil)
	return err
}

func deleteSSOConfigMap(opts ssoOptions) error {
	_, err := runSSOCommand("kubectl", []string{"-n", opts.Namespace, "delete", "configmap", opts.ConfigMapName, "--ignore-not-found"}, nil)
	return err
}

func deleteSSOSecret(opts ssoOptions) error {
	_, err := runSSOCommand("kubectl", []string{"-n", opts.Namespace, "delete", "secret", opts.SecretName, "--ignore-not-found"}, nil)
	return err
}

func rolloutSSOWorkload(opts ssoOptions) error {
	if _, err := runSSOCommand("kubectl", []string{"-n", opts.Namespace, "rollout", "restart", "statefulset/" + opts.WorkloadName}, nil); err != nil {
		return err
	}
	_, err := runSSOCommand("kubectl", []string{"-n", opts.Namespace, "rollout", "status", "statefulset/" + opts.WorkloadName, "--timeout=180s"}, nil)
	return err
}

func ssoClientMode(opts ssoOptions) string {
	if opts.ClientSecret != "" {
		return "confidential"
	}
	return "public"
}

func printableSSOValue(key string, value interface{}) interface{} {
	if isSecretLikeKey(key) {
		return "<redacted>"
	}
	return value
}

func isSecretLikeKey(key string) bool {
	lower := strings.ToLower(key)
	for _, marker := range []string{"secret", "token", "credential", "password"} {
		if strings.Contains(lower, marker) && !strings.HasSuffix(lower, "_configured") {
			return true
		}
	}
	return false
}

func realCommandRunner(name string, args []string, stdin []byte) ([]byte, error) {
	cmdName := name
	if _, err := exec.LookPath(cmdName); err != nil && name == "kubectl" {
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			opsKubectl := home + "/.ops/linux-amd64/bin/kubectl"
			if _, statErr := os.Stat(opsKubectl); statErr == nil {
				cmdName = opsKubectl
			}
		}
	}

	cmd := exec.Command(cmdName, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := append(stdout.Bytes(), stderr.Bytes()...)
	if err != nil {
		if len(output) > 0 {
			return output, fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
		}
		return output, fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	if len(output) > 0 {
		fmt.Print(string(output))
	}
	return output, nil
}
