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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/exp/slices"
)

func Example_nuvArg() {
	// test
	_ = os.Chdir(workDir)
	olaris, _ := filepath.Abs(joinpath("tests", "olaris"))
	err := Nuv(olaris, split("testcmd"))
	pr(2, err)
	err = Nuv(olaris, split("testcmd arg"))
	pr(3, err)
	err = Nuv(olaris, split("testcmd arg VAR=1"))
	pr(4, err)
	err = Nuv(olaris, split("testcmd VAR=1 arg"))
	pr(5, err)
	// Output:
	// (olaris) task [-t nuvfile.yml testcmd --]
	// 2 <nil>
	// (olaris) task [-t nuvfile.yml testcmd -- arg]
	// 3 <nil>
	// (olaris) task [-t nuvfile.yml testcmd VAR=1 -- arg]
	// 4 <nil>
	// (olaris) task [-t nuvfile.yml testcmd VAR=1 -- arg]
	//5 <nil>
}

func ExampleNuv() {
	// test
	_ = os.Chdir(workDir)
	os.Setenv("TEST_VAR", "evar")
	olaris, _ := filepath.Abs(joinpath("tests", "olaris"))
	err := Nuv(olaris, split(""))
	pr(1, err)
	err = Nuv(olaris, split("sub"))
	pr(4, err)
	err = Nuv(olaris, split("sub opts"))
	pr(5, err)
	_ = Nuv(olaris, split("sub opts ciao 1"))
	// pr(6, err)
	// Output:
	// (olaris) task [-t nuvfile.yml -l]
	//
	// Plugins:
	// 1 <nil>
	// (sub) task [-t nuvfile.yml -l]
	//
	// Plugins:
	// 4 <nil>
	// Usage:
	//   opts hello
	//   opts ciao <name>... [-c] [-e evar]
	//   opts salve <name> hi <x> <y> [--fl=<flag>]
	//   opts sayonara (opt1|opt2) <x> <y> [--fa|--fb]
	//   opts -h | --help | --version
	//
	// Plugins:
	// 5 <nil>
	// (opts) task [-t nuvfile.yml ciao $TEST_VAR= __fa=false __fb=false __fl= __help=false __version=false _c=false _e=false _h=false _name_=('1') _x_= _y_= ciao=true hello=false hi=false opt1=false opt2=false salve=false sayonara=false]
	// 6 <nil>
}

func ExampleParseArgs() {
	_ = os.Chdir(workDir)
	usage := readfile("tests/olaris/sub/opts/nuvopts.txt")
	args := parseArgs(usage, split("ciao mike miri max"))
	pr(1, args)
	args = parseArgs(usage, split("ciao mike -c"))
	pr(2, args)
	args = parseArgs(usage, split("salve max hi 1 2 --fl=3"))
	pr(3, args)
	args = parseArgs(usage, split("sayonara opt2 4 5 --fb"))
	pr(4, args)
	// Output:
	// 1 [$TEST_VAR= __fa=false __fb=false __fl= __help=false __version=false _c=false _e=false _h=false _name_=('mike' 'miri' 'max') _x_= _y_= ciao=true hello=false hi=false opt1=false opt2=false salve=false sayonara=false]
	// 2 [$TEST_VAR= __fa=false __fb=false __fl= __help=false __version=false _c=true _e=false _h=false _name_=('mike') _x_= _y_= ciao=true hello=false hi=false opt1=false opt2=false salve=false sayonara=false]
	// 3 [$TEST_VAR= __fa=false __fb=false __fl=3 __help=false __version=false _c=false _e=false _h=false _name_=('max') _x_=1 _y_=2 ciao=false hello=false hi=true opt1=false opt2=false salve=true sayonara=false]
	// 4 [$TEST_VAR= __fa=false __fb=true __fl= __help=false __version=false _c=false _e=false _h=false _name_=() _x_=4 _y_=5 ciao=false hello=false hi=false opt1=false opt2=true salve=false sayonara=true]
}

func Test_validateTaskName(t *testing.T) {
	testNuvfile := "tasks:\n  task1: a\n  task2: b\n  test: c\n"

	type validateTaskTest struct {
		argTask  string
		expected string
	}

	var validateTaskTests = []validateTaskTest{
		{"help", "help"},
		{"task1", "task1"},
		{"te", "test"},
		{"t", "ambiguous command: t."},
		{"no-task", "no command named no-task found"},
		{"", "command name is empty"},
	}

	tmpDir := createTmpNuvfile(t, testNuvfile)
	defer os.RemoveAll(tmpDir)
	for _, tt := range validateTaskTests {
		task, err := validateTaskName(tmpDir, tt.argTask)
		if err != nil && !strings.Contains(err.Error(), tt.expected) {
			t.Fatalf("want error: %s, got: %v", tt.expected, err)
		}
		if err == nil && task != tt.expected {
			t.Fatalf("want command: %s, got: %s", tt.argTask, task)
		}

	}
}

func Example_setupTmp() {
	_ = os.Chdir(workDir)
	nuvdir, _ := homedir.Expand("~/.nuv")
	os.RemoveAll(nuvdir)
	setupTmp()
	fmt.Println(nhpath(os.Getenv("NUV_TMP")))
	os.RemoveAll(nuvdir)
	// Output:
	// /home/.nuv/tmp
}

func Example_loadArgs() {
	_ = os.Chdir(workDir)
	fmt.Println(1, loadSavedArgs())
	_ = os.Chdir(joinpath("tests", "testdata"))
	fmt.Println(2, loadSavedArgs())
	// Output:
	// 1 []
	// 2 [V1=hello V2=hello V2=world]
}

func Test_getTaskNamesList(t *testing.T) {
	t.Run("empty nuvfile should return empty array", func(t *testing.T) {
		tmpDir := createTmpNuvfile(t, "")

		tasks := getTaskNamesList(tmpDir)
		if len(tasks) != 0 {
			t.Fatalf("expected 0 tasks, got %d", len(tasks))
		}
	})

	t.Run("should return array of task names if tasks in nuvfile", func(t *testing.T) {
		tmpDir := createTmpNuvfile(t, "tasks:\n  task1: a\n  task2: b\n")
		defer os.RemoveAll(tmpDir)

		tasks := getTaskNamesList(tmpDir)
		if len(tasks) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(tasks))
		}

		if !slices.Contains(tasks, "task1") || !slices.Contains(tasks, "task2") {
			t.Fatalf("expected task1 and task2, got %v", tasks)
		}
	})

	t.Run("should return array of task names if tasks in nuvfile + subfolders names as tasks with nuvfile in them", func(t *testing.T) {

		tmpDir := createTmpNuvfile(t, "tasks:\n  task1: a\n  task2: b\n")
		defer os.RemoveAll(tmpDir)

		// create subfolder with nuvfile
		subDir := filepath.Join(tmpDir, "sub")
		err := os.Mkdir(subDir, 0755)
		if err != nil {
			t.Fatal(err)
		}
		subNuvfile := filepath.Join(subDir, "nuvfile.yml")
		err = os.WriteFile(subNuvfile, []byte("tasks:\n  task3: a\n  task4: b\n"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// create subfolder without nuvfile
		subDir2 := filepath.Join(tmpDir, "sub2")
		err = os.Mkdir(subDir2, 0755)
		if err != nil {
			t.Fatal(err)
		}

		tasks := getTaskNamesList(tmpDir)
		if len(tasks) != 3 {
			t.Fatalf("expected 3 tasks, got %d", len(tasks))
		}

		if !slices.Contains(tasks, "task1") || !slices.Contains(tasks, "task2") || !slices.Contains(tasks, "sub") {
			t.Fatalf("expected task1, task2, sub, got %v", tasks)
		}
	})

	t.Run("avoids duplicate when subfolder with nuvfile has same name as task", func(t *testing.T) {

		tmpDir := createTmpNuvfile(t, "tasks:\n  task1: a\n  task2: b\n")
		defer os.RemoveAll(tmpDir)

		// create subfolder with nuvfile
		subDir := filepath.Join(tmpDir, "task1")
		err := os.Mkdir(subDir, 0755)
		if err != nil {
			t.Fatal(err)
		}
		subNuvfile := filepath.Join(subDir, "nuvfile.yml")
		err = os.WriteFile(subNuvfile, []byte("tasks:\n  task3: a\n  task4: b\n"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		tasks := getTaskNamesList(tmpDir)
		if len(tasks) != 2 {
			t.Fatalf("expected 2 tasks, got %d: %v", len(tasks), tasks)
		}

		if !slices.Contains(tasks, "task1") || !slices.Contains(tasks, "task2") {
			t.Fatalf("expected task1, task2, got %v", tasks)
		}
	})

}

// createTmpNuvfile creates a temp folder with nuvfile.yml
func createTmpNuvfile(t *testing.T, content string) string {
	t.Helper()
	// create temp folder with nuvfile.yml
	tmpDir, err := os.MkdirTemp("", "nuv-test")
	if err != nil {
		t.Fatal(err)
	}

	// create nuvfile.yml
	nuvfile := filepath.Join(tmpDir, "nuvfile.yml")
	err = os.WriteFile(nuvfile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return tmpDir
}
