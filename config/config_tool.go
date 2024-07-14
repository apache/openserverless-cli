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
	"flag"
	"fmt"
	"os"
	"strings"
)

func printConfigToolUsage() {
	fmt.Print(`Usage:
nuv -config [options] [KEY | KEY=VALUE [KEY=VALUE ...]]

Set config values passed as key-value pairs. 
If the keys are passed (without '='), their values are printed, if they exists.

If you want to override a value, pass KEY="". This can be used to disable values in nuvroot.json.
Removing values from nuvroot.json is not supported, disable them instead.

-h, --help    	show this help
-r, --remove    remove config values by passing keys
-d, --dump    	dump the configs
`)
}

func ConfigTool(configMap ConfigMap) error {
	flag := flag.NewFlagSet("config", flag.ExitOnError)
	var helpFlag bool
	var dumpFlag bool
	var removeFlag bool

	flag.Usage = printConfigToolUsage

	flag.BoolVar(&helpFlag, "h", false, "show this help")
	flag.BoolVar(&helpFlag, "help", false, "show this help")
	flag.BoolVar(&dumpFlag, "dump", false, "dump the config file")
	flag.BoolVar(&dumpFlag, "d", false, "dump the config file")
	flag.BoolVar(&removeFlag, "remove", false, "remove config values")
	flag.BoolVar(&removeFlag, "r", false, "remove config values")

	err := flag.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	if helpFlag {
		flag.Usage()
		return nil
	}

	if dumpFlag {
		dumped := configMap.Flatten()
		for k, v := range dumped {
			fmt.Printf("%s=%s\n", k, v)
		}
		return nil
	}

	// Get the input string from the remaining command line arguments
	input := flag.Args()

	if len(input) == 0 {
		flag.Usage()
		return nil
	}

	var cErr error
	noAssigns := inputWithoutAssigns(input)

	if removeFlag {
		cErr = removeInConfigJSON(configMap, input)
	} else if noAssigns {
		// print the values and return
		return printValues(configMap, input)
	} else {
		cErr = insertInConfigJSON(configMap, input)
	}

	if cErr != nil {
		return cErr
	}

	return configMap.SaveConfig()
}

func insertInConfigJSON(configMap ConfigMap, input []string) error {
	// Parse the input string into key-value pairs
	pairs, err := buildInputKVMap(input)
	if err != nil {
		return err
	}
	for k, v := range pairs {
		if err := configMap.Insert(k, v); err != nil {
			return err
		}
	}

	return nil
}

func removeInConfigJSON(configMap ConfigMap, input []string) error {
	for _, k := range input {
		if err := configMap.Delete(k); err != nil {
			return err
		}
	}
	return nil
}

func inputWithoutAssigns(input []string) bool {
	for _, arg := range input {
		if strings.Contains(arg, "=") {
			return false
		}
	}
	return true
}

func printValues(configMap ConfigMap, keys []string) error {
	for _, k := range keys {
		val, err := configMap.Get(k)
		if err != nil {
			return err
		}
		fmt.Println(val)
	}
	return nil
}

type keyValues map[string]string

func (kv *keyValues) String() string {
	return fmt.Sprintf("%v", *kv)
}

func (kv *keyValues) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid key-value pair: %q", value)
	}
	key := parts[0]
	val := parts[1]

	if key == "" || val == "" {
		return fmt.Errorf("invalid key-value pair: %q", value)
	}

	if *kv == nil {
		*kv = make(keyValues)
	}
	(*kv)[key] = val
	return nil
}

func buildInputKVMap(pairs []string) (keyValues, error) {
	var kv keyValues

	if len(pairs) == 0 {
		return nil, fmt.Errorf("no key-value pairs provided")
	}

	for _, pair := range pairs {
		if err := kv.Set(pair); err != nil {
			return nil, err
		}
	}
	return kv, nil
}
