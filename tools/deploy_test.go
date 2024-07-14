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
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractArgs(t *testing.T) {
	// Prepare a temporary test file
	tmpDir := t.TempDir()
	file, err := os.CreateTemp(tmpDir, "testfile*.txt")
	if err != nil {
		t.Fatal(err)
	}

	content := `#-arg1
//-arg2
other content
#-arg3
//-arg4`

	_, _ = file.WriteString(content)
	file.Close()

	// Test cases
	tests := []struct {
		name     string
		files    []string
		expected []string
	}{
		{
			name:     "Extract args from existing file",
			files:    []string{file.Name()},
			expected: []string{"-arg1", "-arg2", "-arg3", "-arg4"},
		},
		{
			name:     "Extract args from non-existing file",
			files:    []string{"non_existing_file.txt"},
			expected: []string{},
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractArgs(tc.files)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestDeployPackage(t *testing.T) {
	// create a testpkg.args file
	tmpDir := t.TempDir()
	//create packages folder
	_ = os.MkdirAll(tmpDir+"/packages", 0755)
	file, err := os.Create(filepath.Join(tmpDir, "packages", "testpkg1.args"))
	if err != nil {
		t.Fatal(err)
	}

	content := `#-arg1
//-arg2
other content
#-arg3
//-arg4`

	_, _ = file.WriteString(content)
	file.Close()

	// Test cases
	tests := []struct {
		name             string
		ctx              deployCtx
		pkg              string
		expectedExecuted map[string]bool
		expectedLog      string
	}{
		{
			name: "Deploy package with no args",
			ctx:  deployCtx{packageCmdExecuted: make(map[string]bool), dryRun: true, path: "."},
			pkg:  "testpkg",
			expectedExecuted: map[string]bool{
				"nuv package update testpkg": true,
			},
			expectedLog: "Would run: nuv package update testpkg\n",
		},
		{
			name: "Deploy package with args",
			ctx:  deployCtx{packageCmdExecuted: make(map[string]bool), dryRun: true, path: tmpDir},
			pkg:  "testpkg1",
			expectedExecuted: map[string]bool{
				fmt.Sprintf("nuv package update %s -arg1 -arg2 -arg3 -arg4", "testpkg1"): true,
			},
			expectedLog: fmt.Sprintf("Would run: nuv package update %s -arg1 -arg2 -arg3 -arg4\n", "testpkg1"),
		},
	}

	// Original log output
	originalLogOutput := log.Writer()
	defer func() { log.SetOutput(originalLogOutput) }()

	// Buffer to capture log output
	var logOutput strings.Builder
	log.SetOutput(&logOutput)

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deployPackage(tc.ctx, tc.pkg)
			// Check if the command was executed
			require.Equal(t, tc.expectedExecuted, tc.ctx.packageCmdExecuted, "Expected number of executed commands does not match")
			require.Contains(t, logOutput.String(), tc.expectedLog, "Log output does not match")
			// Reset log output buffer
			logOutput.Reset()
		})
	}
}

func TestDeployAction(t *testing.T) {
	// create temp index.js with some args
	tmpDir := t.TempDir()
	_ = os.Mkdir(filepath.Join(tmpDir, "packages"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "packages", "packageName"), 0755)
	_ = os.Mkdir(filepath.Join(tmpDir, "packages", "packageName", "action"), 0755)
	file, err := os.Create(filepath.Join(tmpDir, "packages", "packageName", "action", "index.js"))
	if err != nil {
		t.Fatal(err)
	}

	content := `#--kind nodejs:20`

	_, _ = file.WriteString(content)

	// Test cases
	tests := []struct {
		name             string
		ctx              deployCtx
		artifact         string
		dryRun           bool
		expectedExecuted bool
		expectedLog      string
	}{
		{
			name:        "Deploy action from zip file",
			ctx:         deployCtx{dryRun: true, path: tmpDir, packageCmdExecuted: make(map[string]bool)},
			artifact:    "packages/packageName/action.zip",
			expectedLog: "nuv action update packageName/action packages/packageName/action.zip --kind nodejs:20",
		},
		{
			name:        "Deploy action from non-zip file",
			ctx:         deployCtx{dryRun: true, path: ".", packageCmdExecuted: make(map[string]bool)},
			artifact:    "packages/packageName/index.js",
			expectedLog: "nuv action update packageName/index packages/packageName/index.js",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Original log output
			originalLogOutput := log.Writer()
			defer func() { log.SetOutput(originalLogOutput) }()

			// Buffer to capture log output
			var logOutput strings.Builder
			log.SetOutput(&logOutput)

			err := deployAction(tc.ctx, tc.artifact)

			// Assertions
			require.NoError(t, err, "Error deploying action")
			require.Contains(t, logOutput.String(), tc.expectedLog, "Log output does not match")
			// Reset log output buffer
			logOutput.Reset()
		})
	}
}
func TestSplitPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "Split empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "Split path with single directory",
			path:     "dir1",
			expected: []string{"dir1"},
		},
		{
			name:     "Split path with multiple directories",
			path:     "dir1/dir2/dir3",
			expected: []string{"dir1", "dir2", "dir3"},
		},
		{
			name:     "Split path with leading slash",
			path:     "/dir1/dir2/dir3",
			expected: []string{"dir1", "dir2", "dir3"},
		},
		{
			name:     "Split path with trailing slash",
			path:     "dir1/dir2/dir3/",
			expected: []string{"dir1", "dir2", "dir3"},
		},
		{
			name:     "Split path with leading and trailing slash",
			path:     "/dir1/dir2/dir3/",
			expected: []string{"dir1", "dir2", "dir3"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := splitPath(tc.path)
			require.Equal(t, tc.expected, result)
		})
	}
}
