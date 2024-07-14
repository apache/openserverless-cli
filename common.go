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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// default files
const NUVFILE = "nuvfile.yml"
const NUVROOT = "nuvroot.json"
const NUVOPTS = "nuvopts.txt"
const CONFIGFILE = "config.json"

// repo where download tasks
const NUVREPO = "http://github.com/nuvolaris/olaris"

// branch where download tasks
// defaults to test - will be changed in compilation
var NuvVersion = "main"
var NuvBranch = "3.0.0"

// Represents nuvroot.json. It is used to parse the file.
type NuvRootJSON struct {
	Version string                 `json:"version"`
	Config  map[string]interface{} `json:"config"`
}

// default port for nuv server
const DefaultNuvPort = 9768

func getNuvPort() string {
	port := os.Getenv("NUV_PORT")
	if port == "" {
		port = fmt.Sprintf("%d", DefaultNuvPort)
	}
	//nolint:errcheck
	os.Setenv("NUV_PORT", port)
	return port
}

// get defaults
func getNuvRoot() (string, error) {
	root := os.Getenv("NUV_ROOT")
	if root == "" {
		dir, err := os.Getwd()
		if err == nil {
			root, err = locateNuvRoot(dir)
		}
		if err != nil {
			return "", err
		}
	}
	//nolint:errcheck
	os.Setenv("NUV_ROOT", root)
	return root, nil
}

func getNuvRepo() string {
	repo := os.Getenv("NUV_REPO")
	if repo == "" {
		repo = NUVREPO
	}
	//nolint:errcheck
	os.Setenv("NUV_REPO", repo)
	return repo
}

func getNuvBranch() string {
	branch := os.Getenv("NUV_BRANCH")
	if branch == "" {
		branch = NuvBranch
	}
	//nolint:errcheck
	os.Setenv("NUV_BRANCH", branch)
	return branch
}

func readNuvRootFile(dir string) (NuvRootJSON, error) {
	data := NuvRootJSON{}
	json_buf, err := os.ReadFile(joinpath(dir, NUVROOT))
	if err != nil {
		return NuvRootJSON{}, err
	}
	if err := json.Unmarshal(json_buf, &data); err != nil {
		warn("nuvroot.json parsed with an error", err)
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

var debugging = os.Getenv("DEBUG") != ""

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
