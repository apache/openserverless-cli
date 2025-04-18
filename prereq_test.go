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
)

func Example_execPrereqTask() {
	fmt.Println(execPrereqTask("bin", "bun"))
	// Output:
	// invoking prereq for bun
	// <nil>
}

func Example_loadPrereq() {
	//downloadPrereq("")
	dir := joinpath(workDir, "tests")
	t, v, err := loadPrereq(dir)
	fmt.Println(err, len(t), len(v))
	dir = joinpath(dir, "prereq")
	//dir = "/home/msciab/.ops/0.1.0/olaris/"
	tasks, versions, err := loadPrereq(dir)
	//fmt.Println(prq)
	fmt.Println(err, tasks)
	fmt.Println(versions)
	// Output:
	// <nil> 0 0
	// <nil> [bun coreutils]
	// [v1.11.20 0.0.27]
}

func Example_ensureBindir() {
	// cleanup
	bindir, err := EnsureBindir()
	if err == nil {
		RemoveAll(bindir)
	} else {
		fmt.Println(err)
	}
	// ensure no dir
	_, err = os.Stat(bindir)
	if err != nil {
		fmt.Println(1, after(":", err.Error()))
	}
	bindir1, err := EnsureBindir()
	if err != nil {
		fmt.Println(err)
	}
	_, err = os.Stat(bindir1)
	fmt.Println(2, err)
	fmt.Println(3, bindir == bindir1)

	// Output:
	// 1  no such file or directory
	// 2 <nil>
	// 3 true
}

func Example_touchAndClean() {
	bindir, err := EnsureBindir()
	if err != nil {
		RemoveAll(bindir)
	}
	bindir, _ = EnsureBindir()
	touch(bindir, "hello")
	err = touchAndClean(bindir, "hello", "1.2.3")
	fmt.Println(err, exists(bindir, "hello"), exists(bindir, "hello-1.2.3"), exists(bindir, "hello-1.2.4"))
	err = touchAndClean(bindir, "hello", "1.2.4")
	fmt.Println(err, exists(bindir, "hello"), exists(bindir, "hello-1.2.3"), exists(bindir, "hello-1.2.4"))
	// Output:
	// <nil> true true false
	// <nil> true false true
}

func Example_downloadPrereq() {
	bindir, err := EnsureBindir()
	if err != nil {
		RemoveAll(bindir)
	}
	PrereqSeenMap = map[string]string{}

	prqdir := joinpath(joinpath(workDir, "tests"), "prereq")
	tasks, versions, _ := loadPrereq(prqdir)
	fmt.Println("1", downloadPrereq(tasks[0], versions[0]))
	fmt.Println("2", downloadPrereq(tasks[0], versions[0]))

	tasks, versions, _ = loadPrereq(joinpath(prqdir, "sub"))
	//fmt.Println(tasks, versions)
	//fmt.Println(PrereqSeenMap)
	fmt.Println("3", downloadPrereq("bun", versions[0]))
	// Output:
	// downloading bun v1.11.20
	// 1 <nil>
	// 2 <nil>
	// 3 WARNING: bun prerequisite found twice with different versions!
	// Previous version: v1.11.20, ignoring v1.11.21
}

func Example_ensurePrereq() {
	bindir, err := EnsureBindir()
	if err == nil {
		RemoveAll(bindir)
	} else {
		fmt.Printf("ERROR CANNOT REMOVE DIR %v\n", err)
	}
	PrereqSeenMap = map[string]string{}
	dir := joinpath(joinpath(workDir, "tests"), "prereq")
	fmt.Println(ensurePrereq(dir))
	fmt.Println(ensurePrereq(joinpath(dir, "sub")))
	// Unordered output:
	// downloading bun v1.11.20
	// downloading coreutils 0.0.27
	// <nil>
	// error in prereq bun: WARNING: bun prerequisite found twice with different versions!
	// Previous version: v1.11.20, ignoring v1.11.21
	// <nil>
}
