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
	// create the olaris-test folder
	olarisTestDir := filepath.Join(dir, "olaris-test")
	err := os.MkdirAll(olarisTestDir, 0755)
	require.NoError(t, err)

	// copy the nuvroot.json from tests/olaris into the olaris-test folder
	nuvRootJSON := filepath.Join("tests", "olaris", "nuvroot.json")
	err = copyFile(nuvRootJSON, filepath.Join(olarisTestDir, "nuvroot.json"))
	require.NoError(t, err)

	// copy nuvfile.yml from tests/olaris into the olaris-test folder
	nuvfileYML := filepath.Join("tests", "olaris", "nuvfile.yml")
	err = copyFile(nuvfileYML, filepath.Join(olarisTestDir, "nuvfile.yml"))
	require.NoError(t, err)

	return olarisTestDir
}

func TestGetAllNuvRootPlugins(t *testing.T) {
	t.Run("success: get all the nuvroots.json from plugins with 1 plugin", func(t *testing.T) {
		tempDir := t.TempDir()
		plgFolder := setupPluginTest(tempDir, t)
		os.Setenv("NUV_ROOT_PLUGIN", tempDir)

		nuvRoots, err := GetNuvRootPlugins()
		require.NoError(t, err)
		require.Len(t, nuvRoots, 1)
		require.Equal(t, joinpath(plgFolder, NUVROOT), nuvRoots[getPluginName(plgFolder)])
	})

	t.Run("success: get all the nuvroots.json from plugins with 2 plugins", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("NUV_ROOT_PLUGIN", tempDir)
		plgFolder := setupPluginTest(tempDir, t)

		// create the olaris-test2 folder
		olarisTestDir := filepath.Join(tempDir, "olaris-test2")
		err := os.MkdirAll(olarisTestDir, 0755)
		require.NoError(t, err)

		// copy the nuvroot.json from tests/olaris into the olaris-test folder
		nuvRootJSON := filepath.Join("tests", "olaris", "nuvroot.json")
		err = copyFile(nuvRootJSON, filepath.Join(olarisTestDir, "nuvroot.json"))
		require.NoError(t, err)

		// copy nuvfile.yml from tests/olaris into the olaris-test folder
		nuvfileYML := filepath.Join("tests", "olaris", "nuvfile.yml")
		err = copyFile(nuvfileYML, filepath.Join(olarisTestDir, "nuvfile.yml"))
		require.NoError(t, err)

		nuvRoots, err := GetNuvRootPlugins()
		require.NoError(t, err)
		require.Len(t, nuvRoots, 2)
		require.Equal(t, joinpath(plgFolder, NUVROOT), nuvRoots[getPluginName(plgFolder)])
		require.Equal(t, joinpath(olarisTestDir, NUVROOT), nuvRoots[getPluginName(olarisTestDir)])
	})

	t.Run("empty: no plugins folder found (olaris-*)", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("NUV_ROOT_PLUGIN", tempDir)

		// Test when the folder is not found
		nuvRoots, err := GetNuvRootPlugins()
		require.NoError(t, err)
		require.Empty(t, nuvRoots)
	})
}

func TestFindPluginTask(t *testing.T) {
	t.Run("success: plugin task found in ./olaris-test", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("NUV_ROOT_PLUGIN", tempDir)
		plgFolder := setupPluginTest(tempDir, t)

		fld, err := findTaskInPlugins("test")
		require.NoError(t, err)
		require.Equal(t, plgFolder, fld)
	})

	t.Run("error: no plugins folder found (olaris-*)", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("NUV_ROOT_PLUGIN", tempDir)

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

		os.Setenv("NUV_ROOT_PLUGIN", tempDir)

		p, err := newPlugins()
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Len(t, p.local, 1)
		require.Equal(t, plgFolder, p.local[0])
	})

	t.Run("non existent local dir results in empty local field", func(t *testing.T) {
		localDir := "/path/to/nonexistent/dir"
		os.Setenv("NUV_ROOT_PLUGIN", localDir)
		p, err := newPlugins()
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Len(t, p.local, 0)
	})
}

func Example_pluginsPrint() {
	p := plugins{
		local: make([]string, 0),
		nuv:   make([]string, 0),
	}
	p.print()
	// Output
	// No plugins installed. Use 'nuv -plugin' to add new ones.
}

func TestCheckGitRepo(t *testing.T) {
	tests := []struct {
		url          string
		expectedRepo bool
		expectedName string
	}{
		{
			url:          "https://github.com/giusdp/olaris-test",
			expectedRepo: true,
			expectedName: "olaris-test",
		},
		{
			url:          "https://github.com/giusdp/olaris-test.git",
			expectedRepo: true,
			expectedName: "olaris-test",
		},
		{
			url:          "https://github.com/giusdp/some-repo",
			expectedRepo: false,
			expectedName: "",
		},
		{
			url:          "https://github.com/giusdp/olaris-repo.git",
			expectedRepo: true,
			expectedName: "olaris-repo",
		},
		{
			url:          "https://github.com/olaris-1234/repo",
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
		isOlarisRepo, repoName := checkGitRepo(test.url)
		require.Equal(t, test.expectedName, repoName)
		require.Equal(t, test.expectedRepo, isOlarisRepo)
	}
}

func Test_getPluginName(t *testing.T) {

	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "olaris-test",
			expected: "test",
		},
		{
			name:     "olaris-test-123",
			expected: "test-123",
		},
		{
			name:     "test",
			expected: "test",
		},
		{
			name:     "a/fake/path/to/olaris-test",
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
