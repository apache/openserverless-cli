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
package openserverless

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

	"github.com/apache/openserverless-cli/auth"
	"github.com/apache/openserverless-cli/config"
	"github.com/apache/openserverless-cli/tools"

	"github.com/mitchellh/go-homedir"

	_ "embed"
)

func setupCmd(me string) (string, error) {
	if os.Getenv("OPS_CMD") != "" {
		return os.Getenv("OPS_CMD"), nil
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
	//nolint:errcheck
	os.Setenv("OPS_CMD", me)
	trace("OPS_CMD:", me)
	return me, nil
}

func setupBinPath() error {
	// initialize tools (used by the shell to find myself)
	bindir, err := EnsureBindir()
	if err == nil {
		os.Setenv("PATH", fmt.Sprintf("%s%c%s", bindir, os.PathListSeparator, os.Getenv("PATH")))
		debugf("PATH=%s", os.Getenv("PATH"))
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

func info() {
	fmt.Println("NUV_VERSION:", os.Getenv("NUV_VERSION"))
	fmt.Println("NUV_BRANCH:", os.Getenv("NUV_BRANCH"))
	fmt.Println("OPS_CMD:", os.Getenv("OPS_CMD"))
	fmt.Println("OPS_BIN:", os.Getenv("OPS_BIN"))
	fmt.Println("NUV_TMP:", os.Getenv("NUV_TMP"))
	fmt.Println("OPS_HOME:", os.Getenv("OPS_HOME"))
	fmt.Println("NUV_ROOT:", os.Getenv("NUV_ROOT"))
	fmt.Println("NUV_REPO:", os.Getenv("NUV_REPO"))
	fmt.Println("NUV_PWD:", os.Getenv("NUV_PWD"))
	fmt.Println("NUV_OLARIS:", os.Getenv("NUV_OLARIS"))
	fmt.Println("NUV_ROOT_PLUGIN:", os.Getenv("NUV_ROOT_PLUGIN"))
	//fmt.Println("OPS_TOOLS:", os.Getenv("OPS_TOOLS"))
	//fmt.Println("OPS_COREUTILS:", os.Getenv("OPS_COREUTILS"))
}

// not available in taskfiles
var mainTools = []string{
	"task", "version", "info", "help", "serve", "update", "retry", "login", "config", "plugin",
}

func executeMainToolsAndExit(cmd string, args []string, nuvHome string) int {
	if cmd == "" || cmd == "-" || cmd == "task" {
		exitCode, err := Task(args[2:]...)
		if err != nil {
			log.Println(err)
		}
		return exitCode
	}

	switch cmd {
	case "version":
		fmt.Println(NuvVersion)
	case "v":
		fmt.Println(NuvVersion)

	case "info":
		info()

	case "help":
		tools.Help(mainTools)

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
		fmt.Println("OpenServerless host and auth set successfully. You are now ready to use ops!")

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
			return code
		}
		// no embeded tool found
		warn("unknown tool", "-"+cmd)
	}
	return 0
}

func Main() {
	var err error
	args := os.Args

	// disable log prefix
	log.SetFlags(0)

	// set default runtime.json
	if os.Getenv("WSK_RUNTIMES_JSON") == "" {
		os.Setenv("WSK_RUNTIMES_JSON", WSK_RUNTIMES_JSON)
		trace("WSK_RUNTIMES_JSON len=", len(WSK_RUNTIMES_JSON))
	}

	// set runtime version as environment variable
	if os.Getenv("NUV_VERSION") != "" {
		NuvVersion = os.Getenv("NUV_VERSION")
	} else {
		NuvVersion = strings.TrimSpace(NuvVersion)
		os.Setenv("NUV_VERSION", NuvVersion)
	}

	// setup OPS_CMD
	me := args[0]
	if strings.Contains("ops ops.exe nuv nuv.exe", filepath.Base(me)) {
		_, err = setupCmd(me)
		if err != nil {
			log.Fatalf("cannot setup cmd: %s", err.Error())
		}
	}

	// setup home
	nuvHome := os.Getenv("OPS_HOME")
	if nuvHome == "" {
		nuvHome, err = homedir.Expand("~/.nuv")
	}
	if err != nil {
		log.Fatalf("cannot setup home: %s", err.Error())
	}
	os.Setenv("OPS_HOME", nuvHome)

	// add ~/.nuv/<os>-<arch>/bin to the path at the beginning
	err = setupBinPath()
	if err != nil {
		log.Fatalf("cannot setup PATH: %s", err.Error())
	}

	// ensure there is ~/.nuv/tmp
	err = setupTmp()
	if err != nil {
		log.Fatalf("cannot setup NUV_TMP: %s", err.Error())
	}

	//  setup the NUV_PWD variable
	err = setNuvPwdEnv()
	if err != nil {
		log.Fatalf("cannot setup NUV_PWD: %s", err.Error())
	}

	// setup the envvar for the embedded tools
	os.Setenv("OPS_TOOLS", strings.Join(append(mainTools, tools.ToolList...), " "))

	// NUV_REPO && NUV_ROOT_PLUGIN
	getNuvRepo()
	setNuvRootPluginEnv()

	// Check if olaris exists. If not, download tasks
	olarisDir, err := getNuvRoot()
	if err != nil {
		olarisDir := joinpath(joinpath(nuvHome, getNuvBranch()), "olaris")
		if !isDir(olarisDir) {
			log.Println("Welcome to ops! Setting up...")
			olarisDir, err = pullTasks(true, true)
			if err != nil {
				log.Fatalf("cannot locate or download NUV_ROOT: %s", err.Error())
			}
			// just updated, do not repeat
			if len(args) == 2 && args[1] == "-update" {
				os.Exit(0)
			}
		} else {
			// check if olaris was recently updated
			checkUpdated(nuvHome, 24*time.Hour)
		}
	}
	if err = setNuvOlarisHash(olarisDir); err != nil {
		os.Setenv("NUV_OLARIS", "<local>")
	}

	// set the enviroment variables from the config
	nuvRootDir := getRootDirOrExit()
	debug("nuvRootDir", nuvRootDir)
	err = setAllConfigEnvVars(nuvRootDir, nuvHome)
	if err != nil {
		log.Fatalf("cannot apply env vars from configs: %s", err.Error())
	}

	// preflight checks - we need at least ssh curl to proceed
	if err := preflightChecks(); err != nil {
		log.Fatalf("failed preflight check: %s", err.Error())
	}

	// in case args[1] is a wsk wrapper command invoke it and exit
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
			os.Exit(0)
		}
	}

	// first argument with prefix "-" is considered an embedded tool
	// using "-" or "--" or "-task" invokes the embedded task
	if len(args) > 1 && len(args[1]) > 0 && args[1][0] == '-' {
		cmd := args[1][1:]
		trace("executing embedded tool", cmd, args)
		// execute the embeded tool and exit
		exitCode := executeMainToolsAndExit(cmd, args, nuvHome)
		os.Exit(exitCode)
	}

	if err := runNuv(nuvRootDir, args); err != nil {
		log.Fatalf("task execution error: %s", err.Error())
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

func setNuvPwdEnv() error {
	if os.Getenv("NUV_PWD") == "" {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		//nolint:errcheck
		os.Setenv("NUV_PWD", dir)
	}
	trace("set NUV_PWD", os.Getenv("NUV_PWD"))
	return nil
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
