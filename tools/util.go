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
	"fmt"
	"os"
	"os/exec"

	"github.com/nuvolaris/someutils"
	"github.com/nuvolaris/someutils/some"
)

var Utils = []string{
	"basename", "cat", "cp", "dirname",
	"gunzip", "gzip", "head",
	"ls", "mv", "pwd", "rm", "sleep",
	"tail", "tar", "tee", "touch", "tr",
	"unzip", "wc", "which", "zip",
}

func IsUtil(name string) bool {
	some.Init()
	for _, s := range Utils {
		if s == name {
			return true
		}
	}
	return false
}

func RunUtil(name string, args []string) (int, error) {
	if IsUtil(name) {
		full := append([]string{name}, args...)
		var err error
		var code int
		if useCoreutils() {
			code, err = runCoreUtils(full)
		} else {
			err, code = someutils.Call(name, full)
		}
		return code, err
	}
	return 1, fmt.Errorf("command %s not found", name)
}

func useCoreutils() bool {
	return os.Getenv("NUV_USE_COREUTILS") != ""
}

func runCoreUtils(full []string) (int, error) {
	cmd := exec.Command("coreutils", full...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	code := cmd.ProcessState.ExitCode()
	return code, err
}
