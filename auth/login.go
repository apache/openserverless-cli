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

package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/apache/openserverless-cli/config"
	"github.com/zalando/go-keyring"
)

type LoginResult struct {
	Login   string
	Auth    string
	ApiHost string
}

const usage = `Usage:
ops -login <apihost> [<user>]

Login to an OpenServerless instance. If no user is specified, the default user "nuvolaris" is used.
You can set the environment variables OPS_APIHOST and OPS_USER to avoid specifying them on the command line.
You can set OPS_PASSWORD to avoid entering the password interactively.

Options:
  -h, --help   Show usage`

const whiskLoginPath = "/api/v1/web/whisk-system/nuv/login"
const defaultUser = "nuvolaris"
const opsSecretServiceName = "nuvolaris"

func LoginCmd() (*LoginResult, error) {

	// enable log output if requested
	if os.Getenv("DEBUG")+os.Getenv("TRACE") != "" {
		log.SetOutput(os.Stdout)
	}

	flag := flag.NewFlagSet("-login", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Println(usage)
	}

	var helpFlag bool
	flag.BoolVar(&helpFlag, "h", false, "Show usage")
	flag.BoolVar(&helpFlag, "help", false, "Show usage")
	err := flag.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if helpFlag {
		flag.Usage()
		return nil, nil
	}

	args := flag.Args()

	if len(args) == 0 && os.Getenv("OPS_APIHOST") == "" {
		flag.Usage()
		return nil, errors.New("missing apihost")
	}

	apihost := os.Getenv("OPS_APIHOST")
	if apihost == "" {
		apihost = args[0]
	}
	url := apihost + whiskLoginPath

	apihost = ensureSchema(apihost)

	// try to get the user from the environment
	user := os.Getenv("OPS_USER")
	if user == "" {
		// if env var not set, try to get it from the command line
		if os.Getenv("OPS_APIHOST") != "" {
			// if apihost env var was set, treat the first arg as the user
			if len(args) > 0 {
				user = args[0]
			}
		} else {
			// if apihost env var was not set, treat the second arg as the user
			if len(args) > 1 {
				user = args[1]
			}
		}
	}

	// if still not set, use the default user
	if user == "" {
		fmt.Println("Using the default user:", defaultUser)
		user = defaultUser
	}

	fmt.Println("Logging in", apihost, "as", user)

	password := os.Getenv("OPS_PASSWORD")
	if password == "" {
		fmt.Print("Enter Password: ")
		pwd, err := AskPassword()
		if err != nil {
			return nil, err
		}
		password = pwd
		fmt.Println()
	}

	creds, err := doLogin(url, user, password)
	if err != nil {
		return nil, err
	}

	if _, ok := creds["AUTH"]; !ok {
		return nil, errors.New("missing AUTH token from login response")
	}

	opsHome := os.Getenv("OPS_HOME")
	if opsHome == "" {
		return nil, fmt.Errorf("OPS_HOME not defined")
	}

	configMap, err := config.NewConfigMapBuilder().
		WithConfigJson(filepath.Join(opsHome, "config.json")).
		Build()

	if err != nil {
		return nil, err
	}

	for k, v := range creds {
		if err := configMap.Insert(k, v); err != nil {
			return nil, err
		}
	}

	if err := configMap.Insert("STATUS_LOGGED_USER", user); err != nil {
		log.Println("[Warning] Failed to insert STATUS_LOGGED_USER")
	}

	err = configMap.SaveConfig()
	if err != nil {
		return nil, err
	}

	// if err := storeCredentials(creds); err != nil {
	// 	return nil, err
	// }

	// auth, err := keyring.Get(opsSecretServiceName, "AUTH")
	// if err != nil {
	// 	return nil, err
	// }

	return &LoginResult{
		Login:   user,
		Auth:    creds["AUTH"],
		ApiHost: apihost,
	}, nil
}

func ensureSchema(apihost string) string {
	if !strings.HasPrefix(apihost, "http://") && !strings.HasPrefix(apihost, "https://") {
		if apihost == "localhost" {
			apihost = "http://" + apihost
		} else {
			apihost = "https://" + apihost
		}
	}
	return apihost
}

func doLogin(url, user, password string) (map[string]string, error) {
	data := map[string]string{
		"login":    user,
		"password": password,
	}
	loginJson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(loginJson))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("login failed with status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("login failed (%d): %s", resp.StatusCode, string(body))
	}

	var creds map[string]string
	err = json.NewDecoder(resp.Body).Decode(&creds)
	if err != nil {
		return nil, errors.New("failed to decode response from login request")
	}

	return creds, nil
}

func storeCredentials(creds map[string]string) error {
	for k, v := range creds {
		err := keyring.Set(opsSecretServiceName, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
