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

package tools

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/radovskyb/watcher"
)

type deployCtx struct {
	path               string
	dryRun             bool
	packageCmdExecuted map[string]bool
}

func DeployTool() error {
	flag := flag.NewFlagSet("deploy", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Println(`Command to deploy Nuvolaris projects. Must be given the path of a valid project (containing 'packages').

Usage:
	nuv -deploy [-s <string> | --single <string>] [-d | --dry-run] <path> 
	nuv -deploy [-w | --watch] [-d | --dry-run] <path> 

Options:
	-s, --single <string>     Deploy a single action with the given path, either a single file or a directory.
	-w, --watch     Watch for changes and deploy automatically.
	-d, --dry-run   Do not deploy, just print the deployment plan.
	-h, --help      Show this help message.`)
	}

	var watchFlag bool
	var singleFlag string
	var helpFlag bool
	var dryRunFlag bool
	flag.StringVar(&singleFlag, "s", "", "Deploy a single action with the given path, either a single file or a directory.")
	flag.StringVar(&singleFlag, "single", "", "Deploy a single action with the given path, either a single file or a directory.")
	flag.BoolVar(&helpFlag, "h", false, "Show this help message")
	flag.BoolVar(&helpFlag, "help", false, "Show this help message")
	flag.BoolVar(&dryRunFlag, "d", false, "Do not deploy, just print the deployment plan.")
	flag.BoolVar(&dryRunFlag, "dry-run", false, "Do not deploy, just print the deployment plan.")
	flag.BoolVar(&watchFlag, "w", false, "Watch for changes and deploy automatically.")
	flag.BoolVar(&watchFlag, "watch", false, "Watch for changes and deploy automatically.")

	// Parse command line flags
	err := flag.Parse(os.Args[2:])
	if err != nil {
		return err
	}

	if flag.NArg() != 1 || helpFlag {
		flag.Usage()
		return nil
	}

	path, err := homedir.Expand(flag.Arg(0))
	if err != nil {
		return err
	}

	var absPath string
	if filepath.IsLocal(path) {
		absPath = filepath.Join(os.Getenv("NUV_PWD"), path)
	} else {
		absPath = path
	}

	os.Setenv("NUV_PWD", absPath)

	ctx := deployCtx{
		path:               os.Getenv("NUV_PWD"),
		packageCmdExecuted: make(map[string]bool),
		dryRun:             dryRunFlag,
	}

	// if there is not "packages" folder from where deploy was called, abort
	if !exists(filepath.Join(ctx.path, "packages")) {
		return fmt.Errorf("no 'packages' folder found in the current directory")
	}

	if singleFlag != "" {
		action := singleFlag
		if !strings.HasPrefix(action, "packages") {
			action = filepath.Join("packages", action)
			if !exists(filepath.Join(ctx.path, action)) {
				return fmt.Errorf("action %s not found: must be either a file or a directory under packages", action)
			}
		}
		return deploy(ctx, action)
	}

	if watchFlag {
		err = scan(ctx)
		if err != nil {
			return err
		}

		err = watch(ctx)
		return err
	}

	err = scan(ctx)
	if err != nil {
		return err
	}

	return nil
}

var supportedMains = []string{"__main__.py", "index.js", "main.js", "main.go"}

func deploy(ctx deployCtx, actionPath string) error {
	fullPath := filepath.Join(ctx.path, actionPath)
	if !exists(fullPath) {
		return fmt.Errorf("action %s not found: must be either a file or a directory", actionPath)
	}
	log.Println("***", filepath.Base(actionPath))

	action, err := checkActionDir(ctx.path, actionPath)
	if err != nil {
		return err
	}

	sp := splitPath(action)
	if len(sp) > 3 {
		action, err = buildAction(ctx, sp[1], sp[2])
		if err != nil {
			return err
		}
	}

	return deployAction(ctx, action)
}

func checkActionDir(rootPath string, actionPath string) (string, error) {
	fullPath := filepath.Join(rootPath, actionPath)
	isActionDir := false
	isActionDirSupported := false
	if fileInfo, err := os.Stat(fullPath); err == nil && fileInfo.Mode().IsDir() {
		isActionDir = true
		for _, start := range supportedMains {
			sub := filepath.Join(actionPath, start)
			if exists(filepath.Join(rootPath, sub)) {
				actionPath = sub
				isActionDirSupported = true
				break
			}
		}
	}
	if isActionDir && !isActionDirSupported {
		return "", fmt.Errorf("action %s is a directory but does not contain a supported main file", actionPath)
	}
	return actionPath, nil
}

func deployAction(ctx deployCtx, artifact string) error {
	sp := splitPath(artifact)
	nameType := strings.Split(sp[len(sp)-1], ".")
	name := nameType[0]
	typ := nameType[1]
	packageName := filepath.Base(filepath.Dir(artifact))

	deployPackage(ctx, packageName)

	var toInspect []string
	if typ == "zip" {
		base := filepath.Join(ctx.path, artifact[:len(artifact)-4])

		// TODO: add support for other languages
		toInspect = []string{filepath.Join(base, "/__main__.py"), filepath.Join(base, "/index.js")}
	} else {
		toInspect = []string{artifact}
	}

	args := extractArgs(toInspect)

	action := packageName + "/" + name // the action name, it's not a file path
	if packageName == "packages" {
		action = name
	}

	cmd := []string{"nuv", "action", "update", action, artifact}
	cmd = append(cmd, args...)
	if !ctx.dryRun {

		execCmd := exec.Command("nuv", cmd[1:]...)
		execCmd.Stderr = os.Stderr
		execCmd.Stdout = os.Stdout
		err := execCmd.Run()
		if err != nil {
			return fmt.Errorf("error deploying action %s: %v", action, err)
		}
	} else {
		log.Println("Would run:", strings.Join(cmd, " "))
	}

	return nil
}

func deployPackage(ctx deployCtx, pkg string) {
	// package args
	ppath := filepath.Join(ctx.path, "packages", pkg+".args")
	pargs := extractArgs([]string{ppath})
	cmd := []string{"nuv", "package", "update", pkg}

	cmd = append(cmd, pargs...)

	cmdToSave := strings.Join(cmd, " ")
	if _, ok := ctx.packageCmdExecuted[cmdToSave]; !ok {
		if !ctx.dryRun {
			execCmd := exec.Command("nuv", cmd[1:]...)
			execCmd.Stderr = os.Stderr
			execCmd.Stdout = os.Stdout
			err := execCmd.Run()
			if err != nil {
				log.Println("Error deploying package", pkg, err)
			}
		} else {
			log.Println("Would run:", strings.Join(cmd, " "))
		}

		ctx.packageCmdExecuted[cmdToSave] = true
	}
}

func extractArgs(files []string) []string {
	res := []string{}
	for _, file := range files {
		if exists(file) {
			f, err := os.Open(file)
			if err != nil {
				log.Println("Error opening file", file, err)
				continue
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "#-") {
					res = append(res, strings.TrimSpace(line[1:]))
				}
				if strings.HasPrefix(line, "//-") {
					res = append(res, strings.TrimSpace(line[2:]))
				}
			}

			if err := scanner.Err(); err != nil {
				log.Println("Error reading file", file, err)
			}
		}
	}
	// for each string, split by space and add to args
	args := []string{}
	for _, arg := range res {
		args = append(args, strings.Split(arg, " ")...)
	}
	return args
}

func buildAction(ctx deployCtx, pkg string, action string) (string, error) {
	if !ctx.dryRun {
		cmd := exec.Command("nuv", "ide", "util", "action", fmt.Sprintf("A=%s/%s", pkg, action))
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return "", fmt.Errorf("error building action %s/%s: %v", pkg, action, err)
		}
	} else {
		log.Println("Would run: nuv ide util action A=" + pkg + "/" + action)
	}
	return fmt.Sprintf("packages/%s/%s.zip", pkg, action), nil
}

func exists(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	dir, last := filepath.Split(filepath.Clean(path))
	if dir == "." || dir == "/" || dir == "" {
		return []string{last}
	}
	return append(splitPath(dir), last)
}

// region: scan

// scan scans the packages and deployments.
func scan(ctx deployCtx) error {
	wd, _ := os.Getwd()
	_ = os.Chdir(ctx.path)
	defer func() {
		_ = os.Chdir(wd)
	}()

	// First look for requirements.txt and build the venv (add in set)
	deployments := make(map[string]bool)
	packages := make(map[string]bool)

	log.Println(">>> Scan:")
	pyGlob := filepath.Join("packages", "*", "*", "requirements.txt")
	jsGlob := filepath.Join("packages", "*", "*", "package.json")
	reqs, _ := filepath.Glob(pyGlob)
	reqs2, _ := filepath.Glob(jsGlob)
	reqs = append(reqs, reqs2...)

	for _, req := range reqs {
		log.Println(">", req)
		sp := splitPath(req)
		action, err := buildZip(ctx, sp[1], sp[2])
		if err != nil {
			return fmt.Errorf("error building zip for %s/%s: %v", sp[1], sp[2], err)
		}
		deployments[action] = true
		packages[sp[1]] = true
	}

	pyMainGlob := filepath.Join("packages", "*", "*", "__main__.py")
	jsMainGlob := filepath.Join("packages", "*", "*", "index.js")
	mains, _ := filepath.Glob(pyMainGlob)
	pymains, _ := filepath.Glob(jsMainGlob)
	mains = append(mains, pymains...)
	for _, main := range mains {
		log.Println(">", main)
		sp := splitPath(main)
		action, err := buildAction(ctx, sp[1], sp[2])
		if err != nil {
			return fmt.Errorf("error building action %s/%s: %v", sp[1], sp[2], err)
		}
		deployments[action] = true
		packages[sp[1]] = true
	}

	pySinglesGlob := filepath.Join("packages", "*", "*.py")
	jsSinglesGlob := filepath.Join("packages", "*", "*.js")
	singles, _ := filepath.Glob(pySinglesGlob)
	jsSingles, _ := filepath.Glob(jsSinglesGlob)
	singles = append(singles, jsSingles...)
	for _, single := range singles {
		log.Println(">", single)
		sp := splitPath(single)
		deployments[single] = true
		packages[sp[1]] = true
	}

	if len(deployments) == 0 && len(packages) == 0 {
		log.Println("Nothing to do")
		return nil
	}

	log.Println(">>> Deploying:")

	for p := range packages {
		log.Println("%", p)
		deployPackage(ctx, p)
	}

	for a := range deployments {
		log.Println("^", a)
		err := deployAction(ctx, a)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildZip(ctx deployCtx, pkg string, action string) (string, error) {
	if !ctx.dryRun {
		cmd := exec.Command("nuv", "ide", "util", "zip", fmt.Sprintf("A=%s/%s", pkg, action))
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return "", fmt.Errorf("error building zip %s/%s: %v", pkg, action, err)
		}
	} else {
		log.Println("Would run: nuv ide util zip A=" + pkg + "/" + action)
	}

	return fmt.Sprintf("packages/%s/%s.zip", pkg, action), nil
}

// endregion

// region: watch

var SKIPDIR = []string{"virtualenv", "node_modules", "__pycache__"}

func watch(ctx deployCtx) error {
	log.Println(">>> Watching:", filepath.Join(ctx.path, "packages"))

	w := watcher.New()

	err := w.AddRecursive(filepath.Join(ctx.path, "packages"))
	if err != nil {
		return err
	}

	w.IgnoreHiddenFiles(true)
	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	// w.SetMaxEvents(1)

	watcherEvent := &watcherEvent{
		lastModified: make(map[string]time.Time),
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				watcherEvent.changeHandler(ctx, event)
			case err := <-w.Error:
				log.Println(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		return err
	}

	return nil
}

type watcherEvent struct {
	lastModified map[string]time.Time
}

func (w *watcherEvent) changeHandler(ctx deployCtx, event watcher.Event) {
	if event.Op != watcher.Write {
		return // only handle write events
	}
	if event.IsDir() {
		return // skip directories
	}

	src := event.Path
	if !exists(src) {
		return // no missing files
	}

	for _, dir := range splitPath(src)[:len(splitPath(src))-1] {
		for _, skip := range SKIPDIR { // no generated directories
			if dir == skip {
				return
			}
		}
	}

	if strings.HasSuffix(src, ".zip") {
		return // no generated files
	}

	// cache last modified to do only once (mod time of file)
	f, _ := os.Stat(src)
	cur := f.ModTime()
	lastMod, ok := w.lastModified[src]
	if ok && lastMod.Equal(cur) {
		return
	}
	w.lastModified[src] = cur

	log.Println("Changed:", src)
	// remove the ctx.path prefix
	err := deploy(ctx, strings.TrimPrefix(src, ctx.path+"/"))
	if err != nil {
		log.Println("Error deploying action", src, err)
	}
}

// endregion
