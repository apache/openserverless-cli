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
		nuvRootPath := createFakeConfigFile(t, "nuvroot.json", tmpDir, `
	{
		"version": "0.3.0",
		"config": {
			"nuvroot": "value",
			"another": {
				"key": 123
			}
		}
	}`)

		testCases := []struct {
			name       string
			configJson string
			nuvRoot    string
			want       ConfigMap
			err        error
		}{
			{
				name:       "should return empty configmap when no files are added",
				configJson: "",
				nuvRoot:    "",
				want: ConfigMap{
					config:               map[string]interface{}{},
					nuvRootConfig:        map[string]interface{}{},
					pluginNuvRootConfigs: map[string]map[string]interface{}{},
				},
				err: nil,
			},
			{
				name:       "should return with config when a valid config.json is added",
				configJson: configJsonPath,
				nuvRoot:    "",
				want: ConfigMap{
					nuvRootConfig: map[string]interface{}{},
					config: map[string]interface{}{
						"key": "value",
						"nested": map[string]interface{}{
							"key": 123.0,
						},
					},
					configPath:           configJsonPath,
					pluginNuvRootConfigs: map[string]map[string]interface{}{},
				},
			},
			{
				name:       "should return with nuvroot when a valid nuvroot.json is added",
				configJson: "",
				nuvRoot:    nuvRootPath,
				want: ConfigMap{
					nuvRootConfig: map[string]interface{}{
						"nuvroot": "value",
						"another": map[string]interface{}{
							"key": 123.0,
						},
					},
					config:               map[string]interface{}{},
					pluginNuvRootConfigs: map[string]map[string]interface{}{},
				},
			},
			{
				name:       "should return with both when both config.json and nuvroot.json are added",
				configJson: configJsonPath,
				nuvRoot:    nuvRootPath,
				want: ConfigMap{
					config: map[string]interface{}{
						"key": "value",
						"nested": map[string]interface{}{
							"key": 123.0,
						},
					},
					nuvRootConfig: map[string]interface{}{
						"nuvroot": "value",
						"another": map[string]interface{}{
							"key": 123.0,
						},
					},
					configPath:           configJsonPath,
					pluginNuvRootConfigs: map[string]map[string]interface{}{},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {

				got, err := NewConfigMapBuilder().
					WithConfigJson(tc.configJson).
					WithNuvRoot(tc.nuvRoot).
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
		nuvRootPath := createFakeConfigFile(t, "nuvroot.json", tmpDir, `
	{
		"version": "0.3.0",
		"config": {
			"nuvroot": "value",
			"another": {
				"key": 123
			}
		}
	}`)

		pluginNuvRoot := createFakeConfigFile(t, "nuvroot.json", tmpDir, `
	{
		"version": "0.3.0",
		"config": {
			"nuvroot": "value",
			"another": {
				"key": 123
			}
		}
	}`)

		testCases := []struct {
			name           string
			configJson     string
			nuvRoot        string
			pluginNuvRoots map[string]string
			want           ConfigMap
			err            error
		}{
			{
				name:       "should return configmap containing plugin nuvroot",
				configJson: configJsonPath,
				nuvRoot:    nuvRootPath,
				pluginNuvRoots: map[string]string{
					"plugin": pluginNuvRoot,
				},

				want: ConfigMap{
					config: map[string]interface{}{
						"key": "value",
						"nested": map[string]interface{}{
							"key": 123.0,
						},
					},
					nuvRootConfig: map[string]interface{}{
						"nuvroot": "value",
						"another": map[string]interface{}{
							"key": 123.0,
						},
					},
					configPath: configJsonPath,
					pluginNuvRootConfigs: map[string]map[string]interface{}{
						"plugin": {
							"nuvroot": "value",
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
					WithNuvRoot(tc.nuvRoot).
					WithPluginNuvRoots(tc.pluginNuvRoots).
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
