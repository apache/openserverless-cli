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
	"strings"
	"time"

	"github.com/apache/openserverless-cli/auth"
	"github.com/apache/openserverless-cli/config"
	"github.com/apache/openserverless-cli/tools"
	"golang.org/x/exp/slices"

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
		me, err = filepath.EvalSymlinks(me)
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

func setOpsRootPluginEnv() {
	if os.Getenv("OPS_ROOT_PLUGIN") == "" {
		//nolint:errcheck
		os.Setenv("OPS_ROOT_PLUGIN", os.Getenv("OPS_PWD"))
	}
	trace("set OPS_ROOT_PLUGIN", os.Getenv("OPS_ROOT_PLUGIN"))
}

func info() {
	fmt.Println("OPS & OPS_CMD:", os.Getenv("OPS_CMD"))
	fmt.Println("OPS_VERSION:", os.Getenv("OPS_VERSION"))
	fmt.Println("OPS_BRANCH:", os.Getenv("OPS_BRANCH"))
	fmt.Println("OPS_BIN:", os.Getenv("OPS_BIN"))
	fmt.Println("OPS_TMP:", os.Getenv("OPS_TMP"))
	fmt.Println("OPS_HOME:", os.Getenv("OPS_HOME"))
	fmt.Println("OPS_ROOT:", os.Getenv("OPS_ROOT"))
	fmt.Println("OPS_REPO:", os.Getenv("OPS_REPO"))
	fmt.Println("OPS_PWD:", os.Getenv("OPS_PWD"))
	fmt.Println("OPS_OLARIS:", os.Getenv("OPS_OLARIS"))
	fmt.Println("OPS_ROOT_PLUGIN:", os.Getenv("OPS_ROOT_PLUGIN"))
	//fmt.Println("OPS_TOOLS:", os.Getenv("OPS_TOOLS"))
	//fmt.Println("OPS_COREUTILS:", os.Getenv("OPS_COREUTILS"))
}

func banner() {
	fmt.Println("Welcome to ops, the all-mighty, extensibile apache OPenServerless CLI Tool.")
	fmt.Println("General syntax:")
	fmt.Println("   ops -<tool>  <args>...")
	fmt.Println("   ops <task>   <args>...")
	fmt.Println("Essential tools (chek -h and -t for more):")
	fmt.Println("-h | -help    list tools      (embedded tools, prefixed by '-', no download)")
	fmt.Println("-v | -version current version (always mention this when asking for help)")
	fmt.Println("-t | -tasks   list tasks      (show top level tasks, download if needed)")
	fmt.Println("-i | -info    CLI infos       (show the cli environment)")
	fmt.Println("-u | -update  download latest (get the latest tasks and prerequisites)")
	fmt.Println("-c | -config  manage config   (openserverless server configuration)")
	fmt.Println("-l | -login   access system   (required to access openserverless)")
	fmt.Println("-reset        clean downloads (if nothing works, try this)")
	fmt.Println()
}

// simple tools provide info and exit
func executeToolsNoDownloadAndExit(args []string) {
	if len(args) < 2 {
		banner()
		os.Exit(0)
	}
	// if we have at least one arg
	switch args[1] {
	case "-v", "-version":
		fmt.Println(OpsVersion)
		os.Exit(0)
	case "-h", "-help":
		banner()
		tools.Help(mainTools)
		os.Exit(0)
	case "-reset":
		home := os.Getenv("OPS_HOME")
		if home == "" {
			log.Fatal("cannot determine the ops home dir")
			os.Exit(1)
		}
		info, err := os.Stat(home)
		if os.IsNotExist(err) {
			fmt.Printf("%s does not exists - nothing to to do\n", home)
			os.Exit(1)
		}
		if err != nil {
			log.Fatal("error in reading the ops home dir", err.Error())
		}
		if !info.IsDir() {
			log.Fatal("cannot reset, not a directory", home)
		}
		if len(args) == 2 || (len(args) > 2 && args[2] != "force") {
			if !confirm(fmt.Sprintf("I am going to remove the subfolder %s (use force to skip this question).\nAre you sure [yes/no]:", home)) {
				log.Fatal("reset aborted")
			}
		} else {
			fmt.Println("removing without asking for confirmation as requested...")
		}
		err = os.RemoveAll(home)
		if err != nil {
			log.Fatal("ops reset error:", err.Error())
		}
		fmt.Println("ops -reset complete - execute ops -update to reload")
		os.Exit(0)

	default:
		return
	}
}

var mainTools = []string{
	"task", "info", "update", "login", "config",
	"retry", "plugin", "reset", "serve",
}

// CLI: ops -<cmd> <args>...
func executeTools(args []string, opsHome string) int {
	trace("TOOL:", args)
	cmd := args[0]
	switch cmd {
	case "", "task":
		args = args[1:]
		exitCode, err := Task(args...)
		if err != nil {
			log.Println(err)
		}
		return exitCode

	case "i", "info":
		info()
		return 0

	case "u", "update":
		// ok no up, nor down, let's download it
		dir, err := pullTasks(true, true)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		if err := setOpsOlarisHash(dir); err != nil {
			log.Fatal("unable to set OPS_OLARIS...", err.Error())
		}
		return 0

	case "l", "login":
		args[0] = "-login"
		os.Args = args
		loginResult, err := auth.LoginCmd()
		if err != nil {
			log.Fatalf("error: %s", err.Error())
		}

		if loginResult == nil {
			return 1
		}

		fmt.Println("Successfully logged in as " + loginResult.Login + ".")
		if err := wskPropertySet(loginResult.ApiHost, loginResult.Auth); err != nil {
			log.Fatalf("error: %s", err.Error())
		}
		fmt.Println("OpenServerless host and auth set successfully. You are now ready to use ops!")
		return 0

	case "c", "config":
		args[0] = "-config"
		os.Args = args
		opsRootPath := joinpath(getRootDirOrExit(), OPSROOT)
		configPath := joinpath(opsHome, CONFIGFILE)
		configMap, err := buildConfigMap(opsRootPath, configPath)
		if err != nil {
			log.Fatalf("error: %s", err.Error())
		}

		if err := config.ConfigTool(*configMap); err != nil {
			log.Fatalf("error: %s", err.Error())
		}
		return 0

	case "retry":
		args[0] = "-retry"
		if err := tools.ExpBackoffRetry(args); err != nil {
			log.Fatalf("error: %s", err.Error())
		}
		return 0

	case "plugin":
		args[0] = "-plugin"
		os.Args = args
		if err := pluginTool(); err != nil {
			log.Fatalf("error: %s", err.Error())
		}
		return 0

	case "serve":
		args[0] = "-serve"
		opsRootDir := getRootDirOrExit()
		if err := Serve(opsRootDir, args); err != nil {
			log.Fatalf("error: %v", err)
		}
		return 0

	default:
		// check if it is an embedded to and invoke it
		if tools.IsTool(cmd) {
			code, err := tools.RunTool(cmd, args[1:])
			if err != nil {
				log.Print(err.Error())
			}
			return code
		}
		// no embeded tool found
		warn("unknown tool", "-"+cmd)
		return 1
	}
	//return 0 // unreachable - or it should be!
}

func Main() {
	var err error

	// disable log prefix
	log.SetFlags(0)

	// set default runtime.json
	if os.Getenv("WSK_RUNTIMES_JSON") == "" {
		os.Setenv("WSK_RUNTIMES_JSON", WSK_RUNTIMES_JSON)
		trace("WSK_RUNTIMES_JSON len=", len(WSK_RUNTIMES_JSON))
	}

	// set runtime version as environment variable
	if os.Getenv("OPS_VERSION") != "" {
		OpsVersion = os.Getenv("OPS_VERSION")
	} else {
		OpsVersion = strings.TrimSpace(OpsVersion)
		os.Setenv("OPS_VERSION", OpsVersion)
	}

	// setup OPS_CMD
	// CLI: ops ...
	me := os.Args[0]
	if strings.Contains("ops ops.exe", filepath.Base(me)) {
		_, err = setupCmd(me)
		if err != nil {
			log.Fatalf("cannot setup cmd: %s", err.Error())
		}
	}
	os.Setenv("OPS", me)

	// setup home
	opsHome := os.Getenv("OPS_HOME")
	if opsHome == "" {
		opsHome, err = homedir.Expand("~/.ops")
	}
	if err != nil {
		log.Fatalf("cannot setup home: %s", err.Error())
	}
	os.Setenv("OPS_HOME", opsHome)
	trace("OPS_HOME", opsHome)

	// add ~/.ops/<os>-<arch>/bin to the path at the beginning
	err = setupBinPath()
	if err != nil {
		log.Fatalf("cannot setup PATH: %s", err.Error())
	}

	// ensure there is ~/.ops/tmp
	err = setupTmp()
	if err != nil {
		log.Fatalf("cannot setup OPS_TMP: %s", err.Error())
	}

	//  setup the OPS_PWD variable
	err = setOpsPwdEnv()
	if err != nil {
		log.Fatalf("cannot setup OPS_PWD: %s", err.Error())
	}

	// setup the envvar for the embedded tools
	os.Setenv("OPS_TOOLS", strings.Join(tools.MergeToolsList(mainTools), " "))

	// CLI: ops -v | --version | -h | --help | -reset
	// preliminanre processing not requiring to  downloading anything
	executeToolsNoDownloadAndExit(os.Args)

	// in case args[1] is a wsk wrapper command invoke it and exit
	// CLI: ops action ... (wsk wrapper)
	if len(os.Args) > 1 {
		if expand, ok := IsWskWrapperCommand(os.Args[1]); ok {
			debug("wsk wrapper cmd", expand)
			rest := os.Args[2:]
			debug("extracted args", rest)

			// if "invoke" is in the command, parse all a=b into -p a b
			if (len(expand) > 2 && expand[2] == "invoke") || slices.Contains(rest, "invoke") {
				rest = parseInvokeArgs(rest)
			}

			if err := tools.Wsk(expand, rest...); err != nil {
				log.Fatalf("error: %s", err.Error())
			}
			os.Exit(0)
		}
	}

	// OPS_REPO && OPS_ROOT_PLUGIN
	getOpsRepo()
	setOpsRootPluginEnv()
	// Check if olaris exists. If not, download tasks
	olarisDir, err := getOpsRoot()
	if err != nil {
		olarisDir := joinpath(joinpath(opsHome, getOpsBranch()), "olaris")
		if !isDir(olarisDir) {
			log.Println("Welcome to ops! Setting up...")
			olarisDir, err = pullTasks(true, true)
			if err != nil {
				log.Fatalf("cannot locate or download OPS_ROOT: %s", err.Error())
			}
			// if just updated, do not repeat
			if len(os.Args) > 1 && os.Args[1] == "-update" {
				os.Exit(0)
			}
		} else {
			// check if olaris was recently updated
			checkUpdated(opsHome, 24*time.Hour)
		}
	}
	if err = setOpsOlarisHash(olarisDir); err != nil {
		os.Setenv("OPS_OLARIS", "<local>")
	}

	// set the enviroment variables from the config
	opsRootDir := getRootDirOrExit()
	debug("opsRootDir", opsRootDir)
	err = setAllConfigEnvVars(opsRootDir, opsHome)
	if err != nil {
		log.Fatalf("cannot apply env vars from configs: %s", err.Error())
	}

	// preflight checks - we need at least ssh curl to proceed
	if err := preflightChecks(); err != nil {
		log.Fatalf("failed preflight check: %s", err.Error())
	}

	// first argument with prefix "-" is considered an embedded tool
	// using "-" or "--" or "-task" invokes the embedded task
	// CLI: ops -<tool> (embedded tool)
	args := os.Args
	if len(args) > 1 && len(args[1]) > 0 && args[1][0] == '-' {
		cmd := args[1][1:]
		if cmd == "t" || cmd == "tasks" {
			banner()
			// remove -t to show tasks and continue to execute and list top level tasks
			args = args[1:]
		} else {
			// execute the embeded tool and exit
			fullargs := append([]string{cmd}, args[2:]...)
			exitCode := executeTools(fullargs, opsHome)
			os.Exit(exitCode)
		}
	}

	if err := runOps(opsRootDir, args); err != nil {
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
	dir, err := getOpsRoot()
	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}
	return dir
}

func setAllConfigEnvVars(opsRootDir string, configDir string) error {
	trace("setting all config env vars")

	configMap, err := buildConfigMap(joinpath(opsRootDir, OPSROOT), joinpath(configDir, CONFIGFILE))
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

func runOps(baseDir string, args []string) error {
	err := Ops(baseDir, args[1:])
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
		if err := Ops(plgDir, args[2:]); err != nil {
			log.Fatalf("error: %s", err.Error())
		}
		return nil
	}

	return err
}

func setOpsPwdEnv() error {
	if os.Getenv("OPS_PWD") == "" {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		//nolint:errcheck
		os.Setenv("OPS_PWD", dir)
	}
	trace("set OPS_PWD", os.Getenv("OPS_PWD"))
	return nil
}

func buildConfigMap(opsRootPath string, configPath string) (*config.ConfigMap, error) {
	plgOpsRootMap, err := GetOpsRootPlugins()
	if err != nil {
		return nil, err
	}

	configMap, err := config.NewConfigMapBuilder().
		WithOpsRoot(opsRootPath).
		WithConfigJson(configPath).
		WithPluginOpsRoots(plgOpsRootMap).
		Build()

	if err != nil {
		return nil, err
	}

	return &configMap, nil
}

var wskWrapperCommands = map[string][]string{
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

func IsWskWrapperCommand(name string) ([]string, bool) {
	cmd, ok := wskWrapperCommands[name]
	return cmd, ok
}
