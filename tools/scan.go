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
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const actionsFolder = "actions"

// The scan tool visits a folder and creates an action plan to execute the given cmd (called cmdPlan).
// The cmd is the given nuv command to run, and the args is an array of arrays.
// Each entry is one execution instance, which holds the folder path (first arg) and files names (the rest).
// The plan is then executed by running the cmd once for each entry of args (so args.length times)
// in the form of `cmd args[i][0] args[i][1] ... args[i][n]` for each i in args.
type cmdPlan struct {
	par    bool
	dryRun bool
	cmd    []string
	args   [][]string
}

func (p *cmdPlan) apply() error {
	var ars []string
	if len(p.cmd) > 1 {
		ars = append(ars, p.cmd[1:]...)
	}

	if p.par {
		var wg sync.WaitGroup
		for _, args := range p.args {
			wg.Add(1)
			go func(args []string) {
				if p.dryRun {
					fmt.Println(strings.Join(append(p.cmd, args...), " "))
				} else {
					cmd := exec.Command(p.cmd[0], append(ars, args...)...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					if err := cmd.Run(); err != nil {
						fmt.Println(err)
					}
				}
				wg.Done()
			}(args)
		}
		wg.Wait()
		return nil
	}

	// run sequentially
	for _, args := range p.args {
		if p.dryRun {
			fmt.Println(strings.Join(append(p.cmd, args...), " "))
			continue
		}
		cmd := exec.Command(p.cmd[0], append(ars, args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (p *cmdPlan) setCmd(cmd []string) {
	p.cmd = cmd
}
func (p *cmdPlan) appendArg(args []string) {
	p.args = append(p.args, args)
}

func scanTool() error {
	flag := flag.NewFlagSet("scan", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Println(`Scan tool

The scan tool runs a command on an 'actions' folder and on all its sub-folders, recursively.
If the 'actions' folder does not exist in the path used, the tool stops.

In the 'actions' folder, the tool invokes the given nuv command with the following arguments:
  - the absolute path of the folder it is processing
  - all the files in the folder that matches the glob (none by default, use '*' to select all)

You can pass a glob pattern with the -g flag to filter the files used as input.

Usage: 
  nuv -scan [options] <nuv cmd> [args...]

Example:
  nuv -scan -g "*" nuv -js script.js

This results in running the script.js file on the ./actions folder and all subdirectories. 
For example, if $NUV_PWD/actions contains a subfolder called 'i-am-sub' with a file called 'bar.js',
the following commands are executed:

	- nuv -js script.js ./actions
	- nuv -js script.js ./actions/i-am-sub bar.js

Options:`)
		flag.PrintDefaults()
		fmt.Println()
	}

	var (
		helpFlag     bool
		dirFlag      string
		globFlag     string
		parallelFlag bool
		dryRunFlag   bool
	)

	// flag.BoolVar(&helpFlag, "h", false, "show help")
	flag.BoolVar(&helpFlag, "help", false, "show help")
	// flag.StringVar(&dirFlag, "d", getCurrentDir(), "directory to scan (default: $NUV_PWD)")
	flag.StringVar(&dirFlag, "dir", os.Getenv("NUV_PWD"), "directory to scan (default: $NUV_PWD)")
	// flag.StringVar(&globFlag, "g", "", "glob pattern to filter files (default: no files are passed to the nuv command)")
	flag.StringVar(&globFlag, "glob", "", "glob pattern between quotes to filter files (default: none -> no files are passed to the nuv command)")
	flag.BoolVar(&dryRunFlag, "dry-run", false, "print the plan without executing it")
	// flag.BoolVar(&parallelFlag, "p", false, "run in parallel (default: false)")
	flag.BoolVar(&parallelFlag, "par", false, "run in parallel (default: false)")

	if err := flag.Parse(os.Args[1:]); err != nil {
		return err
	}

	if helpFlag {
		flag.Usage()
		return nil
	}

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return errors.New("missing required nuv command")
	}

	p, err := filepath.Abs(dirFlag)
	if err != nil {
		return err
	}

	plan, err := buildCmdPlan(p, args, globFlag)
	if err != nil {
		return err
	}

	plan.par = parallelFlag
	plan.dryRun = dryRunFlag

	return plan.apply()
}

func buildCmdPlan(scanPath string, cmd []string, glob string) (*cmdPlan, error) {
	// check if actions folder exists
	if err := checkActionsFolderExists(scanPath); err != nil {
		return nil, err
	}

	dirs, err := getAllDirs(filepath.Join(scanPath, actionsFolder))
	if err != nil {
		return nil, err
	}

	plan := &cmdPlan{}
	plan.setCmd(cmd)

	for _, dir := range dirs {
		files := make([]string, 0)

		if len(glob) > 0 {
			fls, err := getAllFiles(dir)
			if err != nil {
				return nil, err
			}

			if len(fls) > 0 {
				files, err = filterFiles(fls, glob)
				if err != nil {
					return nil, err
				}
			}
		}

		plan.appendArg(append([]string{dir}, files...))
	}

	return plan, nil
}

func filterFiles(files []string, glob string) ([]string, error) {
	filtered := make([]string, 0)

	if len(glob) == 0 {
		return filtered, nil
	}

	for _, file := range files {
		matched, err := filepath.Match(glob, file)
		if err != nil {
			return filtered, err
		}

		if matched {
			filtered = append(filtered, file)
		}
	}

	return filtered, nil
}

func getAllDirs(rootDir string) ([]string, error) {
	var dirs []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			dirs = append(dirs, path)
		}

		return nil
	})

	return dirs, err
}

func getAllFiles(dir string) ([]string, error) {
	var files []string

	children, err := os.ReadDir(dir)
	if err != nil {
		return files, err
	}

	for _, child := range children {
		if !child.IsDir() {
			files = append(files, child.Name())
		}
	}

	return files, err
}

func checkActionsFolderExists(path string) error {
	info, err := os.Stat(filepath.Join(path, actionsFolder))
	if os.IsNotExist(err) {
		return fmt.Errorf("%s folder not found in %s", actionsFolder, path)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s in %s is not a folder", actionsFolder, path)
	}
	return nil
}
