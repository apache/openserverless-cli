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
	dir, err := locateNuvRoot("tests")
	pr(1, err, npath(dir))
	dir, err = locateNuvRoot(joinpath("tests", "olaris"))
	pr(2, err, npath(dir))
	dir, err = locateNuvRoot(joinpath("tests", joinpath("olaris", "sub")))
	pr(3, err, npath(dir))
	// Output:
	// 1 <nil> /work/tests/olaris
	// 2 <nil> /work/tests/olaris
	// 3 <nil> /work/tests/olaris
}

// TODO: undestand why it fails when executed with others
// it works executed alone
func Failing_Example_download() {
	pr(1)
	_ = os.Chdir(workDir)
	NuvBranch = "0.1.0"
	nuvdir, _ := homedir.Expand("~/.nuv")
	_ = os.RemoveAll(nuvdir)
	_, _ = downloadTasksFromGitHub(true, true)
	dir, err := locateNuvRoot(".")
	pr(1, err, nhpath(dir))
	_, _ = downloadTasksFromGitHub(true, true)
	dir, err = locateNuvRoot(".")
	pr(2, err, nhpath(dir))
	// Output:
	// 1
	// Cloning tasks...
	// Tasks downloaded successfully
	// 1 <nil> /home/.nuv/0.1.0/olaris
	// Updating tasks...
	// Tasks are already up to date!
	// 2 <nil> /home/.nuv/0.1.0/olaris

}

func Example_locate_root() {
	_ = os.Chdir(workDir)
	NuvBranch = "0.1.0"
	nuvdir, _ := homedir.Expand("~/.nuv")
	_ = os.RemoveAll(nuvdir)
	_, err := locateNuvRoot(".")
	pr(1, err)
	dir, err := locateNuvRoot("tests")
	pr(2, err, npath(dir))
	// Output:
	// 1 we cannot find nuvfiles, download them with nuv -update
	// 2 <nil> /work/tests/olaris
}

func Test_setNuvOlarisHash(t *testing.T) {
	_ = os.Chdir(workDir)
	NuvBranch = "0.1.0-testing"
	nuvdir, _ := homedir.Expand("~/.nuv")
	_ = os.RemoveAll(nuvdir)
	_ = os.Setenv("NUV_BIN", workDir)
	dir, _ := downloadTasksFromGitHub(true, true)
	err := setNuvOlarisHash(dir)
	require.NoError(t, err)
	require.NotEmpty(t, os.Getenv("NUV_OLARIS"))
}
