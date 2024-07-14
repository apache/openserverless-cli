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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func ExampleConfigTool_readValue() {
	tmpDir, _ := os.MkdirTemp("", "nuv")
	defer os.RemoveAll(tmpDir)

	nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
	configPath := filepath.Join(tmpDir, "config.json")
	cm, _ := NewConfigMapBuilder().WithNuvRoot(nuvRootPath).WithConfigJson(configPath).Build()

	os.Args = []string{"config", "FOO=bar"}
	err := ConfigTool(cm)
	if err != nil {
		fmt.Println("error:", err)
	}

	os.Args = []string{"config", "FOO"}
	err = ConfigTool(cm)
	if err != nil {
		fmt.Println("error:", err)
	}

	// nested key
	os.Args = []string{"config", "NESTED_VAL=val"}
	err = ConfigTool(cm)
	if err != nil {
		fmt.Println("error:", err)
	}

	os.Args = []string{"config", "NESTED_VAL"}
	err = ConfigTool(cm)
	if err != nil {
		fmt.Println("error:", err)
	}
	// Output:
	// bar
	// val
}

func TestConfigTool(t *testing.T) {
	readConfigJson := func(path string) (map[string]interface{}, error) {
		return readConfig(filepath.Join(path, "config.json"), fromConfigJson)
	}

	buildCM := func(nv string, c string) ConfigMap {
		cm, _ := NewConfigMapBuilder().WithConfigJson(c).WithNuvRoot(nv).Build()
		return cm
	}
	t.Run("new config.json", func(t *testing.T) {
		tmpDir := t.TempDir()

		nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
		configPath := filepath.Join(tmpDir, "config.json")
		cm := buildCM(nuvRootPath, configPath)

		os.Args = []string{"config", "FOO=bar"}
		err := ConfigTool(cm)
		require.NoError(t, err)

		got, err := readConfigJson(tmpDir)
		require.NoError(t, err)

		want := map[string]interface{}{
			"foo": "bar",
		}

		require.Equal(t, want, got)
	})

	t.Run("write values on existing config.json", func(t *testing.T) {
		tmpDir := t.TempDir()
		nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
		configPath := filepath.Join(tmpDir, "config.json")
		cm := buildCM(nuvRootPath, configPath)

		os.Args = []string{"config", "FOO=bar"}
		err := ConfigTool(cm)
		require.NoError(t, err)

		os.Args = []string{"config", "BAR=baz"}
		err = ConfigTool(cm)
		require.NoError(t, err)

		got, err := readConfigJson(tmpDir)
		require.NoError(t, err)

		want := map[string]interface{}{
			"foo": "bar",
			"bar": "baz",
		}

		require.Equal(t, want, got)
	})

	t.Run("write existing value is overridden", func(t *testing.T) {
		tmpDir := t.TempDir()
		nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
		configPath := filepath.Join(tmpDir, "config.json")
		cm := buildCM(nuvRootPath, configPath)

		os.Args = []string{"config", "FOO=bar"}
		err := ConfigTool(cm)
		require.NoError(t, err)

		os.Args = []string{"config", "FOO=new"}
		err = ConfigTool(cm)
		require.NoError(t, err)

		got, err := readConfigJson(tmpDir)
		require.NoError(t, err)

		want := map[string]interface{}{
			"foo": "new",
		}

		require.Equal(t, want, got)
	})

	t.Run("write existing key object is merged", func(t *testing.T) {
		tmpDir := t.TempDir()
		nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
		configPath := filepath.Join(tmpDir, "config.json")
		cm := buildCM(nuvRootPath, configPath)

		os.Args = []string{"config", "FOO_BAR=bar"}
		err := ConfigTool(cm)
		require.NoError(t, err)

		os.Args = []string{"config", "FOO_BAZ=baz"}
		err = ConfigTool(cm)
		require.NoError(t, err)

		got, err := readConfigJson(tmpDir)
		require.NoError(t, err)

		want := map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "bar",
				"baz": "baz",
			},
		}

		require.Equal(t, want, got)
	})

	t.Run("write empty string to disable key", func(t *testing.T) {
		tmpDir := t.TempDir()
		nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
		configPath := filepath.Join(tmpDir, "config.json")
		cm := buildCM(nuvRootPath, configPath)

		os.Args = []string{"config", "FOO_BAR=bar"}
		err := ConfigTool(cm)
		require.NoError(t, err)

		os.Args = []string{"config", "FOO_BAR=\"\""}
		err = ConfigTool(cm)
		require.NoError(t, err)

		got, err := readConfigJson(tmpDir)
		require.NoError(t, err)

		want := map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "",
			},
		}

		require.Equal(t, want, got)
	})

	t.Run("remove existing key", func(t *testing.T) {
		tmpDir := t.TempDir()
		nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
		configPath := filepath.Join(tmpDir, "config.json")
		cm := buildCM(nuvRootPath, configPath)

		os.Args = []string{"config", "FOO=bar"}
		err := ConfigTool(cm)
		require.NoError(t, err)

		os.Args = []string{"config", "-r", "FOO"}
		err = ConfigTool(cm)
		require.NoError(t, err)

		got, err := readConfigJson(tmpDir)
		require.NoError(t, err)

		want := map[string]interface{}{}

		require.Equal(t, want, got)
	})

	t.Run("remove nested key object", func(t *testing.T) {
		tmpDir := t.TempDir()
		nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
		configPath := filepath.Join(tmpDir, "config.json")
		cm := buildCM(nuvRootPath, configPath)

		os.Args = []string{"config", "FOO_BAR=bar", "FOO_BAZ=baz"}
		err := ConfigTool(cm)
		require.NoError(t, err)

		got, err := readConfigJson(tmpDir)
		require.NoError(t, err)
		want := map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "bar",
				"baz": "baz",
			},
		}
		require.Equal(t, want, got)

		os.Args = []string{"config", "-r", "FOO_BAR"}
		err = ConfigTool(cm)
		require.NoError(t, err)

		got, err = readConfigJson(tmpDir)
		require.NoError(t, err)

		want = map[string]interface{}{
			"foo": map[string]interface{}{
				"baz": "baz",
			},
		}

		require.Equal(t, want, got)
	})

	t.Run("remove non-existing key", func(t *testing.T) {
		tmpDir := t.TempDir()
		nuvRootPath := filepath.Join(tmpDir, "nuvroot.json")
		configPath := filepath.Join(tmpDir, "config.json")
		cm := buildCM(nuvRootPath, configPath)

		os.Args = []string{"config", "-r", "FOO"}
		err := ConfigTool(cm)
		require.Error(t, err)

		got, err := readConfigJson(tmpDir)
		require.NoError(t, err)

		want := map[string]interface{}{}

		require.Equal(t, want, got)
	})
}

func Test_buildKeyValueMap(t *testing.T) {
	testCases := []struct {
		name  string
		input []string
		want  keyValues
		err   error
	}{
		{
			name:  "Empty string",
			input: []string{},
			want:  nil,
			err:   fmt.Errorf("no key-value pairs provided"),
		},
		{
			name:  "Single key-value pair",
			input: []string{"foo=bar"},
			want:  keyValues{"foo": "bar"},
			err:   nil,
		},
		{
			name:  "Multiple key-value pairs",
			input: []string{"foo=bar", "baz=qux"},
			want:  keyValues{"foo": "bar", "baz": "qux"},
			err:   nil,
		},
		{
			name:  "Invalid key-value pair",
			input: []string{"foo"},
			want:  nil,
			err:   fmt.Errorf("invalid key-value pair: %q", "foo"),
		},
		{
			name:  "Invalid key-value pair",
			input: []string{"foo="},
			want:  nil,
			err:   fmt.Errorf("invalid key-value pair: %q", "foo="),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildInputKVMap(tc.input)

			require.Equal(t, tc.err, err)
			require.Equal(t, tc.want, got)
		})
	}
}
