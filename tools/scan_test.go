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

package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_buildActionPlan(t *testing.T) {
	cmd := []string{"nuv", "-js", "script.js"}

	t.Run("returns error if no actions folder", func(t *testing.T) {
		tempDir := t.TempDir()
		_, err := buildCmdPlan(tempDir, cmd, "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("returns one arg cmd plan if empty folder", func(t *testing.T) {
		tempDir := t.TempDir()
		_ = os.Mkdir(filepath.Join(tempDir, actionsFolder), 0755)

		expectedPath := filepath.Join(tempDir, actionsFolder)

		plan, err := buildCmdPlan(tempDir, cmd, "")
		require.NoError(t, err)
		require.Len(t, plan.args, 1)
		require.Len(t, plan.args[0], 1)
		require.Equal(t, expectedPath, plan.args[0][0])
	})

	t.Run("returns one arg cmd plan if folder with one file", func(t *testing.T) {
		tempDir := t.TempDir()
		actionsDir := filepath.Join(tempDir, actionsFolder)
		_ = os.Mkdir(actionsDir, 0755)
		_ = os.WriteFile(filepath.Join(actionsDir, "file.js"), []byte("test"), 0644)

		expectedPath := filepath.Join(tempDir, actionsFolder)

		plan, err := buildCmdPlan(tempDir, cmd, "*")
		require.NoError(t, err)
		require.Len(t, plan.args, 1)
		require.Len(t, plan.args[0], 2)
		require.Equal(t, expectedPath, plan.args[0][0])
		require.Equal(t, "file.js", plan.args[0][1])
	})

	t.Run("returns plan if folder with multiple subfolders and files", func(t *testing.T) {
		tempDir := t.TempDir()
		actionsDir := filepath.Join(tempDir, actionsFolder)
		_ = os.Mkdir(actionsDir, 0755)
		_ = os.WriteFile(filepath.Join(actionsDir, "file.js"), []byte("test"), 0644)
		_ = os.Mkdir(filepath.Join(actionsDir, "subdir1"), 0755)
		_ = os.WriteFile(filepath.Join(actionsDir, "subdir1", "file1.js"), []byte("test"), 0644)
		_ = os.Mkdir(filepath.Join(actionsDir, "subdir2"), 0755)
		_ = os.WriteFile(filepath.Join(actionsDir, "subdir2", "file2.js"), []byte("test"), 0644)
		_ = os.Mkdir(filepath.Join(actionsDir, "subdir3"), 0755)

		// Expected args:
		// /actions file.js
		// /actions/subdir1 file1.js
		// /actions/subdir2 file2.js
		// /actions/subdir3

		plan, err := buildCmdPlan(tempDir, cmd, "*")

		require.NoError(t, err)
		require.Len(t, plan.args, 4)
		require.Len(t, plan.args[0], 2)
		require.Len(t, plan.args[1], 2)
		require.Len(t, plan.args[2], 2)
		require.Len(t, plan.args[3], 1)
		require.Equal(t, filepath.Join(tempDir, actionsFolder), plan.args[0][0])
		require.Equal(t, "file.js", plan.args[0][1])
		require.Equal(t, filepath.Join(tempDir, actionsFolder, "subdir1"), plan.args[1][0])
		require.Equal(t, "file1.js", plan.args[1][1])
		require.Equal(t, filepath.Join(tempDir, actionsFolder, "subdir2"), plan.args[2][0])
		require.Equal(t, "file2.js", plan.args[2][1])
		require.Equal(t, filepath.Join(tempDir, actionsFolder, "subdir3"), plan.args[3][0])
	})

	t.Run("returns plan if folder with multiple subfolders and files and pattern", func(t *testing.T) {
		tempDir := t.TempDir()
		actionsDir := filepath.Join(tempDir, actionsFolder)
		_ = os.Mkdir(actionsDir, 0755)
		_ = os.WriteFile(filepath.Join(actionsDir, "file.js"), []byte("test"), 0644)
		_ = os.Mkdir(filepath.Join(actionsDir, "subdir1"), 0755)
		_ = os.WriteFile(filepath.Join(actionsDir, "subdir1", "file1.js"), []byte("test"), 0644)
		_ = os.Mkdir(filepath.Join(actionsDir, "subdir2"), 0755)
		_ = os.WriteFile(filepath.Join(actionsDir, "subdir2", "file2.py"), []byte("test"), 0644)
		_ = os.Mkdir(filepath.Join(actionsDir, "subdir3"), 0755)

		// Expected args:
		// /actions file.js
		// /actions/subdir1 file1.js
		// /actions/subdir2
		// /actions/subdir3

		plan, err := buildCmdPlan(tempDir, cmd, "*.js")

		require.NoError(t, err)
		require.Len(t, plan.args, 4)
		require.Len(t, plan.args[0], 2)
		require.Len(t, plan.args[1], 2)
		require.Len(t, plan.args[2], 1)
		require.Len(t, plan.args[3], 1)
		require.Equal(t, filepath.Join(tempDir, actionsFolder), plan.args[0][0])
		require.Equal(t, "file.js", plan.args[0][1])
		require.Equal(t, filepath.Join(tempDir, actionsFolder, "subdir1"), plan.args[1][0])
		require.Equal(t, "file1.js", plan.args[1][1])
		require.Equal(t, filepath.Join(tempDir, actionsFolder, "subdir2"), plan.args[2][0])
		require.Equal(t, filepath.Join(tempDir, actionsFolder, "subdir3"), plan.args[3][0])
	})
}

func Test_filterFiles(t *testing.T) {

	testCases := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:    "returns all files if pattern is *",
			pattern: "*",
			expected: []string{
				"file1.js",
				"file2.py",
				"file3.go",
			},
		},
		{
			name:     "returns no files if pattern is empty",
			pattern:  "",
			expected: []string{},
		},
		{
			name:    "returns all .js files if pattern is *.js",
			pattern: "*.js",
			expected: []string{
				"file1.js",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files := []string{
				"file1.js",
				"file2.py",
				"file3.go",
			}

			filtered, err := filterFiles(files, tc.pattern)
			require.NoError(t, err)
			require.Equal(t, tc.expected, filtered)
		})
	}
}

func Test_checkActionsFolder(t *testing.T) {
	t.Run("returns error if actions folder does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := checkActionsFolderExists(tmpDir)
		require.Error(t, err)
	})

	t.Run("returns error if actions folder is not a folder", func(t *testing.T) {
		tmpDir := t.TempDir()
		actionsFile := filepath.Join(tmpDir, actionsFolder)
		_, err := os.Create(actionsFile)
		require.NoError(t, err)
		err = checkActionsFolderExists(tmpDir)
		require.Error(t, err)
	})

	t.Run("returns nil if actions folder exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		actionsFile := filepath.Join(tmpDir, actionsFolder)
		err := os.Mkdir(actionsFile, 0755)
		require.NoError(t, err)
		err = checkActionsFolderExists(tmpDir)
		require.NoError(t, err)
	})
}

func Test_getAllDirs(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	expectedDirs := []string{tempDir}

	// Create some subdirectories inside the temporary directory
	subDirs := []string{"dir1", "dir2", "dir3"}
	subSubDirs := []string{"subdir1", "subdir2", "subdir3"}
	for i, dir := range subDirs {
		tDir := filepath.Join(tempDir, dir)
		tSubDir := filepath.Join(tDir, subSubDirs[i])
		expectedDirs = append(expectedDirs, tDir)
		expectedDirs = append(expectedDirs, tSubDir)
		err := os.Mkdir(filepath.Join(tempDir, dir), 0755)

		require.NoError(t, err)
		err = os.Mkdir(tSubDir, 0755)

		require.NoError(t, err)
	}

	dirs, err := getAllDirs(tempDir)
	require.NoError(t, err)

	// Verify that the expected directories are present
	require.Equal(t, len(expectedDirs), len(dirs))

	require.ElementsMatch(t, expectedDirs, dirs)
}

func Test_getAllFiles(t *testing.T) {
	tempDir := t.TempDir()

	expectedFiles := []string{"file1", "file2", "file3"}

	// Create some files inside the temporary directory
	for _, file := range expectedFiles {
		f, err := os.Create(filepath.Join(tempDir, file))
		f.Close()
		require.NoError(t, err)
	}

	files, err := getAllFiles(tempDir)
	require.NoError(t, err)

	// Verify that the expected files are present
	require.Equal(t, len(expectedFiles), len(files))
	require.ElementsMatch(t, expectedFiles, files)
}
