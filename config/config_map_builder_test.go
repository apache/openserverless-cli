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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigMapBuilder(t *testing.T) {
	t.Run("build without plugins", func(t *testing.T) {
		tmpDir := t.TempDir()

		configJsonPath := createFakeConfigFile(t, "config.json", tmpDir, `
	{
		"key": "value",
		"nested": {
			"key": 123
		}
	}`)
		opsRootPath := createFakeConfigFile(t, "opsroot.json", tmpDir, `
	{
		"version": "0.3.0",
		"config": {
			"opsroot": "value",
			"another": {
				"key": 123
			}
		}
	}`)

		testCases := []struct {
			name       string
			configJson string
			opsRoot    string
			want       ConfigMap
			err        error
		}{
			{
				name:       "should return empty configmap when no files are added",
				configJson: "",
				opsRoot:    "",
				want: ConfigMap{
					config:               map[string]interface{}{},
					opsRootConfig:        map[string]interface{}{},
					pluginOpsRootConfigs: map[string]map[string]interface{}{},
				},
				err: nil,
			},
			{
				name:       "should return with config when a valid config.json is added",
				configJson: configJsonPath,
				opsRoot:    "",
				want: ConfigMap{
					opsRootConfig: map[string]interface{}{},
					config: map[string]interface{}{
						"key": "value",
						"nested": map[string]interface{}{
							"key": 123.0,
						},
					},
					configPath:           configJsonPath,
					pluginOpsRootConfigs: map[string]map[string]interface{}{},
				},
			},
			{
				name:       "should return with opsroot when a valid opsroot.json is added",
				configJson: "",
				opsRoot:    opsRootPath,
				want: ConfigMap{
					opsRootConfig: map[string]interface{}{
						"opsroot": "value",
						"another": map[string]interface{}{
							"key": 123.0,
						},
					},
					config:               map[string]interface{}{},
					pluginOpsRootConfigs: map[string]map[string]interface{}{},
				},
			},
			{
				name:       "should return with both when both config.json and opsroot.json are added",
				configJson: configJsonPath,
				opsRoot:    opsRootPath,
				want: ConfigMap{
					config: map[string]interface{}{
						"key": "value",
						"nested": map[string]interface{}{
							"key": 123.0,
						},
					},
					opsRootConfig: map[string]interface{}{
						"opsroot": "value",
						"another": map[string]interface{}{
							"key": 123.0,
						},
					},
					configPath:           configJsonPath,
					pluginOpsRootConfigs: map[string]map[string]interface{}{},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {

				got, err := NewConfigMapBuilder().
					WithConfigJson(tc.configJson).
					WithOpsRoot(tc.opsRoot).
					Build()

				// if we expect an error but got none
				if tc.err != nil && err == nil {
					t.Errorf("want error %e, got %e", tc.err, err)
				}

				// if we expect no error but got one
				if tc.err == nil && err != nil {
					t.Errorf("want no error, but got %e", err)
				}

				require.Equal(t, tc.want, got)
			})
		}
	})

	t.Run("build with plugins", func(t *testing.T) {
		tmpDir := t.TempDir()

		configJsonPath := createFakeConfigFile(t, "config.json", tmpDir, `
	{
		"key": "value",
		"nested": {
			"key": 123
		}
	}`)
		opsRootPath := createFakeConfigFile(t, "opsroot.json", tmpDir, `
	{
		"version": "0.3.0",
		"config": {
			"opsroot": "value",
			"another": {
				"key": 123
			}
		}
	}`)

		pluginOpsRoot := createFakeConfigFile(t, "opsroot.json", tmpDir, `
	{
		"version": "0.3.0",
		"config": {
			"opsroot": "value",
			"another": {
				"key": 123
			}
		}
	}`)

		testCases := []struct {
			name           string
			configJson     string
			opsRoot        string
			pluginOpsRoots map[string]string
			want           ConfigMap
			err            error
		}{
			{
				name:       "should return configmap containing plugin opsroot",
				configJson: configJsonPath,
				opsRoot:    opsRootPath,
				pluginOpsRoots: map[string]string{
					"plugin": pluginOpsRoot,
				},

				want: ConfigMap{
					config: map[string]interface{}{
						"key": "value",
						"nested": map[string]interface{}{
							"key": 123.0,
						},
					},
					opsRootConfig: map[string]interface{}{
						"opsroot": "value",
						"another": map[string]interface{}{
							"key": 123.0,
						},
					},
					configPath: configJsonPath,
					pluginOpsRootConfigs: map[string]map[string]interface{}{
						"plugin": {
							"opsroot": "value",
							"another": map[string]interface{}{
								"key": 123.0,
							},
						},
					},
				},

				err: nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {

				got, err := NewConfigMapBuilder().
					WithConfigJson(tc.configJson).
					WithOpsRoot(tc.opsRoot).
					WithPluginOpsRoots(tc.pluginOpsRoots).
					Build()

				// if we expect an error but got none
				if tc.err != nil && err == nil {
					t.Errorf("want error %e, got %e", tc.err, err)
				}

				// if we expect no error but got one
				if tc.err == nil && err != nil {
					t.Errorf("want no error, but got %e", err)
				}

				require.Equal(t, tc.want, got)
			})
		}
	})
}

func createFakeConfigFile(t *testing.T, name, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}
