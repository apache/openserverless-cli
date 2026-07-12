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
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordedCommand struct {
	name  string
	args  []string
	stdin string
}

func ssoWorkloadJSON(envFrom ...ssoEnvFromSource) []byte {
	workload := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":    defaultSSOContainer,
							"envFrom": envFrom,
						},
					},
				},
			},
		},
	}
	payload, _ := json.Marshal(workload)
	return payload
}

func isSSOWorkloadGet(args []string) bool {
	return len(args) >= 7 && args[2] == "get" && args[3] == "statefulset" && args[6] == "json"
}

func managedConfigMapSource() ssoEnvFromSource {
	return ssoEnvFromSource{ConfigMapRef: &ssoLocalObjectReference{Name: defaultSSOConfigMap}}
}

func managedSecretSource() ssoEnvFromSource {
	return ssoEnvFromSource{SecretRef: &ssoLocalObjectReference{Name: defaultSSOSecret}}
}

type fakeSSOWorkloadState struct {
	Env          []map[string]string
	EnvFrom      []ssoEnvFromSource
	VolumeMounts []map[string]string
	Volumes      []map[string]interface{}
}

func (state *fakeSSOWorkloadState) workloadJSON() []byte {
	workload := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":         defaultSSOContainer,
							"env":          state.Env,
							"envFrom":      state.EnvFrom,
							"volumeMounts": state.VolumeMounts,
						},
					},
					"volumes": state.Volumes,
				},
			},
		},
	}
	payload, _ := json.Marshal(workload)
	return payload
}

func (state *fakeSSOWorkloadState) applyJSONPatch(payload string) error {
	var operations []struct {
		Op    string          `json:"op"`
		Path  string          `json:"path"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal([]byte(payload), &operations); err != nil {
		return err
	}
	for _, operation := range operations {
		switch operation.Op {
		case "remove":
			parts := strings.Split(operation.Path, "/")
			index, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				return err
			}
			state.EnvFrom = append(state.EnvFrom[:index], state.EnvFrom[index+1:]...)
		case "add":
			if strings.HasSuffix(operation.Path, "/-") {
				var source ssoEnvFromSource
				if err := json.Unmarshal(operation.Value, &source); err != nil {
					return err
				}
				state.EnvFrom = append(state.EnvFrom, source)
				continue
			}
			var sources []ssoEnvFromSource
			if err := json.Unmarshal(operation.Value, &sources); err != nil {
				return err
			}
			state.EnvFrom = sources
		}
	}
	return nil
}

func TestConfigSSOToolKeycloak(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	var commands []recordedCommand
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
		if isSSOWorkloadGet(args) {
			return ssoWorkloadJSON(), nil
		}
		return []byte("ok"), nil
	}
	defer func() { runSSOCommand = oldRunner }()

	err = ConfigSSOTool(cm, []string{
		"keycloak",
		"--enable",
		"--issuer-url", "http://localhost:8080/realms/openserverless-lab",
		"--jwks-url", "http://172.18.0.1:8080/realms/openserverless-lab/protocol/openid-connect/certs",
		"--audience", "openserverless-admin-api",
		"--required-group", "openserverless-users",
		"--no-rollout",
	})
	require.NoError(t, err)

	gotConfig, err := readConfig(configPath, fromConfigJson)
	require.NoError(t, err)
	require.Equal(t, true, gotConfig["sso"].(map[string]interface{})["enabled"])
	require.Equal(t, "keycloak", gotConfig["sso"].(map[string]interface{})["provider"])
	require.Equal(t, true, gotConfig["sso"].(map[string]interface{})["autoprovision"].(map[string]interface{})["on"].(map[string]interface{})["login"])
	require.Equal(t, float64(120), gotConfig["sso"].(map[string]interface{})["autoprovision"].(map[string]interface{})["timeout"].(map[string]interface{})["seconds"])

	require.Len(t, commands, 3)
	require.Equal(t, []string{"apply", "-f", "-"}, commands[0].args)
	require.Equal(t, []string{"-n", "nuvolaris", "get", "statefulset", "nuvolaris-system-api", "-o", "json"}, commands[1].args)
	require.Equal(t, []string{"-n", "nuvolaris", "patch", "statefulset", "nuvolaris-system-api", "--type=json", "-p", commands[2].args[7]}, commands[2].args)

	var cmObj map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(commands[0].stdin), &cmObj))
	require.Equal(t, "ConfigMap", cmObj["kind"])
	data := cmObj["data"].(map[string]interface{})
	require.Equal(t, "openserverless-admin-api", data["OIDC_AUDIENCE"])
	require.Equal(t, "openserverless-admin-api", data["OIDC_CLIENT_ID"])
	require.Equal(t, "preferred_username", data["OIDC_USERNAME_CLAIM"])
	require.Equal(t, "groups", data["OIDC_GROUPS_CLAIM"])
	require.Equal(t, "true", data["SSO_AUTOPROVISION_ON_LOGIN"])
	require.Equal(t, "120", data["SSO_AUTOPROVISION_TIMEOUT_SECONDS"])
	require.Equal(t, "2", data["SSO_AUTOPROVISION_POLL_SECONDS"])
	require.Equal(t, "all", data["SSO_AUTOPROVISION_DEFAULT_SERVICES"])
	require.Equal(t, "true", data["SSO_NAMESPACE_PRESERVE_VALID"])
	require.Equal(t, "8", data["SSO_NAMESPACE_HASH_LENGTH"])
	require.Equal(t, "61", data["SSO_NAMESPACE_MAX_LENGTH"])
}

func TestConfigSSOToolKeycloakRollsOutAdminAPIByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	var commands []recordedCommand
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
		if isSSOWorkloadGet(args) {
			return ssoWorkloadJSON(), nil
		}
		return []byte("ok"), nil
	}
	defer func() { runSSOCommand = oldRunner }()

	err = ConfigSSOTool(cm, []string{
		"keycloak",
		"--enable",
		"--issuer-url", "http://localhost:8080/realms/openserverless-lab",
		"--jwks-url", "http://localhost:8080/realms/openserverless-lab/protocol/openid-connect/certs",
		"--client-id", "openserverless-admin-api",
		"--required-group", "openserverless-users",
	})
	require.NoError(t, err)

	require.Len(t, commands, 5)
	require.Equal(t, []string{"apply", "-f", "-"}, commands[0].args)
	require.Equal(t, []string{"-n", "nuvolaris", "get", "statefulset", "nuvolaris-system-api", "-o", "json"}, commands[1].args)
	require.Equal(t, []string{"-n", "nuvolaris", "patch", "statefulset", "nuvolaris-system-api", "--type=json", "-p", commands[2].args[7]}, commands[2].args)
	require.Equal(t, []string{"-n", "nuvolaris", "rollout", "restart", "statefulset/nuvolaris-system-api"}, commands[3].args)
	require.Equal(t, []string{"-n", "nuvolaris", "rollout", "status", "statefulset/nuvolaris-system-api", "--timeout=180s"}, commands[4].args)
}

func TestConfigSSOToolKeycloakWithClientSecret(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	var commands []recordedCommand
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
		if isSSOWorkloadGet(args) {
			return ssoWorkloadJSON(), nil
		}
		return []byte("ok"), nil
	}
	defer func() { runSSOCommand = oldRunner }()

	err = ConfigSSOTool(cm, []string{
		"keycloak",
		"--enable",
		"--issuer-url", "http://localhost:8080/realms/openserverless-lab",
		"--jwks-url", "http://172.18.0.1:8080/realms/openserverless-lab/protocol/openid-connect/certs",
		"--client-id", "openserverless-admin-api",
		"--client-secret", "super-secret",
		"--secret", "custom-sso-secret",
		"--required-group", "openserverless-users",
		"--no-rollout",
	})
	require.NoError(t, err)

	configBytes, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.NotContains(t, string(configBytes), "super-secret")

	flat := cm.Flatten()
	require.Equal(t, "openserverless-admin-api", flat["SSO_OIDC_AUDIENCE"])
	require.Equal(t, "openserverless-admin-api", flat["SSO_OIDC_CLIENT_ID"])
	require.Equal(t, "confidential", flat["SSO_CLIENT_MODE"])
	require.Equal(t, "true", flat["SSO_OIDC_CLIENT_SECRET_CONFIGURED"])
	require.Equal(t, "custom-sso-secret", flat["SSO_KUBE_SECRET"])

	require.Len(t, commands, 4)

	var cmObj map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(commands[0].stdin), &cmObj))
	require.Equal(t, "ConfigMap", cmObj["kind"])
	cmData := cmObj["data"].(map[string]interface{})
	require.Equal(t, "openserverless-admin-api", cmData["OIDC_AUDIENCE"])
	require.Equal(t, "openserverless-admin-api", cmData["OIDC_CLIENT_ID"])
	require.NotContains(t, commands[0].stdin, "super-secret")

	var secretObj map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(commands[1].stdin), &secretObj))
	require.Equal(t, "Secret", secretObj["kind"])
	require.Equal(t, "custom-sso-secret", secretObj["metadata"].(map[string]interface{})["name"])
	secretData := secretObj["stringData"].(map[string]interface{})
	require.Equal(t, "super-secret", secretData["OIDC_CLIENT_SECRET"])

	require.Equal(t, []string{"-n", "nuvolaris", "get", "statefulset", "nuvolaris-system-api", "-o", "json"}, commands[2].args)
	require.Equal(t, []string{"-n", "nuvolaris", "patch", "statefulset", "nuvolaris-system-api", "--type=json", "-p", commands[3].args[7]}, commands[3].args)
	require.Contains(t, commands[3].args[7], "custom-sso-secret")
}

func TestConfigSSOToolKeycloakRequiresValues(t *testing.T) {
	cm, err := NewConfigMapBuilder().Build()
	require.NoError(t, err)

	err = ConfigSSOTool(cm, []string{"keycloak", "--enable"})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "missing --issuer-url"))
}

func TestConfigSSOToolDisable(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{
  "sso": {
    "enabled": "true",
    "provider": "keycloak",
    "oidc": {
      "issuer": {
        "url": "http://issuer"
      }
    },
    "external": "preserve-me"
  }
}`), 0644))

	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	var commands []recordedCommand
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
		if isSSOWorkloadGet(args) {
			return ssoWorkloadJSON(managedConfigMapSource(), managedSecretSource()), nil
		}
		return []byte("ok"), nil
	}
	defer func() { runSSOCommand = oldRunner }()

	err = ConfigSSOTool(cm, []string{"disable", "--no-rollout"})
	require.NoError(t, err)

	gotConfig, err := readConfig(configPath, fromConfigJson)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{
		"sso": map[string]interface{}{
			"external": "preserve-me",
		},
	}, gotConfig)
	require.Len(t, commands, 4)
	require.Equal(t, []string{"-n", "nuvolaris", "get", "statefulset", "nuvolaris-system-api", "-o", "json"}, commands[0].args)
	require.Equal(t, []string{"-n", "nuvolaris", "patch", "statefulset", "nuvolaris-system-api", "--type=json", "-p", commands[1].args[7]}, commands[1].args)
	require.Equal(t, []string{"-n", "nuvolaris", "delete", "configmap", "openserverless-sso-config", "--ignore-not-found"}, commands[2].args)
	require.Equal(t, []string{"-n", "nuvolaris", "delete", "secret", "openserverless-sso-secret", "--ignore-not-found"}, commands[3].args)
}

func TestConfigSSOToolDisableWaitsForAdminAPIRolloutByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	var commands []recordedCommand
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
		if isSSOWorkloadGet(args) {
			return ssoWorkloadJSON(managedConfigMapSource(), managedSecretSource()), nil
		}
		return []byte("ok"), nil
	}
	defer func() { runSSOCommand = oldRunner }()

	err = ConfigSSOTool(cm, []string{"disable"})
	require.NoError(t, err)

	require.Len(t, commands, 5)
	require.Equal(t, []string{"-n", "nuvolaris", "get", "statefulset", "nuvolaris-system-api", "-o", "json"}, commands[0].args)
	require.Equal(t, []string{"-n", "nuvolaris", "patch", "statefulset", "nuvolaris-system-api", "--type=json", "-p", commands[1].args[7]}, commands[1].args)
	require.Equal(t, []string{"-n", "nuvolaris", "delete", "configmap", "openserverless-sso-config", "--ignore-not-found"}, commands[2].args)
	require.Equal(t, []string{"-n", "nuvolaris", "delete", "secret", "openserverless-sso-secret", "--ignore-not-found"}, commands[3].args)
	require.Equal(t, []string{"-n", "nuvolaris", "rollout", "status", "statefulset/nuvolaris-system-api", "--timeout=180s"}, commands[4].args)
}

func TestConfigSSOToolDisableAlreadyAbsentDoesNotPatchOrRollOut(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	foreign := ssoEnvFromSource{
		ConfigMapRef: &ssoLocalObjectReference{Name: "application-config"},
	}
	var commands []recordedCommand
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
		if isSSOWorkloadGet(args) {
			return ssoWorkloadJSON(foreign), nil
		}
		return []byte("ok"), nil
	}
	defer func() { runSSOCommand = oldRunner }()

	require.NoError(t, ConfigSSOTool(cm, []string{"disable"}))
	require.Len(t, commands, 3)
	require.Equal(t, []string{"-n", "nuvolaris", "get", "statefulset", "nuvolaris-system-api", "-o", "json"}, commands[0].args)
	require.Equal(t, []string{"-n", "nuvolaris", "delete", "configmap", "openserverless-sso-config", "--ignore-not-found"}, commands[1].args)
	require.Equal(t, []string{"-n", "nuvolaris", "delete", "secret", "openserverless-sso-secret", "--ignore-not-found"}, commands[2].args)
}

func TestConfigSSOToolPreservesForeignWorkloadFieldsAcrossEnableDisableEnable(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	state := fakeSSOWorkloadState{
		Env: []map[string]string{
			{"name": "APPLICATION_MODE", "value": "production"},
		},
		EnvFrom: []ssoEnvFromSource{
			{ConfigMapRef: &ssoLocalObjectReference{Name: "application-config"}},
			{SecretRef: &ssoLocalObjectReference{Name: "database-credentials"}},
		},
		VolumeMounts: []map[string]string{
			{"name": "application-data", "mountPath": "/data"},
		},
		Volumes: []map[string]interface{}{
			{"name": "application-data", "emptyDir": map[string]interface{}{}},
		},
	}
	originalEnv, err := json.Marshal(state.Env)
	require.NoError(t, err)
	originalMounts, err := json.Marshal(state.VolumeMounts)
	require.NoError(t, err)
	originalVolumes, err := json.Marshal(state.Volumes)
	require.NoError(t, err)

	var commands []recordedCommand
	var patchPayloads []string
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
		if isSSOWorkloadGet(args) {
			return state.workloadJSON(), nil
		}
		if len(args) >= 8 && args[2] == "patch" && args[5] == "--type=json" {
			patchPayloads = append(patchPayloads, args[7])
			if err := state.applyJSONPatch(args[7]); err != nil {
				return nil, err
			}
		}
		return []byte("ok"), nil
	}
	defer func() { runSSOCommand = oldRunner }()

	enableArgs := []string{
		"keycloak",
		"--enable",
		"--issuer-url", "https://keycloak.example.test/realms/openserverless",
		"--jwks-url", "https://keycloak.example.test/realms/openserverless/protocol/openid-connect/certs",
		"--client-id", "openserverless-admin-api",
		"--client-secret", "test-secret",
		"--required-group", "openserverless-users",
		"--no-rollout",
	}
	require.NoError(t, ConfigSSOTool(cm, enableArgs))
	require.Equal(t, []ssoEnvFromSource{
		{ConfigMapRef: &ssoLocalObjectReference{Name: "application-config"}},
		{SecretRef: &ssoLocalObjectReference{Name: "database-credentials"}},
		managedConfigMapSource(),
		managedSecretSource(),
	}, state.EnvFrom)

	require.NoError(t, ConfigSSOTool(cm, []string{"disable", "--no-rollout"}))
	require.Equal(t, []ssoEnvFromSource{
		{ConfigMapRef: &ssoLocalObjectReference{Name: "application-config"}},
		{SecretRef: &ssoLocalObjectReference{Name: "database-credentials"}},
	}, state.EnvFrom)
	require.Len(t, patchPayloads, 2)
	require.JSONEq(t, `[
  {"op":"test","path":"/spec/template/spec/containers/0/envFrom/3","value":{"secretRef":{"name":"openserverless-sso-secret"}}},
  {"op":"remove","path":"/spec/template/spec/containers/0/envFrom/3"},
  {"op":"test","path":"/spec/template/spec/containers/0/envFrom/2","value":{"configMapRef":{"name":"openserverless-sso-config"}}},
  {"op":"remove","path":"/spec/template/spec/containers/0/envFrom/2"}
]`, patchPayloads[1])

	// Repeating disable is idempotent: resources are deleted with
	// --ignore-not-found and no StatefulSet patch or rollout is produced.
	require.NoError(t, ConfigSSOTool(cm, []string{"disable", "--no-rollout"}))
	require.Len(t, patchPayloads, 2)

	require.NoError(t, ConfigSSOTool(cm, enableArgs))
	require.Equal(t, []ssoEnvFromSource{
		{ConfigMapRef: &ssoLocalObjectReference{Name: "application-config"}},
		{SecretRef: &ssoLocalObjectReference{Name: "database-credentials"}},
		managedConfigMapSource(),
		managedSecretSource(),
	}, state.EnvFrom)
	require.Len(t, patchPayloads, 3)

	currentEnv, err := json.Marshal(state.Env)
	require.NoError(t, err)
	currentMounts, err := json.Marshal(state.VolumeMounts)
	require.NoError(t, err)
	currentVolumes, err := json.Marshal(state.Volumes)
	require.NoError(t, err)
	require.JSONEq(t, string(originalEnv), string(currentEnv))
	require.JSONEq(t, string(originalMounts), string(currentMounts))
	require.JSONEq(t, string(originalVolumes), string(currentVolumes))
	for _, command := range commands {
		require.NotContains(t, command.args, "rollout")
	}
}
