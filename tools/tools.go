// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
package tools

import (
	"fmt"
	"os"
	"strings"

	gojq "github.com/itchyny/gojq/cli"
	envsubst "github.com/nuvolaris/envsubst/cmd/envsubstmain"
	replace "github.com/nuvolaris/go-replace"
	"github.com/nuvolaris/goawk"
	"github.com/nuvolaris/gron"
	"github.com/nuvolaris/jj"
	"golang.org/x/exp/slices"
)

// available in taskfiles
// note some of them are implemented in main.go (config, retry)
var tools = []string{
	"awk", "die", "jq", "js",
	"envsubst", "wsk", "mkdir",
	"filetype", "random", "datefmt",
	"config", "retry", "urlenc",
	"replace", "base64", "validate",
	"echoif", "echoifempty", "echoifexists",
	"realpath", "zipf", "needupdate",
	"gron", "jj", "deploy",
}

// not available in taskfiles
var extraTools = []string{
	"update", "login", "help", "info", "version", "task", "plugin", "serve",
}

func IsTool(name string) bool {
	if IsUtil(name) {
		return true
	}
	for _, s := range tools {
		if s == name {
			return true
		}
	}
	return false
}

var NuvCmd = ""

func GetNuvCmd() string {
	if NuvCmd != "" {
		return NuvCmd
	}
	nuv := os.Getenv("NUV_CMD")
	if nuv != "" {
		return nuv
	}
	return ""
}

func RunTool(name string, args []string) (int, error) {
	if IsUtil(name) {
		return RunUtil(name, args)
	}
	switch name {
	case "wsk":
		//fmt.Println("=== wsk ===")
		cmd := append([]string{"wsk"}, args...)
		if err := Wsk(cmd); err != nil {
			return 1, err
		}
	case "awk":
		// fmt.Println("=== awk ===")
		os.Args = append([]string{"goawk"}, args...)
		if err := goawk.AwkMain(); err != nil {
			return 1, err
		}
	case "jq":
		os.Args = append([]string{"gojq"}, args...)
		return gojq.Run(), nil
	case "js":
		os.Args = append([]string{"goja"}, args...)
		if err := jsToolMain(); err != nil {
			return 1, err
		}
	case "envsubst":
		os.Args = append([]string{"envsubst"}, args...)
		if err := envsubst.EnvsubstMain(); err != nil {
			return 1, err
		}
	case "mkdir":
		os.Args = append([]string{"mkdir"}, args...)
		if err := Mkdirs(); err != nil {
			return 1, err
		}
	case "filetype":
		os.Args = append([]string{"mkdir"}, args...)
		if err := Filetype(); err != nil {
			return 1, err
		}
	case "random":
		if err := RandTool(args...); err != nil {
			return 1, err
		}
	case "datefmt":
		os.Args = append([]string{"datefmt"}, args...)
		if err := DateFmtTool(); err != nil {
			return 1, err
		}
	case "die":
		if len(args) > 0 {
			fmt.Println(strings.Join(args, " "))
		}
		return 1, nil
	case "urlenc":
		os.Args = append([]string{"urlenc"}, args...)
		if err := URLEncTool(); err != nil {
			return 1, err
		}
	case "replace":
		os.Args = append([]string{"replace"}, args...)
		return replace.ReplaceMain()
	case "base64":
		os.Args = append([]string{"base64"}, args...)
		if err := base64Tool(); err != nil {
			return 1, err
		}
	case "validate":
		os.Args = append([]string{"validate"}, args...)
		if err := validateTool(); err != nil {
			return 1, err
		}
	case "scan":
		os.Args = append([]string{"scan"}, args...)
		if err := scanTool(); err != nil {
			return 1, err
		}
	case "echoif":
		os.Args = append([]string{"echoif"}, args...)
		if err := echoIfTool(); err != nil {
			return 1, err
		}
	case "echoifempty":
		os.Args = append([]string{"echoifempty"}, args...)
		if err := echoIfEmptyTool(); err != nil {
			return 1, err
		}
	case "echoifexists":
		os.Args = append([]string{"echoifexists"}, args...)
		if err := echoIfExistsTool(); err != nil {
			return 1, err
		}
	case "realpath":
		os.Args = append([]string{"realpath"}, args...)
		if err := realpathTool(); err != nil {
			return 1, err
		}

	case "zipf":
		if err := zipfTool(args); err != nil {
			return 1, err
		}

	case "needupdate":
		if err := needUpdateTool(args); err != nil {
			return 1, err
		}

	case "gron":
		os.Args = append([]string{"gron"}, args...)
		return gron.GronMain()

	case "jj":
		os.Args = append([]string{"jj"}, args...)
		return jj.JJMain()

	case "deploy":
		if err := DeployTool(); err != nil {
			return 1, err
		}
	}

	return 0, nil
}

func Help() {
	fmt.Println("Available tools:")
	availableTools := append(tools, extraTools...)
	availableTools = append(availableTools, Utils...)
	slices.Sort(availableTools)
	for _, x := range availableTools {
		fmt.Printf("-%s\n", x)
	}
}
