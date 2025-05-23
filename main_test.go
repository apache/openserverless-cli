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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/require"
)

/// test utils

// pr print args
func pr(args ...any) {
	fmt.Println(args...)
}

func after(where, all string) string {
	// Find the index of the first occurrence of `where` in `all`
	index := strings.Index(all, where)

	// If `where` is not found, return an empty string
	if index == -1 {
		return all
	}

	// Return the substring after `where`
	return all[index+len(where):]
}

// as creates a string array
// func as(s ...string) []string {
// 	return s
// }

var homeDir = ""
var workDir = ""

// normalize path replacing the variable path with /work
func npath(dir string) string {
	return strings.Replace(dir, workDir, "/work", -1)
}

// normalize path replaching the variable home part with /home
func nhpath(dir string) string {
	//fmt.Println("dir", dir, "home", homeDir)
	return strings.Replace(dir, homeDir, "/home", -1)
}

func RemoveAll(dir string) error {
	//fmt.Println("****" + dir)
	os.RemoveAll(dir)
	return nil
}

func TestMain(m *testing.M) {
	debugging = true
	tracing = true
	wd, _ := os.Getwd()
	workDir, _ = filepath.Abs(wd)
	homeDir, _ = homedir.Dir()
	taskDryRun = true
	os.Exit(m.Run())
}

func TestSetupOpsRootPlugin(t *testing.T) {
	// Test case 1: OPS_ROOT_PLUGIN is not set
	os.Unsetenv("OPS_ROOT_PLUGIN")
	os.Setenv("OPS_PWD", "/path/to/ops")
	setOpsRootPluginEnv()
	if os.Getenv("OPS_ROOT_PLUGIN") != "/path/to/ops" {
		t.Errorf("OPS_ROOT_PLUGIN not set correctly, expected /path/to/ops but got %s", os.Getenv("OPS_ROOT_PLUGIN"))
	}

	// Test case 2: OPS_ROOT_PLUGIN is already set
	os.Setenv("OPS_ROOT_PLUGIN", "/path/to/ops/root")
	setOpsRootPluginEnv()
	if os.Getenv("OPS_ROOT_PLUGIN") != "/path/to/ops/root" {
		t.Errorf("OPS_ROOT_PLUGIN not set correctly, expected /path/to/ops/root but got %s", os.Getenv("OPS_ROOT_PLUGIN"))
	}
}

func TestParseInvokeArgs(t *testing.T) {
	t.Run("Test case 1: No arguments with \"=\"", func(t *testing.T) {
		input1 := []string{}
		expected1 := []string{}
		output1 := parseInvokeArgs(input1)
		require.Equal(t, expected1, output1)
	})

	t.Run("Test case 2: Single argument with \"=\"", func(t *testing.T) {
		input2 := []string{"key=value"}
		expected2 := []string{"-p", "key", "value"}
		output2 := parseInvokeArgs(input2)
		require.Equal(t, expected2, output2)
	})

	t.Run("Test case 3: Multiple arguments with \"=\"", func(t *testing.T) {

		input3 := []string{"key1=value1", "key2=value2", "key3=value3"}
		expected3 := []string{"-p", "key1", "value1", "-p", "key2", "value2", "-p", "key3", "value3"}
		output3 := parseInvokeArgs(input3)
		require.Equal(t, expected3, output3)
	})

	t.Run("Test case 4: Mixed arguments with \"=\" and without \"=\"", func(t *testing.T) {
		input4 := []string{"key1=value1", "-p", "key2", "value2", "key3=value3"}
		expected4 := []string{"-p", "key1", "value1", "-p", "key2", "value2", "-p", "key3", "value3"}
		output4 := parseInvokeArgs(input4)
		require.Equal(t, expected4, output4)
	})
}
