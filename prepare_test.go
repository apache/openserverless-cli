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
	"os"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/require"
)

func Example_locate() {
	_ = os.Chdir(workDir)
	dir, err := locateOpsRoot("tests")
	pr(1, err, npath(dir))
	dir, err = locateOpsRoot(joinpath("tests", "olaris"))
	pr(2, err, npath(dir))
	dir, err = locateOpsRoot(joinpath("tests", joinpath("olaris", "sub")))
	pr(3, err, npath(dir))
	// Output:
	// 1 <nil> /work/tests/olaris
	// 2 <nil> /work/tests/olaris
	// 3 <nil> /work/tests/olaris
}

// TODO: undestand why it fails when executed with others
// it works executed alone
func Example_download() {
	_ = os.Chdir(workDir)
	OpsBranch = "0.1.0"
	os.Setenv("OPS_BRANCH", OpsBranch)
	opsdir, _ := homedir.Expand("~/.ops")
	_ = RemoveAll(opsdir)
	_, _ = downloadTasksFromGitHub(true, true)
	dir, err := locateOpsRoot(".")
	pr(1, err, nhpath(dir))
	_, _ = downloadTasksFromGitHub(true, true)
	dir, err = locateOpsRoot(".")
	pr(2, err, nhpath(dir))
	// Output:
	// Cloning tasks...
	// Tasks downloaded successfully
	// 1 <nil> /home/.ops/0.1.0/olaris
	// Updating tasks...
	// Tasks are already up to date!
	// 2 <nil> /home/.ops/0.1.0/olaris
}

func Example_locate_root() {
	_ = os.Chdir(workDir)
	OpsBranch = "0.1.0"
	opsdir, _ := homedir.Expand("~/.ops")
	_ = RemoveAll(opsdir)
	_, err := locateOpsRoot(".")
	pr(1, err)
	dir, err := locateOpsRoot("tests")
	pr(2, err, npath(dir))
	// Output:
	// 1 cannot find opsfiles, download them with ops -update
	// 2 <nil> /work/tests/olaris
}

func Test_setOpsOlarisHash(t *testing.T) {
	_ = os.Chdir(workDir)
	OpsBranch = "0.1.0"
	opsdir, _ := homedir.Expand("~/.ops")
	_ = RemoveAll(opsdir)
	dir, _ := downloadTasksFromGitHub(true, true)
	err := setOpsOlarisHash(dir)
	require.NoError(t, err)
	require.NotEmpty(t, os.Getenv("OPS_OLARIS"))
}
