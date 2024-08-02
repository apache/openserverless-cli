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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "embed"
)

// default files
const OPSFILE = "opsfile.yml"
const OPSROOT = "opsroot.json"
const DOCOPTS = "docopts.txt"
const PREREQ = "prereq.yml"
const CONFIGFILE = "config.json"

// repo where download tasks
const OPSREPO = "http://github.com/apache/openserverless-task"

// branch where download tasks
// defaults to test - will be changed in compilation

//go:embed version.txt
var OpsVersion string

//go:embed branch.txt
var OpsBranch string

//go:embed runtimes.json
var WSK_RUNTIMES_JSON string

// Represents opsroot.json. It is used to parse the file.
type OpsRootJSON struct {
	Version string                 `json:"version"`
	Config  map[string]interface{} `json:"config"`
}

// default port for ops server
const DefaultOpsPort = 9768

func getOpsPort() string {
	port := os.Getenv("OPS_PORT")
	if port == "" {
		port = fmt.Sprintf("%d", DefaultOpsPort)
	}
	//nolint:errcheck
	os.Setenv("OPS_PORT", port)
	return port
}

// get defaults
func getOpsRoot() (string, error) {
	root := os.Getenv("OPS_ROOT")
	if root == "" {
		dir, err := os.Getwd()
		if err == nil {
			root, err = locateOpsRoot(dir)
		}
		if err != nil {
			return "", err
		}
	}
	//nolint:errcheck
	os.Setenv("OPS_ROOT", root)
	return root, nil
}

func getOpsRepo() string {
	repo := os.Getenv("OPS_REPO")
	if repo == "" {
		repo = OPSREPO
	}
	//nolint:errcheck
	os.Setenv("OPS_REPO", repo)
	return repo
}

func getOpsBranch() string {
	branch := os.Getenv("OPS_BRANCH")
	if branch == "" {
		branch = strings.TrimSpace(OpsBranch)
	}
	//nolint:errcheck
	os.Setenv("OPS_BRANCH", branch)
	return branch
}

func readOpsRootFile(dir string) (OpsRootJSON, error) {
	data := OpsRootJSON{}
	json_buf, err := os.ReadFile(joinpath(dir, OPSROOT))
	if err != nil {
		return OpsRootJSON{}, err
	}
	if err := json.Unmarshal(json_buf, &data); err != nil {
		warn("opsroot.json parsed with an error", err)
	}
	return data, nil
}

// utils
func joinpath(dir string, file string) string {
	return filepath.Join(dir, file)
}

func split(s string) []string {
	return strings.Fields(s)
}

func exists(dir string, file string) bool {
	_, err := os.Stat(joinpath(dir, file))
	return !os.IsNotExist(err)
}

func isDir(fileOrDir string) bool {
	fileInfo, err := os.Stat(fileOrDir)
	if err != nil {
		debug(err)
		return false
	}

	// Check if the file is a directory
	if fileInfo.IsDir() {
		return true
	}
	return false
}

func parent(dir string) string {
	return filepath.Dir(dir)
}

func readfile(file string) string {
	dat, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return string(dat)
}

//var logger log.Logger = log.New(os.Stderr, "", 0)

func warn(args ...any) {
	log.Println(args...)
}

var tracing = os.Getenv("TRACE") != ""

func trace(args ...any) {
	if tracing {
		log.Println(append([]any{"TRACE: "}, args...))
	}
}

func Trace(args ...any) {
	trace(args...)
}

var debugging = os.Getenv("DEBUG") != "" || os.Getenv("TRACE") != ""

func debug(args ...any) {
	if debugging {
		log.Println(append([]any{"DEBUG: "}, args...))
	}
}
func debugf(format string, args ...any) {
	if debugging {
		log.Printf("DEBUG: "+format+"\n", args...)
	}
}

func touch(dir string, name string) error {
	tgt := filepath.Join(dir, name)
	trace("touch", tgt)
	f, err := os.Create(tgt)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}
