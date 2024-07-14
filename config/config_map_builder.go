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
	"log"
	"os"
)

type configMapBuilder struct {
	configJsonPath string
	nuvRootPath    string
	pluginNuvRoots map[string]string
}

func NewConfigMapBuilder() *configMapBuilder {
	return &configMapBuilder{
		pluginNuvRoots: make(map[string]string),
	}
}

// WithConfigJson adds a config.json file that will be read and used
// to build a ConfigMap. If there both a config.json and a nuvroot.json are
// added, the 2 configs will be merged. It assumes the input file
// is valid.
func (b *configMapBuilder) WithConfigJson(file string) *configMapBuilder {
	b.configJsonPath = file
	return b
}

// WithNuvRoot works like WithConfigJson, with the difference that
// the NuvRoot is read and only it's inner "config":{} object is parsed
// ignoring the rest of the content.
func (b *configMapBuilder) WithNuvRoot(file string) *configMapBuilder {
	b.nuvRootPath = file
	return b
}

// WithPluginNuvRoots works like WithPluginNuvRoot, but it allows to add multiple
// plugin nuvroot.json files at once.
func (b *configMapBuilder) WithPluginNuvRoots(nrts map[string]string) *configMapBuilder {
	b.pluginNuvRoots = nrts
	return b
}

func (b *configMapBuilder) Build() (ConfigMap, error) {
	configJsonMap, err := readConfig(b.configJsonPath, fromConfigJson)
	if err != nil {
		return ConfigMap{}, err
	}

	nuvRootMap, err := readConfig(b.nuvRootPath, fromNuvRoot)
	if err != nil {
		return ConfigMap{}, err
	}

	pluginNuvRootConfigs := make(map[string]map[string]interface{}, 0)
	for plgName, nuvRootPath := range b.pluginNuvRoots {
		pluginNuvRootMap, err := readConfig(nuvRootPath, fromNuvRoot)
		if err != nil {
			return ConfigMap{}, err
		}
		pluginNuvRootConfigs[plgName] = pluginNuvRootMap
	}

	return ConfigMap{
		pluginNuvRootConfigs: pluginNuvRootConfigs,
		nuvRootConfig:        nuvRootMap,
		config:               configJsonMap,
		configPath:           b.configJsonPath,
	}, nil
}

func readConfig(path string, read func(string) (map[string]interface{}, error)) (map[string]interface{}, error) {
	if path == "" {
		return make(map[string]interface{}), nil
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}
	if err != nil {
		return nil, err
	}

	cMap, err := read(path)
	if err != nil {
		return nil, err
	}

	return cMap, nil
}

func fromConfigJson(configPath string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	json_buf, err := os.ReadFile(configPath)
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal(json_buf, &data); err != nil {
		if data == nil {
			return data, err
		}
		log.Println("config.json parsed with an error", err)
	}

	return data, nil
}

func fromNuvRoot(nuvRootPath string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	json_buf, err := os.ReadFile(nuvRootPath)
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal(json_buf, &data); err != nil {
		if data == nil {
			return data, err
		}
		log.Println("nuvroot.json parsed with an error", err)
	}

	cm, ok := data["config"].(map[string]interface{})
	if !ok {
		return nil, nil
	}
	return cm, nil
}
