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
	"log"
	"os"
	"strconv"
	"strings"
)

// A ConfigMap is a map where the keys are in the form of: A_KEY_WITH_UNDERSCORES.
// The map splits the key by the underscores and creates a nested map that
// represents the key. For example, the key "A_KEY_WITH_UNDERSCORES" would be
// represented as:
//
//	{
//		"a": {
//			"key": {
//				"with": {
//					"underscores": "value",
//				},
//			},
//		},
//	}
//
// To interact with the ConfigMap, use the Insert, Get, and Delete by passing
// keys in the form above. Only the config map is modified by these functions.
// The nuvRootConfig map is only used to read the config keys in nuvroot.json.
// The pluginNuvRootConfigs map is only used to read the config keys in
// plugins (from their nuvroot.json). It is a map that maps the plugin name to
// the config map for that plugin.
type ConfigMap struct {
	pluginNuvRootConfigs map[string]map[string]interface{}
	nuvRootConfig        map[string]interface{}
	config               map[string]interface{}
	configPath           string
}

// Insert inserts a key and value into the ConfigMap. If the key already exists,
// the value is overwritten. The expected key format is A_KEY_WITH_UNDERSCORES.
func (c *ConfigMap) Insert(key string, value string) error {
	keys, err := parseKey(strings.ToLower(key))
	if err != nil {
		return err
	}

	currentMap := c.config
	lastIndex := len(keys) - 1
	for i, subKey := range keys {
		// If we are at the last key, set the value
		if i == lastIndex {
			v, err := parseValue(value)
			if err != nil {
				return err
			}

			currentMap[subKey] = v
		} else {
			// If the sub-map doesn't exist, create it
			if _, ok := currentMap[subKey]; !ok {
				currentMap[subKey] = make(map[string]interface{})
			}
			// Update the current map to the sub-map
			m, ok := currentMap[subKey].(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid key: '%s' - '%s' is already being used for a value", key, subKey)
			}
			currentMap = m
		}
	}

	return nil
}

func (c *ConfigMap) Get(key string) (string, error) {
	cmap := c.Flatten()

	val, ok := cmap[key]
	if !ok {
		return "", fmt.Errorf("invalid key: '%s' - key does not exist", key)
	}

	return val, nil
}

func (c *ConfigMap) Delete(key string) error {
	delFunc := func(config map[string]interface{}, key string) bool {
		if _, ok := config[key]; !ok {
			return false
		}

		delete(config, key)
		return true
	}
	keys, err := parseKey(strings.ToLower(key))
	if err != nil {
		return err
	}

	ok := visit(c.config, 0, keys, delFunc)
	if !ok {
		return fmt.Errorf("invalid key: '%s' - key does not exist in config.json", key)
	}
	return nil
}

func (c *ConfigMap) Flatten() map[string]string {
	outputMap := make(map[string]string)

	merged := mergeMaps(c.nuvRootConfig, c.config)

	for name, pluginConfig := range c.pluginNuvRootConfigs {
		// edge case: check that merged does not contain name already
		if _, ok := merged[name]; ok {
			log.Printf("config has key with same name as plugin %s. Plugin config will be ignored.", name)
			continue
		}

		merged[name] = pluginConfig
	}

	flatten("", merged, outputMap)

	return outputMap
}

func (c *ConfigMap) SaveConfig() error {
	var configJSON, err = json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.configPath, configJSON, 0644)
}

// ///
func flatten(prefix string, inputMap map[string]interface{}, outputMap map[string]string) {
	if len(prefix) > 0 {
		prefix += "_"
	}
	for k, v := range inputMap {
		key := strings.ToUpper(prefix + k)
		switch child := v.(type) {
		case map[string]interface{}:
			flatten(key, child, outputMap)
		default:
			outputMap[key] = fmt.Sprintf("%v", v)
		}
	}
}

type configOperationFunc func(config map[string]interface{}, key string) bool

func visit(config map[string]interface{}, index int, keys []string, f configOperationFunc) bool {
	// base case: if the key is the last key in the list, call the function f
	if index == len(keys)-1 {
		return f(config, keys[index])
	}

	// recursive case: if the key is not the last key in the list, call visit on the next key (if cast ok)
	conf, ok := config[keys[index]].(map[string]interface{})
	if !ok {
		return false
	}
	success := visit(conf, index+1, keys, f)
	// if the parent map is empty, clean up
	if success && len(conf) == 0 {
		delete(config, keys[index])
	}
	return success
}

func parseKey(key string) ([]string, error) {
	parts := strings.Split(key, "_")
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid key: %s", key)
		}
	}
	return parts, nil
}

/*
VALUEs are parsed in the following way:

  - try to parse as a jsos first, and if it is a json, store as a json
  - then try to parse as a number, and if it is a (float) number store as a number
  - then try to parse as true or false and store as a boolean
  - then check if it's null and store as a null
  - otherwise store as a string
*/
func parseValue(value string) (interface{}, error) {
	// Try to parse as json
	var jsonValue interface{}
	if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
		return jsonValue, nil
	}

	// Try to parse as a integer with strconv
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue, nil
	}

	// Try to parse as a float with strconv
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue, nil
	}

	// Try to parse as a boolean
	if value == "true" || value == "false" {
		return value == "true", nil
	}

	// Try to parse as null
	if value == "null" {
		return nil, nil
	}

	// Otherwise, return the string
	return value, nil
}

// mergeMaps merges map2 into map1 overwriting any values in map1 with values from map2
// when there are conflicts. It returns the merged map.
func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	if len(map1) == 0 {
		return map2
	}
	if len(map2) == 0 {
		return map1
	}

	mergedMap := make(map[string]interface{})

	for key, value := range map1 {

		map2Value, ok := map2[key]
		// key doesn't exist in map2 so add it to the merged map
		if !ok {
			mergedMap[key] = value
			continue
		}

		// key exists in map2 but map1 value is NOT a map, so add value from map2
		mapFromMap1, ok := value.(map[string]interface{})
		if !ok {
			mergedMap[key] = map2Value
			continue
		}

		mapFromMap2, ok := map2Value.(map[string]interface{})
		// key exists in map2, map1 value IS a map but map2 value is not, so overwrite with map2
		if !ok {
			mergedMap[key] = mapFromMap2
			continue
		}

		// key exists in map2, map1 value IS a map, map2 value IS a map, so merge recursively
		mergedMap[key] = mergeMaps(mapFromMap1, mapFromMap2)
	}

	// add any keys that exist in map2 but not in map1
	for key, value := range map2 {
		if _, ok := mergedMap[key]; !ok {
			mergedMap[key] = value
		}
	}

	return mergedMap
}
