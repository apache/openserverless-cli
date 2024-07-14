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
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/nuvolaris/nuv/auth"
	"github.com/nuvolaris/nuv/config"
	"github.com/nuvolaris/nuv/tools"

	_ "embed"
)

func setupCmd(me string) (string, error) {
	if os.Getenv("NUV_CMD") != "" {
		return os.Getenv("NUV_CMD"), nil
	}

	// look in path
	me, err := exec.LookPath(me)
	if err != nil {
		return "", err
	}
	trace("found", me)

	// resolve links
	fileInfo, err := os.Lstat(me)
	if err != nil {
		return "", err
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		me, err = os.Readlink(me)
		if err != nil {
			return "", err
		}
		trace("resolving link to", me)
	}

	// get the absolute path
	me, err = filepath.Abs(me)
	if err != nil {
		return "", err
	}
	trace("ME:", me)
	//nolint:errcheck
	os.Setenv("NUV_CMD", me)
	return me, nil
}

func setupBinPath(cmd string) {
	// initialize tools (used by the shell to find myself)
	if os.Getenv("NUV_BIN") == "" {
		os.Setenv("NUV_BIN", filepath.Dir(cmd))
	}
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", os.Getenv("NUV_BIN"), os.PathListSeparator, os.Getenv("PATH")))
	debugf("PATH=%s", os.Getenv("PATH"))

	//subpath := fmt.Sprintf("\"%s\"%c\"%s\"", os.Getenv("NUV_BIN"), os.PathListSeparator, joinpath(os.Getenv("NUV_BIN"), runtime.GOOS+"-"+runtime.GOARCH))
	//os.Setenv("PATH", fmt.Sprintf("%s%c%s", subpath, os.PathListSeparator, os.Getenv("PATH")))
}

func info() {
	fmt.Println("NUV_VERSION:", NuvVersion)
	fmt.Println("NUV_BRANCH:", getNuvBranch())
	fmt.Println("NUV_CMD:", tools.GetNuvCmd())
	fmt.Println("NUV_REPO:", getNuvRepo())
	fmt.Println("NUV_BIN:", os.Getenv("NUV_BIN"))
	fmt.Println("NUV_TMP:", os.Getenv("NUV_TMP"))
	root, _ := getNuvRoot()
	fmt.Println("NUV_ROOT:", root)
	fmt.Println("NUV_PWD:", os.Getenv("NUV_PWD"))
	fmt.Println("NUV_OLARIS:", os.Getenv("NUV_OLARIS"))
}

//go:embed runtimes.json
var WSK_RUNTIMES_JSON string

func main() {
	// set runtime version as environment variable
	if os.Getenv("WSK_RUNTIMES_JSON") == "" {
		os.Setenv("WSK_RUNTIMES_JSON", WSK_RUNTIMES_JSON)
		trace(WSK_RUNTIMES_JSON)
	}

	// set version
	if os.Getenv("NUV_VERSION") != "" {
		NuvVersion = os.Getenv("NUV_VERSION")
	}

	// disable log prefix
	log.SetFlags(0)

	var err error
	me := os.Args[0]
	if filepath.Base(me) == "nuv" || filepath.Base(me) == "nuv.exe" {
		tools.NuvCmd, err = setupCmd(me)
		if err != nil {
			log.Fatalf("cannot setup cmd: %s", err.Error())
		}
		setupBinPath(tools.NuvCmd)
	}

	nuvHome, err := homedir.Expand("~/.nuv")
	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}

	// Check if olaris exists. If not, run `-update` to auto setup
	olarisDir := joinpath(joinpath(nuvHome, getNuvBranch()), "olaris")
	if !isDir(olarisDir) {
		if !(len(os.Args) == 2 && os.Args[1] == "-update") {
			log.Println("Welcome to nuv! Setting up...")
			dir, err := pullTasks(true, true)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			if err := setNuvOlarisHash(dir); err != nil {
				warn("unable to set NUV_OLARIS...", err.Error())
			}
		}
	}

	setupTmp()
	setNuvPwdEnv()
	setNuvRootPluginEnv()
	if err := setNuvOlarisHash(olarisDir); err != nil {
		warn("unable to set NUV_OLARIS...", err.Error())
	}

	nuvRootDir := getRootDirOrExit()
	debug("nuvRootDir", nuvRootDir)
	err = setAllConfigEnvVars(nuvRootDir, nuvHome)
	if err != nil {
		log.Fatalf("cannot apply env vars from configs: %s", err.Error())
	}

	// first argument with prefix "-" is an embedded tool
	// using "-" or "--" or "-task" invokes embedded task
	trace("OS args:", os.Args)
	args := os.Args

	if len(args) > 1 && len(args[1]) > 0 && args[1][0] == '-' {
		cmd := args[1][1:]
		if cmd == "" || cmd == "-" || cmd == "task" {
			exitCode, err := Task(args[2:]...)
			if err != nil {
				log.Println(err)
			}
			os.Exit(exitCode)
		}

		switch cmd {
		case "version":
			fmt.Println(NuvVersion)
		case "v":
			fmt.Println(NuvVersion)

		case "info":
			info()

		case "help":
			tools.Help()

		case "serve":
			nuvRootDir := getRootDirOrExit()
			if err := Serve(nuvRootDir, args[1:]); err != nil {
				log.Fatalf("error: %v", err)
			}

		case "update":
			// ok no up, nor down, let's download it
			dir, err := pullTasks(true, true)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			if err := setNuvOlarisHash(dir); err != nil {
				log.Fatal("unable to set NUV_OLARIS...", err.Error())
			}

		case "retry":
			if err := tools.ExpBackoffRetry(args[1:]); err != nil {
				log.Fatalf("error: %s", err.Error())
			}

		case "login":
			os.Args = args[1:]
			loginResult, err := auth.LoginCmd()
			if err != nil {
				log.Fatalf("error: %s", err.Error())
			}

			if loginResult == nil {
				os.Exit(1)
			}

			fmt.Println("Successfully logged in as " + loginResult.Login + ".")
			if err := wskPropertySet(loginResult.ApiHost, loginResult.Auth); err != nil {
				log.Fatalf("error: %s", err.Error())
			}
			fmt.Println("Nuvolaris host and auth set successfully. You are now ready to use nuv -wsk!")

		case "config":
			os.Args = args[1:]
			nuvRootPath := joinpath(getRootDirOrExit(), NUVROOT)
			configPath := joinpath(nuvHome, CONFIGFILE)
			configMap, err := buildConfigMap(nuvRootPath, configPath)
			if err != nil {
				log.Fatalf("error: %s", err.Error())
			}

			if err := config.ConfigTool(*configMap); err != nil {
				log.Fatalf("error: %s", err.Error())
			}

		case "plugin":
			os.Args = args[1:]
			if err := pluginTool(); err != nil {
				log.Fatalf("error: %s", err.Error())
			}

		default:
			// check if it is an embedded to and invoke it
			if tools.IsTool(cmd) {
				code, err := tools.RunTool(cmd, args[2:])
				if err != nil {
					log.Print(err.Error())
				}
				os.Exit(code)
			}
			// no embeded tool found
			warn("unknown tool", "-"+cmd)
		}
		os.Exit(0)
	}

	// check if olaris was recently updated
	checkUpdated(nuvHome, 24*time.Hour)

	// in case args[1] is a wsk wrapper command
	// invoke it and exit
	if len(args) > 1 {
		if cmd, ok := IsWskWrapperCommand(args[1]); ok {
			trace("wsk wrapper command")
			debug("extracted cmd", cmd)
			rest := args[2:]
			debug("extracted args", rest)

			// if "invoke" is in the command, parse all a=b into -p a b
			if (len(cmd) > 2 && cmd[2] == "invoke") || slices.Contains(rest, "invoke") {
				rest = parseInvokeArgs(rest)
			}

			if err := tools.Wsk(cmd, rest...); err != nil {
				log.Fatalf("error: %s", err.Error())
			}
			return //(skip runNuv)
		}
	}

	// preflight checks
	if err := preflightChecks(); err != nil {
		log.Fatalf("[PREFLIGHT CHECK] error: %s", err.Error())
	}
	// ***************

	if err := runNuv(nuvRootDir, args); err != nil {
		log.Fatalf("error: %s", err.Error())
	}
}

// parse all a=b into -p a b
func parseInvokeArgs(rest []string) []string {
	trace("parsing invoke args")
	args := []string{}

	for _, arg := range rest {
		if strings.Contains(arg, "=") {
			kv := strings.Split(arg, "=")
			p := []string{"-p", kv[0], kv[1]}
			args = append(args, p...)
		} else {
			args = append(args, arg)
		}
	}

	debug("parsed invoke args", args)
	return args
}

// getRootDirOrExit returns the olaris dir or exits (Fatal) if not found
func getRootDirOrExit() string {
	dir, err := getNuvRoot()
	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}
	return dir
}

func setAllConfigEnvVars(nuvRootDir string, configDir string) error {
	trace("setting all config env vars")

	configMap, err := buildConfigMap(joinpath(nuvRootDir, NUVROOT), joinpath(configDir, CONFIGFILE))
	if err != nil {
		return err
	}

	kv := configMap.Flatten()
	for k, v := range kv {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
		debug("env var set", k, v)
	}

	return nil
}

func wskPropertySet(apihost, auth string) error {
	args := []string{"property", "set", "--apihost", apihost, "--auth", auth}
	cmd := append([]string{"wsk"}, args...)
	if err := tools.Wsk(cmd); err != nil {
		return err
	}
	return nil
}

func runNuv(baseDir string, args []string) error {
	err := Nuv(baseDir, args[1:])
	if err == nil {
		return nil
	}

	// If the task is not found,
	// fallback to plugins
	var taskNotFoundErr *TaskNotFoundErr
	if errors.As(err, &taskNotFoundErr) {
		trace("task not found, looking for plugin:", args[1])
		plgDir, err := findTaskInPlugins(args[1])
		if err != nil {
			return taskNotFoundErr
		}

		debug("Found plugin", plgDir)
		if err := Nuv(plgDir, args[2:]); err != nil {
			log.Fatalf("error: %s", err.Error())
		}
		return nil
	}

	return err
}

func setNuvRootPluginEnv() {
	if os.Getenv("NUV_ROOT_PLUGIN") == "" {
		//nolint:errcheck
		os.Setenv("NUV_ROOT_PLUGIN", os.Getenv("NUV_PWD"))
	}
	trace("set NUV_ROOT_PLUGIN", os.Getenv("NUV_ROOT_PLUGIN"))
}

func setNuvPwdEnv() {
	if os.Getenv("NUV_PWD") == "" {
		dir, _ := os.Getwd()
		//nolint:errcheck
		os.Setenv("NUV_PWD", dir)
	}
	trace("set NUV_PWD", os.Getenv("NUV_PWD"))
}

func buildConfigMap(nuvRootPath string, configPath string) (*config.ConfigMap, error) {
	plgNuvRootMap, err := GetNuvRootPlugins()
	if err != nil {
		return nil, err
	}

	configMap, err := config.NewConfigMapBuilder().
		WithNuvRoot(nuvRootPath).
		WithConfigJson(configPath).
		WithPluginNuvRoots(plgNuvRootMap).
		Build()

	if err != nil {
		return nil, err
	}

	return &configMap, nil
}

func IsWskWrapperCommand(name string) ([]string, bool) {
	wskWrapperCommands := map[string][]string{
		"action":     {"wsk", "action"},
		"activation": {"wsk", "activation"},
		"invoke":     {"wsk", "action", "invoke", "-r"},
		"logs":       {"wsk", "activation", "logs"},
		"package":    {"wsk", "package"},
		"result":     {"wsk", "activation", "result"},
		"rule":       {"wsk", "rule"},
		"trigger":    {"wsk", "trigger"},
		"url":        {"wsk", "action", "get", "--url"},
	}

	cmd, ok := wskWrapperCommands[name]
	return cmd, ok
}
