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

type ssoEnvFromSource struct {
	Prefix       string                   `json:"prefix,omitempty"`
	ConfigMapRef *ssoLocalObjectReference `json:"configMapRef,omitempty"`
	SecretRef    *ssoLocalObjectReference `json:"secretRef,omitempty"`
}

type ssoLocalObjectReference struct {
	Name     string `json:"name"`
	Optional *bool  `json:"optional,omitempty"`
}

type ssoWorkload struct {
	Spec struct {
		Template struct {
			Spec struct {
				Containers []struct {
					Name    string             `json:"name"`
					EnvFrom []ssoEnvFromSource `json:"envFrom,omitempty"`
				} `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
}

type jsonPatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var managedLocalSSOKeys = []string{
	"SSO_ENABLED",
	"SSO_PROVIDER",
	"SSO_OIDC_ISSUER_URL",
	"SSO_OIDC_JWKS_URL",
	"SSO_OIDC_AUDIENCE",
	"SSO_OIDC_CLIENT_ID",
	"SSO_OIDC_REQUIRED_GROUP",
	"SSO_OIDC_USERNAME_CLAIM",
	"SSO_OIDC_GROUPS_CLAIM",
	"SSO_OIDC_CLIENT_SECRET_CONFIGURED",
	"SSO_CLIENT_MODE",
	"SSO_AUTOPROVISION_ON_LOGIN",
	"SSO_AUTOPROVISION_TIMEOUT_SECONDS",
	"SSO_AUTOPROVISION_POLL_SECONDS",
	"SSO_AUTOPROVISION_DEFAULT_SERVICES",
	"SSO_NAMESPACE_PRESERVE_VALID",
	"SSO_NAMESPACE_HASH_LENGTH",
	"SSO_NAMESPACE_MAX_LENGTH",
	"SSO_KUBE_NAMESPACE",
	"SSO_KUBE_CONFIGMAP",
	"SSO_KUBE_SECRET",
	"SSO_KUBE_STATEFULSET",
	"SSO_KUBE_CONTAINER",
}

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

Managed Kubernetes resources:
  ConfigMap NAME            OIDC_* and SSO_* values created by this command
  Secret NAME               OIDC_CLIENT_SECRET, when --client-secret is used
  admin-api container       Exact envFrom references to those two resources

The command does not manage direct env entries, other envFrom references,
volumes, volumeMounts, or workload annotations. Disable leaves them unchanged.

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

	if _, err := reconcileSSOEnvFrom(opts, true); err != nil {
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
	workloadChanged, err := reconcileSSOEnvFrom(opts, false)
	if err != nil {
		return err
	}
	if err := deleteSSOConfigMap(opts); err != nil {
		return err
	}
	if err := deleteSSOSecret(opts); err != nil {
		return err
	}
	if !opts.NoRollout && workloadChanged {
		if err := waitForSSOWorkloadRollout(opts); err != nil {
			return err
		}
	}

	fmt.Println("SSO configuration disabled for admin-api.")
	return nil
}

func removeLocalSSOConfig(configMap ConfigMap) error {
	values := configMap.Flatten()
	for _, key := range managedLocalSSOKeys {
		if _, exists := values[key]; !exists {
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

func reconcileSSOEnvFrom(opts ssoOptions, enabled bool) (bool, error) {
	output, err := runSSOCommand("kubectl", []string{
		"-n", opts.Namespace,
		"get", "statefulset", opts.WorkloadName,
		"-o", "json",
	}, nil)
	if err != nil {
		return false, err
	}

	var workload ssoWorkload
	if err := json.Unmarshal(output, &workload); err != nil {
		return false, fmt.Errorf("decode statefulset %s/%s: %w", opts.Namespace, opts.WorkloadName, err)
	}

	containerIndex := -1
	var current []ssoEnvFromSource
	for index, container := range workload.Spec.Template.Spec.Containers {
		if container.Name == opts.ContainerName {
			containerIndex = index
			current = container.EnvFrom
			break
		}
	}
	if containerIndex < 0 {
		return false, fmt.Errorf("container %s not found in statefulset %s/%s", opts.ContainerName, opts.Namespace, opts.WorkloadName)
	}

	desired := make([]ssoEnvFromSource, 0, 2)
	if enabled {
		desired = append(desired, ssoEnvFromSource{
			ConfigMapRef: &ssoLocalObjectReference{Name: opts.ConfigMapName},
		})
		if opts.ClientSecret != "" {
			desired = append(desired, ssoEnvFromSource{
				SecretRef: &ssoLocalObjectReference{Name: opts.SecretName},
			})
		}
	}

	seenDesired := make(map[string]bool)
	removeIndexes := make([]int, 0)
	for index, source := range current {
		key, managed := managedSSOEnvFromKey(source, opts)
		if !managed {
			continue
		}
		if enabled && desiredSSOEnvFromKey(key, desired) && !seenDesired[key] {
			seenDesired[key] = true
			continue
		}
		removeIndexes = append(removeIndexes, index)
	}

	missing := make([]ssoEnvFromSource, 0, len(desired))
	for _, source := range desired {
		key, _ := managedSSOEnvFromKey(source, opts)
		if !seenDesired[key] {
			missing = append(missing, source)
		}
	}
	if len(removeIndexes) == 0 && len(missing) == 0 {
		return false, nil
	}

	basePath := fmt.Sprintf("/spec/template/spec/containers/%d/envFrom", containerIndex)
	operations := make([]jsonPatchOperation, 0, len(removeIndexes)+len(missing))
	for index := len(removeIndexes) - 1; index >= 0; index-- {
		removeIndex := removeIndexes[index]
		operations = append(operations, jsonPatchOperation{
			Op:    "test",
			Path:  fmt.Sprintf("%s/%d", basePath, removeIndex),
			Value: current[removeIndex],
		})
		operations = append(operations, jsonPatchOperation{
			Op:   "remove",
			Path: fmt.Sprintf("%s/%d", basePath, removeIndex),
		})
	}
	if len(current) == 0 {
		operations = append(operations, jsonPatchOperation{
			Op:    "add",
			Path:  basePath,
			Value: missing,
		})
	} else {
		for _, source := range missing {
			operations = append(operations, jsonPatchOperation{
				Op:    "add",
				Path:  basePath + "/-",
				Value: source,
			})
		}
	}

	payload, err := json.Marshal(operations)
	if err != nil {
		return false, err
	}
	_, err = runSSOCommand("kubectl", []string{
		"-n", opts.Namespace,
		"patch", "statefulset", opts.WorkloadName,
		"--type=json", "-p", string(payload),
	}, nil)
	return err == nil, err
}

func managedSSOEnvFromKey(source ssoEnvFromSource, opts ssoOptions) (string, bool) {
	if source.Prefix != "" {
		return "", false
	}
	if source.ConfigMapRef != nil && source.SecretRef == nil &&
		source.ConfigMapRef.Name == opts.ConfigMapName && source.ConfigMapRef.Optional == nil {
		return "configmap:" + opts.ConfigMapName, true
	}
	if source.SecretRef != nil && source.ConfigMapRef == nil &&
		source.SecretRef.Name == opts.SecretName && source.SecretRef.Optional == nil {
		return "secret:" + opts.SecretName, true
	}
	return "", false
}

func desiredSSOEnvFromKey(key string, desired []ssoEnvFromSource) bool {
	for _, source := range desired {
		if source.ConfigMapRef != nil && key == "configmap:"+source.ConfigMapRef.Name {
			return true
		}
		if source.SecretRef != nil && key == "secret:"+source.SecretRef.Name {
			return true
		}
	}
	return false
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

func waitForSSOWorkloadRollout(opts ssoOptions) error {
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
	if len(output) > 0 && !isSSOWorkloadJSONGet(name, args) {
		fmt.Print(string(output))
	}
	return output, nil
}

func isSSOWorkloadJSONGet(name string, args []string) bool {
	if name != "kubectl" {
		return false
	}
	for index := 0; index+1 < len(args); index++ {
		if args[index] == "get" && index+2 < len(args) && args[index+1] == "statefulset" {
			for outputIndex := index + 2; outputIndex+1 < len(args); outputIndex++ {
				if args[outputIndex] == "-o" && args[outputIndex+1] == "json" {
					return true
				}
			}
		}
	}
	return false
}
