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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordedCommand struct {
	name  string
	args  []string
	stdin string
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

	require.Len(t, commands, 2)
	require.Equal(t, []string{"apply", "-f", "-"}, commands[0].args)
	require.Equal(t, []string{"-n", "nuvolaris", "patch", "statefulset", "nuvolaris-system-api", "--type=strategic", "-p", commands[1].args[7]}, commands[1].args)

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

func TestConfigSSOToolKeycloakWithClientSecret(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	var commands []recordedCommand
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
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

	require.Len(t, commands, 3)

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

	require.Equal(t, []string{"-n", "nuvolaris", "patch", "statefulset", "nuvolaris-system-api", "--type=strategic", "-p", commands[2].args[7]}, commands[2].args)
	require.Contains(t, commands[2].args[7], "custom-sso-secret")
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
    }
  }
}`), 0644))

	cm, err := NewConfigMapBuilder().WithConfigJson(configPath).Build()
	require.NoError(t, err)

	var commands []recordedCommand
	oldRunner := runSSOCommand
	runSSOCommand = func(name string, args []string, stdin []byte) ([]byte, error) {
		commands = append(commands, recordedCommand{name: name, args: args, stdin: string(stdin)})
		return []byte("ok"), nil
	}
	defer func() { runSSOCommand = oldRunner }()

	err = ConfigSSOTool(cm, []string{"disable", "--no-rollout"})
	require.NoError(t, err)

	gotConfig, err := readConfig(configPath, fromConfigJson)
	require.NoError(t, err)
	require.Empty(t, gotConfig)
	require.Len(t, commands, 3)
	require.Equal(t, []string{"-n", "nuvolaris", "delete", "configmap", "openserverless-sso-config", "--ignore-not-found"}, commands[1].args)
	require.Equal(t, []string{"-n", "nuvolaris", "delete", "secret", "openserverless-sso-secret", "--ignore-not-found"}, commands[2].args)
}
