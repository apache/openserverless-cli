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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	configMap, err := NewConfigMapBuilder().WithConfigJson(filepath.Join(tempDir, "config.json")).Build()
	require.NoError(t, err)

	err = configMap.Insert("key", "value")
	require.NoError(t, err)

	err = configMap.SaveConfig()
	require.NoError(t, err)

	// Read the saved file
	savedJSON, err := os.ReadFile(configMap.configPath)
	require.NoError(t, err)

	// Unmarshal the JSON
	var savedConfig map[string]interface{}
	err = json.Unmarshal(savedJSON, &savedConfig)
	require.NoError(t, err)

	// Verify the saved config matches the original config
	require.Equal(t, configMap.config, savedConfig)
}

func TestInsert(t *testing.T) {

	testCases := []struct {
		name        string
		startingMap ConfigMap
		key         string
		value       string
		expected    ConfigMap
		err         error
	}{
		{
			name: "empty map",
			startingMap: ConfigMap{
				config: map[string]interface{}{},
			},
			key:   "KEY",
			value: "value",
			expected: ConfigMap{
				config: map[string]interface{}{
					"key": "value",
				},
			},
			err: nil,
		},
		{
			name: "map with key",
			startingMap: ConfigMap{
				config: map[string]interface{}{
					"key": "value",
				},
			},
			key:   "KEY2",
			value: "value2",
			expected: ConfigMap{
				config: map[string]interface{}{
					"key":  "value",
					"key2": "value2",
				},
			},
			err: nil,
		},
		{
			name: "map with nested key",
			startingMap: ConfigMap{
				config: map[string]interface{}{
					"key": map[string]interface{}{
						"key": "value",
					},
				},
			},
			key:   "KEY_OTHER",
			value: "value2",
			expected: ConfigMap{
				config: map[string]interface{}{
					"key": map[string]interface{}{
						"key":   "value",
						"other": "value2",
					},
				},
			},
			err: nil,
		},
		{
			name: "existing key is overwritten",
			startingMap: ConfigMap{
				config: map[string]interface{}{
					"key": "value",
				},
			},
			key:   "KEY",
			value: "value2",
			expected: ConfigMap{
				config: map[string]interface{}{
					"key": "value2",
				},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.startingMap.Insert(tc.key, tc.value)
			assertExpectedError(t, err, tc.err)

			if !reflect.DeepEqual(tc.startingMap, tc.expected) {
				t.Errorf("expected '%v' but got '%v'", tc.expected, tc.startingMap)
			}
		})
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name        string
		startingMap ConfigMap
		key         string
		expected    string
		err         error
	}{
		{
			name: "empty maps",
			startingMap: ConfigMap{
				nuvRootConfig: map[string]interface{}{},
				config:        map[string]interface{}{},
			},
			key:      "KEY",
			expected: "",
			err:      fmt.Errorf("invalid key: '%s' - key does not exist", "KEY"),
		},
		{
			name: "config map with key",
			startingMap: ConfigMap{
				nuvRootConfig: map[string]interface{}{},
				config: map[string]interface{}{
					"key": "value",
				},
			},
			key:      "KEY",
			expected: "value",
			err:      nil,
		},
		{
			name: "nuv root config with key",
			startingMap: ConfigMap{
				nuvRootConfig: map[string]interface{}{
					"key": "value",
				},
				config: map[string]interface{}{},
			},
			key:      "KEY",
			expected: "value",
			err:      nil,
		},
		{
			name: "map with nested key",
			startingMap: ConfigMap{
				config: map[string]interface{}{
					"nested": map[string]interface{}{
						"key": "value",
					},
				},
			},
			key:      "NESTED_KEY",
			expected: "value",
			err:      nil,
		},
		{
			name: "map with both nuv root and config keys",
			startingMap: ConfigMap{
				nuvRootConfig: map[string]interface{}{
					"key":  "value",
					"key2": "value2",
				},
				config: map[string]interface{}{
					"key":       "value2",
					"different": "value3",
				},
			},
			key:      "KEY",
			expected: "value2",
			err:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := tc.startingMap.Get(tc.key)
			assertExpectedError(t, err, tc.err)

			if value != tc.expected {
				t.Errorf("expected '%s' but got '%s'", tc.expected, value)
			}
		})
	}
}

func TestFlatten(t *testing.T) {

	testCases := []struct {
		name     string
		input    ConfigMap
		expected map[string]string
	}{
		{
			name: "empty map",
			input: ConfigMap{
				config:        map[string]interface{}{},
				nuvRootConfig: map[string]interface{}{},
			},
			expected: map[string]string{},
		},
		{
			name: "one key map",
			input: ConfigMap{
				nuvRootConfig: map[string]interface{}{},
				config: map[string]interface{}{
					"key": "value",
				},
			},
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name: "nested map",
			input: ConfigMap{
				config: map[string]interface{}{
					"key": map[string]interface{}{
						"key": "value",
					},
				},
			},
			expected: map[string]string{
				"KEY_KEY": "value",
			},
		},
		{
			name: "nested map with multiple keys",
			input: ConfigMap{
				nuvRootConfig: map[string]interface{}{
					"different": "value",
					"key":       "value0",
				},
				config: map[string]interface{}{
					"key": map[string]interface{}{
						"key":   "value1",
						"other": "value2",
					},
				},
			},
			expected: map[string]string{
				"DIFFERENT": "value",
				"KEY_KEY":   "value1",
				"KEY_OTHER": "value2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flattened := tc.input.Flatten()
			if !reflect.DeepEqual(flattened, tc.expected) {
				t.Errorf("expected '%v' but got '%v'", tc.expected, flattened)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	emptyMap := ConfigMap{
		nuvRootConfig: map[string]interface{}{},
		config:        map[string]interface{}{},
	}
	testCases := []struct {
		name        string
		startingMap ConfigMap
		key         string
		expected    ConfigMap
		err         error
	}{
		{
			name:        "empty map",
			startingMap: emptyMap,
			key:         "KEY",
			expected:    emptyMap,
			err:         fmt.Errorf("invalid key: '%s' - key does not exist in config.json", "KEY"),
		},
		{
			name: "map with key",
			startingMap: ConfigMap{
				nuvRootConfig: map[string]interface{}{},
				config:        map[string]interface{}{"key": "value"},
			},
			key:      "KEY",
			expected: emptyMap,
			err:      nil,
		},
		{
			name: "map with nested key",
			startingMap: ConfigMap{
				nuvRootConfig: map[string]interface{}{},
				config: map[string]interface{}{
					"nested": map[string]interface{}{
						"key": "value",
					},
				},
			},
			key: "NESTED_KEY",
			expected: ConfigMap{
				nuvRootConfig: map[string]interface{}{},
				config:        map[string]interface{}{},
			},
			err: nil,
		},
		{
			name: "nuvroot is ignored",
			startingMap: ConfigMap{
				nuvRootConfig: map[string]interface{}{
					"key": "value",
				},
				config: map[string]interface{}{},
			},
			key: "KEY",
			expected: ConfigMap{
				nuvRootConfig: map[string]interface{}{
					"key": "value",
				},
				config: map[string]interface{}{},
			},
			err: fmt.Errorf("invalid key: '%s' - key does not exist in config.json", "KEY"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.startingMap.Delete(tc.key)
			assertExpectedError(t, err, tc.err)

			if !reflect.DeepEqual(tc.startingMap, tc.expected) {
				t.Errorf("expected '%v' but got '%v'", tc.expected, tc.startingMap)
			}
		})
	}

}
func Test_parseKey(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []string
		err   error
	}{
		{
			name:  "Simple Key",
			input: "foo",
			want:  []string{"foo"},
			err:   nil,
		},
		{
			name:  "Complex Key",
			input: "foo_bar",
			want:  []string{"foo", "bar"},
			err:   nil,
		},
		{
			name:  "Complex Key 2",
			input: "foo_bar_baz",
			want:  []string{"foo", "bar", "baz"},
			err:   nil,
		},
		{
			name:  "Invalid Key",
			input: "foo_bar_baz_",
			want:  nil,
			err:   fmt.Errorf("invalid key: %s", "foo_bar_baz_"),
		},
		{
			name:  "Invalid Key 2",
			input: "_foo_bar_baz",
			want:  nil,
			err:   fmt.Errorf("invalid key: %s", "_foo_bar_baz"),
		}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseKey(tc.input)
			assertExpectedError(t, err, tc.err)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("Expected %v, got %v", tc.want, got)
			}
		})
	}
}

func Test_parseValue(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  interface{}
		err   error
	}{
		{
			name:  "String",
			input: "foo",
			want:  "foo",
			err:   nil,
		},
		{
			name:  "Complex String",
			input: "Another foo bar",
			want:  "Another foo bar",
			err:   nil,
		},
		{
			name:  "Number",
			input: "123.456",
			want:  123.456,
			err:   nil,
		},
		{
			name:  "Boolean True",
			input: "true",
			want:  true,
			err:   nil,
		},
		{
			name:  "Boolean False",
			input: "false",
			want:  false,
			err:   nil,
		},
		{
			name:  "Null",
			input: "null",
			want:  nil,
			err:   nil,
		},
		{
			name:  "JSON",
			input: `{"foo": "bar"}`,
			want:  map[string]interface{}{"foo": "bar"},
			err:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseValue(tc.input)
			assertExpectedError(t, err, tc.err)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("Expected %v, got %v", tc.want, got)
			}
		})
	}
}

func assertExpectedError(t *testing.T, err error, expected error) {
	t.Helper()
	if expected != nil {
		if err == nil {
			t.Fatalf("expected error '%s' but got nil", expected)
		}
		if err.Error() != expected.Error() {
			t.Errorf("expected error '%s' but got '%s'", expected, err)
		}
	}

	if expected == nil && err != nil {
		t.Errorf("expected no error but got '%s'", err)
	}
}

func Test_mergeMaps(t *testing.T) {
	testCases := []struct {
		name     string
		m1       map[string]interface{}
		m2       map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "m1 empty",
			m1:   map[string]interface{}{},
			m2: map[string]interface{}{
				"test": "test",
			},
			expected: map[string]interface{}{
				"test": "test",
			},
		},
		{
			name: "m2 empty",
			m1: map[string]interface{}{
				"test": "test",
			},
			m2: map[string]interface{}{},
			expected: map[string]interface{}{
				"test": "test",
			},
		},
		{
			name: "m1 and m2 not empty",
			m1: map[string]interface{}{
				"test": "test",
			},
			m2: map[string]interface{}{
				"test2": "test2",
			},
			expected: map[string]interface{}{
				"test":  "test",
				"test2": "test2",
			},
		},
		{
			name: "m1 and m2 not empty with same key",
			m1: map[string]interface{}{
				"test": "test",
			},
			m2: map[string]interface{}{
				"test": "test2",
			},
			expected: map[string]interface{}{
				"test": "test2",
			},
		},
		{
			name: "m1 and m2 not empty with same key and nested map",
			m1: map[string]interface{}{
				"test": map[string]interface{}{
					"test": "test",
				},
			},
			m2: map[string]interface{}{
				"test": map[string]interface{}{
					"test2": "test2",
				},
			},
			expected: map[string]interface{}{
				"test": map[string]interface{}{
					"test":  "test",
					"test2": "test2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mergeMaps(tc.m1, tc.m2)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected: %v, got: %v", tc.expected, result)
			}
		})
	}
}
