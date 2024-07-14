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
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	docopt "github.com/docopt/docopt-go"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"

	envsubst "github.com/nuvolaris/envsubst/cmd/envsubstmain"
)

type TaskNotFoundErr struct {
	input string
}

func (e *TaskNotFoundErr) Error() string {
	return fmt.Sprintf("no command named %s found", e.input)
}

func help() error {
	if os.Getenv("NUV_NO_NUVOPTS") == "" && exists(".", NUVOPTS) {
		os.Args = []string{"envsubst", "-no-unset", "-i", NUVOPTS}
		return envsubst.EnvsubstMain()
	}
	// In case of syntax error, Task will return an error
	list := "-l"
	if os.Getenv("NUV_NO_NUVOPTS") != "" {
		list = "--list-all"
	}
	_, err := Task("-t", NUVFILE, list)

	return err
}

// parseArgs parse the arguments acording the docopt
// it returns a sequence suitable to be feed as arguments for task.
// note that it will change hyphens for flags ('-c', '--count') to '_' ('_c' '__count')
// and '<' and '>' for parameters '_' (<hosts> => _hosts_)
// boolean are "true" or "false" and arrays in the form ('first' 'second')
// suitable to be used as arrays
// Examples:
// if "Usage: nettool ping [--count=<max>] <hosts>..."
// with "ping --count=3 google apple" returns
// ping=true _count=3 _hosts_=('google' 'apple')
func parseArgs(usage string, args []string) []string {
	res := []string{}
	// parse args
	parser := docopt.Parser{}
	opts, err := parser.ParseArgs(usage, args, NuvVersion)
	if err != nil {
		warn(err)
		return res
	}
	for k, v := range opts {
		kk := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(k, "-", "_"), "<", "_"), ">", "_")
		vv := ""
		//fmt.Println(v, reflect.TypeOf(v))
		switch o := v.(type) {
		case bool:
			vv = "false"
			if o {
				vv = "true"
			}
		case string:
			vv = o
		case []string:
			a := []string{}
			for _, i := range o {
				a = append(a, fmt.Sprintf("'%v'", i))
			}
			vv = "(" + strings.Join(a, " ") + ")"
		case nil:
			vv = ""
		}
		res = append(res, fmt.Sprintf("%s=%s", kk, vv))
	}
	sort.Strings(res)
	return res
}

// setupTmp sets up a tmp folder
func setupTmp() {
	// setup NUV_TMP
	var err error
	tmp := os.Getenv("NUV_TMP")
	if tmp == "" {
		tmp, err = homedir.Expand("~/.nuv/tmp")
		if err == nil {
			//nolint:errcheck
			os.Setenv("NUV_TMP", tmp)
		}
	}
	if err == nil {
		err = os.MkdirAll(tmp, 0755)
	}
	if err != nil {
		warn("cannot create tmp dir", err)
		os.Exit(1)
	}
}

// load saved args in files names _*_ in current directory
func loadSavedArgs() []string {
	res := []string{}
	files, err := os.ReadDir(".")
	if err != nil {
		return res
	}
	r := regexp.MustCompile(`^_.+_$`) // regex to match file names that start and end with '_'
	for _, f := range files {
		if !f.IsDir() && r.MatchString(f.Name()) {
			debug("reading vars from " + f.Name())
			file, err := os.Open(f.Name())
			if err != nil {
				warn("cannot read " + f.Name())
				continue
			}
			scanner := bufio.NewScanner(file)
			r := regexp.MustCompile(`^[a-zA-Z0-9]+=`) // regex to match lines that start with an alphanumeric sequence followed by '='
			for scanner.Scan() {
				line := scanner.Text()
				if r.MatchString(line) {
					debug("found var " + line)
					res = append(res, line)
				}
			}
			err = scanner.Err()
			//nolint:errcheck
			file.Close()
			if err != nil {
				warn(err)
				continue
			}
		}
	}
	return res
}

// Nuv parses args moving into the folder corresponding to args
// then parses them with docopts and invokes the task
func Nuv(base string, args []string) error {
	trace("Nuv run in", base, "with", args)
	// go down using args as subcommands
	err := os.Chdir(base)
	debug("Nuv chdir", base)

	if err != nil {
		return err
	}
	rest := args

	isSubCmd := false
	for _, task := range args {
		trace("task name", task)

		// skip flags
		if strings.HasPrefix(task, "-") {
			continue
		}

		// try to correct name if it's not a flag
		pwd, _ := os.Getwd()
		taskName, err := validateTaskName(pwd, task)
		if err != nil {
			return err
		}
		// if valid, check if it's a folder and move to it
		if isDir(taskName) && exists(taskName, NUVFILE) {
			if err := os.Chdir(taskName); err != nil {
				return err
			}
			//remove it from the args
			rest = rest[1:]
			isSubCmd = true
		} else {
			// stop when non folder reached
			//substitute it with the validated task name
			if len(rest) > 0 {
				rest[0] = taskName
			}
			break
		}
	}

	if len(rest) == 0 || rest[0] == "help" {
		trace("print help")
		err := help()
		if !isSubCmd {
			fmt.Println()
			return printPluginsHelp()
		}
		return err
	}

	// load saved args
	savedArgs := loadSavedArgs()

	// parsed args
	if os.Getenv("NUV_NO_NUVOPTS") == "" && exists(".", NUVOPTS) {
		trace("PREPARSE:", rest)
		parsedArgs := parseArgs(readfile(NUVOPTS), rest)
		prefix := []string{"-t", NUVFILE}
		if len(rest) > 0 && rest[0][0] != '-' {
			prefix = append(prefix, rest[0])
		}

		parsedArgs = append(savedArgs, parsedArgs...)
		parsedArgs = append(prefix, parsedArgs...)
		extra := os.Getenv("EXTRA")
		if extra != "" {
			trace("EXTRA:", extra)
			parsedArgs = append(parsedArgs, strings.Split(extra, " ")...)
		}
		trace("POSTPARSE:", parsedArgs)
		_, err := Task(parsedArgs...)
		return err
	}

	mainTask := rest[0]

	// unparsed args - separate variable assignments from extra args
	pre := []string{"-t", NUVFILE, mainTask}
	pre = append(pre, savedArgs...)
	post := []string{"--"}
	args1 := rest[1:]
	extra := os.Getenv("EXTRA")
	if extra != "" {
		trace("EXTRA:", extra)
		args1 = append(args1, strings.Split(extra, " ")...)
	}
	for _, s := range args1 {
		if strings.Contains(s, "=") {
			pre = append(pre, s)
		} else {
			post = append(post, s)
		}
	}
	taskArgs := append(pre, post...)

	debug("task args: ", taskArgs)
	_, err = Task(taskArgs...)
	return err
}

// validateTaskName does the following:
// 1. Check that the given task name is found in the nuvfile.yaml and return it
// 2. If not found, check if the input is a prefix of any task name, if it is for only one return the proper task name
// 3. If the prefix is valid for more than one task, return an error
// 4. If the prefix is not valid for any task, return an error
func validateTaskName(dir string, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("command name is empty")
	}

	candidates := []string{}
	tasks := getTaskNamesList(dir)
	if !slices.Contains(tasks, "help") {
		tasks = append(tasks, "help")
	}
	for _, t := range tasks {
		if t == name {
			return name, nil
		}
		if strings.HasPrefix(t, name) {
			candidates = append(candidates, t)
		}
	}

	if len(candidates) == 0 {
		return "", &TaskNotFoundErr{input: name}
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	return "", fmt.Errorf("ambiguous command: %s. Possible matches: %v", name, candidates)
}

// obtains the task names from the nuvfile.yaml inside the given directory
func getTaskNamesList(dir string) []string {
	m := make(map[interface{}]interface{})
	var taskNames []string
	if exists(dir, NUVFILE) {
		dat, err := os.ReadFile(joinpath(dir, NUVFILE))
		if err != nil {
			return make([]string, 0)
		}

		err = yaml.Unmarshal(dat, &m)
		if err != nil {
			warn("error reading nuvfile.yml")
			return make([]string, 0)
		}
		tasksMap, ok := m["tasks"].(map[string]interface{})
		if !ok {
			// warn("error checking task list, perhaps no tasks defined?")
			return make([]string, 0)
		}

		for k := range tasksMap {
			taskNames = append(taskNames, k)
		}

	}

	// for each subfolder, check if it has a nuvfile.yaml
	// if it does, add it to the list of tasks

	// get subfolders
	subfolders, err := os.ReadDir(dir)
	if err != nil {
		warn("error reading subfolders of", dir)
		return taskNames
	}

	for _, f := range subfolders {
		if f.IsDir() {
			subfolder := joinpath(dir, f.Name())
			if exists(subfolder, NUVFILE) {
				// check if not contained
				name := f.Name()
				if !slices.Contains(taskNames, name) {
					taskNames = append(taskNames, name)
				}
			}
		}
	}

	return taskNames
}
