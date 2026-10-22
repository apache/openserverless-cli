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
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func copyFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}
func setupPluginTest(dir string, t *testing.T) string {
	t.Helper()
	// create the oplugins-test folder
	opluginsTestDir := filepath.Join(dir, "oplugins-test")
	err := os.MkdirAll(opluginsTestDir, 0755)
	require.NoError(t, err)

	// copy the opsroot.json from tests/oplugins into the oplugins-test folder
	opsRootJSON := filepath.Join("tests", "oplugins", "opsroot.json")
	err = copyFile(opsRootJSON, filepath.Join(opluginsTestDir, "opsroot.json"))
	require.NoError(t, err)

	// copy opsfile.yml from tests/oplugins into the oplugins-test folder
	opsfileYML := filepath.Join("tests", "oplugins", "opsfile.yml")
	err = copyFile(opsfileYML, filepath.Join(opluginsTestDir, "opsfile.yml"))
	require.NoError(t, err)

	return opluginsTestDir
}

func TestGetAllOpsRootPlugins(t *testing.T) {
	t.Run("success: get all the opsroots.json from plugins with 1 plugin", func(t *testing.T) {
		tempDir := t.TempDir()
		plgFolder := setupPluginTest(tempDir, t)
		os.Setenv("OPS_ROOT_PLUGIN", tempDir)

		opsRoots, err := GetOpsRootPlugins()
		require.NoError(t, err)
		require.Len(t, opsRoots, 1)
		require.Equal(t, joinpath(plgFolder, OPSROOT), opsRoots[getPluginName(plgFolder)])
	})

	t.Run("success: get all the opsroots.json from plugins with 2 plugins", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OPS_ROOT_PLUGIN", tempDir)
		plgFolder := setupPluginTest(tempDir, t)

		// create the oplugins-test2 folder
		opluginsTestDir := filepath.Join(tempDir, "oplugins-test2")
		err := os.MkdirAll(opluginsTestDir, 0755)
		require.NoError(t, err)

		// copy the opsroot.json from tests/oplugins into the oplugins-test folder
		opsRootJSON := filepath.Join("tests", "oplugins", "opsroot.json")
		err = copyFile(opsRootJSON, filepath.Join(opluginsTestDir, "opsroot.json"))
		require.NoError(t, err)

		// copy opsfile.yml from tests/oplugins into the oplugins-test folder
		opsfileYML := filepath.Join("tests", "oplugins", "opsfile.yml")
		err = copyFile(opsfileYML, filepath.Join(opluginsTestDir, "opsfile.yml"))
		require.NoError(t, err)

		opsRoots, err := GetOpsRootPlugins()
		require.NoError(t, err)
		require.Len(t, opsRoots, 2)
		require.Equal(t, joinpath(plgFolder, OPSROOT), opsRoots[getPluginName(plgFolder)])
		require.Equal(t, joinpath(opluginsTestDir, OPSROOT), opsRoots[getPluginName(opluginsTestDir)])
	})

	t.Run("empty: no plugins folder found (oplugins-*)", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OPS_ROOT_PLUGIN", tempDir)

		// Test when the folder is not found
		opsRoots, err := GetOpsRootPlugins()
		require.NoError(t, err)
		require.Empty(t, opsRoots)
	})
}

func TestFindPluginTask(t *testing.T) {
	t.Run("success: plugin task found in ./oplugins-test", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OPS_ROOT_PLUGIN", tempDir)
		plgFolder := setupPluginTest(tempDir, t)

		fld, err := findTaskInPlugins("test")
		require.NoError(t, err)
		require.Equal(t, plgFolder, fld)
	})

	t.Run("error: no plugins folder found (oplugins-*)", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OPS_ROOT_PLUGIN", tempDir)

		// Test when the folder is not found
		fld, err := findTaskInPlugins("grep")
		require.Error(t, err)
		require.Empty(t, fld)
	})
}

func TestNewPlugins(t *testing.T) {
	t.Run("create plugins struct with valid local dir", func(t *testing.T) {
		tempDir := t.TempDir()
		plgFolder := setupPluginTest(tempDir, t)

		os.Setenv("OPS_ROOT_PLUGIN", tempDir)

		p, err := newPlugins()
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Len(t, p.local, 1)
		require.Equal(t, plgFolder, p.local[0])
	})

	t.Run("non existent local dir results in empty local field", func(t *testing.T) {
		localDir := "/path/to/nonexistent/dir"
		os.Setenv("OPS_ROOT_PLUGIN", localDir)
		p, err := newPlugins()
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Len(t, p.local, 0)
	})
}

func Example_pluginsPrint() {
	p := plugins{
		local: make([]string, 0),
		ops:   make([]string, 0),
	}
	p.print()
	// Output
	// No plugins installed. Use 'ops -plugin' to add new ones.
}

func TestCheckGitRepo(t *testing.T) {
	tests := []struct {
		url          string
		expectedRepo bool
		expectedName string
	}{
		{
			url:          "https://github.com/giusdp/oplugins-test",
			expectedRepo: true,
			expectedName: "oplugins-test",
		},
		{
			url:          "https://github.com/giusdp/oplugins-test.git",
			expectedRepo: true,
			expectedName: "oplugins-test",
		},
		{
			url:          "https://github.com/giusdp/some-repo",
			expectedRepo: false,
			expectedName: "",
		},
		{
			url:          "https://github.com/giusdp/oplugins-repo.git",
			expectedRepo: true,
			expectedName: "oplugins-repo",
		},
		{
			url:          "https://github.com/oplugins-1234/repo",
			expectedRepo: false,
			expectedName: "",
		},
		{
			url:          "https://github.com/giusdp/another-repo.git",
			expectedRepo: false,
			expectedName: "",
		},
	}

	for _, test := range tests {
		isOpluginsRepo, repoName := checkGitRepo(test.url)
		require.Equal(t, test.expectedName, repoName)
		require.Equal(t, test.expectedRepo, isOpluginsRepo)
	}
}

func Test_getPluginName(t *testing.T) {

	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "oplugins-test",
			expected: "test",
		},
		{
			name:     "oplugins-test-123",
			expected: "test-123",
		},
		{
			name:     "test",
			expected: "test",
		},
		{
			name:     "a/fake/path/to/oplugins-test",
			expected: "test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			name := getPluginName(tc.name)
			require.Equal(t, tc.expected, name)
		})
	}
}
