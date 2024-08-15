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
	"embed"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	gojq "github.com/itchyny/gojq/cli"
	envsubst "github.com/nuvolaris/envsubst/cmd/envsubstmain"
	replace "github.com/nuvolaris/go-replace"
	"github.com/nuvolaris/goawk"
	"github.com/nuvolaris/gron"
	"github.com/nuvolaris/jj"
	"golang.org/x/exp/slices"
)

//go:embed *.md
var markDownHelp embed.FS

var tracing = os.Getenv("TRACE") != ""

func trace(args ...any) {
	if tracing {
		log.Println(append([]any{"TRACE: "}, args...))
	}
}

// shared with main
func GetOS() string {
	res := os.Getenv("__OS")
	if res == "" {
		res = runtime.GOOS
	}
	return res
}

func GetARCH() string {
	res := os.Getenv("__ARCH")
	if res == "" {
		res = runtime.GOARCH
	}
	return res
}

type Tool struct {
	Name   string
	HasDoc bool
}

// available in taskfiles
// note some of them are implemented in main.go (config, retry)
// Note the comment with @DOC which is used to generate the list of tools in documentation
// put one here if you add a markdown file for a tool
var ToolList = []Tool{
	{"wsk", false},
	{"awk", false},
	{"jq", false},
	{"sh", true}, // @DOC
	{"envsubst", false},
	{"filetype", true}, // @DOC
	{"random", true},   // @DOC
	{"datefmt", true},  // @DOC
	{"die", false},
	{"urlenc", true}, // @DOC
	{"replace", false},
	{"base64", true},       // @DOC
	{"validate", true},     // @DOC
	{"echoif", true},       // @DOC
	{"echoifempty", true},  // @DOC
	{"echoifexists", true}, // @DOC
	{"needupdate", true},   // @DOC
	{"gron", false},
	{"jj", false},
	{"rename", true},     // @DOC
	{"remove", true},     // @DOC
	{"executable", true}, // @DOC
	{"empty", true},      // @DOC
	{"extract", true},    // @DOC
}

func IsTool(name string) bool {
	for _, s := range ToolList {
		if s.Name == name {
			return true
		}
	}
	return false
}

func RunTool(name string, args []string) (int, error) {
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

	case "sh":
		os.Args = append([]string{"sh"}, args...)
		return Sh()

	case "envsubst":
		os.Args = append([]string{"envsubst"}, args...)
		if err := envsubst.EnvsubstMain(); err != nil {
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
		if err := DateFmtTool(append([]string{"datefmt"}, args...)); err != nil {
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

	case "rename":
		os.Args = append([]string{"rename"}, args...)
		return Rename()

	case "remove":
		os.Args = append([]string{"remove"}, args...)
		return Remove()

	case "executable":
		os.Args = append([]string{"executable"}, args...)
		return Executable()

	case "extract":
		os.Args = append([]string{"extract"}, args...)
		return Extract()

	case "empty":
		os.Args = append([]string{"empty"}, args...)
		return Empty()

	default:
		return 1, fmt.Errorf("unknown tool")
	}
	return 0, nil
}

func MergeToolsList(mainTools []string) []string {
	availableTools := append(mainTools)
	for _, tool := range ToolList {
		availableTools = append(availableTools, tool.Name)
	}
	return availableTools
}

func Help(mainTools []string) {
	fmt.Println("Tools (use -<tool> -h for help):")
	availableTools := MergeToolsList(mainTools)
	slices.Sort(availableTools)
	for _, x := range availableTools {
		fmt.Printf("-%s\n", x)
	}
}

func GetMarkDown(toolName string) (string, error) {
	// extract markdown from embedded resource
	markDownFile := fmt.Sprintf("%s.md", toolName)
	fileData, err := markDownHelp.ReadFile(markDownFile)
	if err != nil {
		return "", err
	}
	return string(fileData), nil
}

func MarkdownHelp(toolName string) string {
	// extract markdown from embedded resource
	fileData, err := GetMarkDown(toolName)
	if err != nil {
		return ""
	}
	result := string(fileData)
	// convert to markdown
	help := MarkdownToText(result)

	// return opts and help markdown
	return help
}
