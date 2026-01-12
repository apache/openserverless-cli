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
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed embedded-tasks
var embeddedTasks embed.FS

// ExtractEmbeddedTasks extracts the embedded tasks to the specified directory.
// It returns the path to the extracted directory or an error.
func ExtractEmbeddedTasks() (string, error) {
	opsHome, err := getOpsHome()
	if err != nil {
		return "", err
	}

	targetDir := filepath.Join(opsHome, ".olaris")

	// Check if already exists
	if _, err := os.Stat(targetDir); err == nil {
		trace("Embedded tasks already extracted to", targetDir)
		return targetDir, nil
	}

	trace("Extracting embedded tasks to", targetDir)

	// Walk the embedded filesystem
	err = fs.WalkDir(embeddedTasks, "embedded-tasks", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from the root of the embedded folder
		relPath, err := filepath.Rel("embedded-tasks", path)
		if err != nil {
			return err
		}

		// Target path
		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			trace("Creating directory", targetPath)
			return os.MkdirAll(targetPath, 0755)
		}

		// Read the file from embedded FS
		data, err := embeddedTasks.ReadFile(path)
		if err != nil {
			return err
		}

		trace("Writing file", targetPath)
		return os.WriteFile(targetPath, data, 0644)
	})

	if err != nil {
		return "", err
	}

	return targetDir, nil
}

func getOpsHome() (string, error) {
	opsHome := os.Getenv("OPS_HOME")
	if opsHome == "" {
		// Use default
		opsHome = os.ExpandEnv("$HOME/.ops")
	}
	return opsHome, nil
}
