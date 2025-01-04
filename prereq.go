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

package openserverless

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apache/openserverless-cli/tools"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

var PrereqSeenMap = map[string]string{}

// execute prereq task
func execPrereqTask(bindir string, name string) error {
	me, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{
		"-task",
		"-d", bindir,
		"-t", PREREQ,
		name,
	}
	if taskDryRun {
		fmt.Printf("invoking prereq for %s\n", name)
		return nil
	}
	trace("Exec:", me, args)
	err = exec.Command(me, args...).Run()
	if err != nil {
		return err
	}
	return nil
}

// load prerequisites in current dir
func loadPrereq(dir string) (tasks []string, versions []string, err error) {
	err = nil
	tasks = []string{}
	versions = []string{}

	if !exists(dir, PREREQ) {
		return
	}
	trace("found prereq.yml in ", dir)

	data, err := os.ReadFile(joinpath(dir, PREREQ))
	if err != nil {
		return
	}

	/*
			You have tasks = []strings and versions = []strings
			and dat a string with a YAML  in format:
			```yaml
			something:
			tasks:
			   task:
			      vars:
				     VERSION: "123"
			   anothertask:
		    ```
			I want to parse the file, find the entries under `tasks` vith a var with  VERSION
			and append, in order, to tasks the name of the task
			and to version the version found
			I have to skip the entries without a version and everyhing not in tasks
			I wnato just the plain code no procedures
	*/
	var root yaml.Node
	if err = yaml.Unmarshal([]byte(data), &root); err != nil {
		return
	}

	for i := 0; i < len(root.Content[0].Content); i += 2 {
		if root.Content[0].Content[i].Value == "tasks" {
			tasksNode := root.Content[0].Content[i+1]
			for j := 0; j < len(tasksNode.Content); j += 2 {
				taskName := tasksNode.Content[j].Value
				taskVars := tasksNode.Content[j+1]
				for k := 0; k < len(taskVars.Content); k += 2 {
					if taskVars.Content[k].Value == "vars" {
						varsNode := taskVars.Content[k+1]
						for l := 0; l < len(varsNode.Content); l += 2 {
							if varsNode.Content[l].Value == "VERSION" {
								version := varsNode.Content[l+1].Value
								tasks = append(tasks, taskName)
								versions = append(versions, version)
							}
						}
					}
				}
			}
		}
	}
	return
}

func binDir() (string, error) {
	var err error
	bindir := os.Getenv("OPS_BIN")
	if bindir == "" {
		bindir, err = homedir.Expand(fmt.Sprintf("~/.ops/%s-%s/bin", tools.GetOS(), tools.GetARCH()))
		if err != nil {
			return "", err
		}
	}
	os.Setenv("OPS_BIN", bindir)
	return bindir, nil
}

func addExeExt(name string) string {
	if tools.GetOS() == "windows" {
		return name + ".exe"
	}
	return name
}

// ensure there is a bindir for downloading prerequisites
// read it from OPS_BIN and create it
// otherwise setup one in ~/ops/<os>-<arch>/bin
// and sets OPS_BIN
func EnsureBindir() (string, error) {
	var err error = nil
	bindir, err := binDir()
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(bindir, 0755)
	if err != nil {
		return "", err
	}
	trace("bindir", bindir)
	return bindir, nil
}

// create a mark of current version touching <name>-<version> and remove all the other files starting with <name>-
func touchAndClean(dir string, name string, version string) error {

	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file starts with the prefix
		if !info.IsDir() && strings.HasPrefix(info.Name(), name+"-") {
			trace("Removing file:", path)
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	err = touch(dir, name+"-"+version)
	if err != nil {
		return err
	}
	return nil

}

// download a prerequisite
func downloadPrereq(name string, version string) error {

	// names and version
	// xname = executeable name
	// vname = versioned executable name
	xname := addExeExt(name)
	vname := xname + "-" + version

	// ensure bindir
	bindir, err := EnsureBindir()
	if err != nil {
		return err
	}

	// check if file and version exists
	trace("checking", vname, version)
	if exists(bindir, vname) {
		trace("already downloaded", vname)
		return nil
	}

	// checking different versions of the same file
	oldver, seen := PrereqSeenMap[name]
	if seen {
		if oldver == version {
			trace("same version again", vname)
			return nil
		}
		return fmt.Errorf("WARNING: %s prerequisite found twice with different versions!\nPrevious version: %s, ignoring %s", name, oldver, version)
	}
	PrereqSeenMap[name] = version

	if taskDryRun {
		fmt.Printf("downloading %s %s\n", name, version)
		touch(bindir, name)
	} else {
		fmt.Printf("ensuring prerequisite %s %s\n", name, version)
		execPrereqTask(bindir, name)
		// check if file and version exists

		if !exists(bindir, xname) {
			return fmt.Errorf("failed to download %s version %s", name, version)
		}
		// check if a file is zero length and remove in this case
		fileInfo, err := os.Stat(joinpath(bindir, xname))
		if err != nil {
			return fmt.Errorf("failed to download %s version %s", name, version)
		}
		if fileInfo.Size() == 0 {
			trace("removing the empty file ", xname)
			err := os.Remove(joinpath(bindir, xname))
			if err != nil {
				return fmt.Errorf("cannot remove empty %s ", xname)
			}
		}
	}
	return touchAndClean(bindir, xname, version)
}

// ensure prereq are satified looking at the prereq.yml
func ensurePrereq(root string) error {
	// skip prereq - useful for tests
	if os.Getenv("OPS_NO_PREREQ") != "" {
		return nil
	}
	err := os.Chdir(root)
	if err != nil {
		return err
	}
	trace("ensurePrereq in", root)
	tasks, versions, err := loadPrereq(root)
	if err != nil {
		return err
	}
	trace(tasks, versions, err)
	for i, task := range tasks {
		version := versions[i]
		trace("prereq", task, version)
		err = downloadPrereq(task, version)
		if err != nil {
			fmt.Printf("error in prereq %s: %v\n", task, err)
		}
	}
	return nil
}
