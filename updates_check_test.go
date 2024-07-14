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
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
)

func changeLatestCheckTime(base string, t time.Duration) {
	latestCheckPath := joinpath(base, ".latestcheck")
	file, err := os.Stat(latestCheckPath)
	if err != nil {
		pr("failed to get latest_check file info", err)
	}
	mtime := file.ModTime()
	mtime = mtime.Add(t)
	err = os.Chtimes(latestCheckPath, mtime, mtime)
	if err != nil {
		pr("failed to set latest_check file mtime", err)
	}
}

func resetOneCommit(repo *git.Repository) {
	commIter, _ := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})

	if _, err := commIter.Next(); err != nil {
		pr("failed to get first commit", err)
	}
	secondLastCommit, err := commIter.Next()
	if err != nil {
		pr("failed to get second last commit", err)
	}

	w, _ := repo.Worktree()
	if err := w.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: secondLastCommit.Hash,
	}); err != nil {
		pr("failed to reset repo", err)
	}

}

func Example_checkUpdated_uptodate() {
	// clone olaris folder into a temp folder
	tmpDir, err := os.MkdirTemp("", "nuv-test")
	if err != nil {
		pr("failed to create temp dir", err)
	}
	defer os.RemoveAll(tmpDir)
	tmpDirBranch := joinpath(tmpDir, getNuvBranch())
	olarisTmpPath := joinpath(tmpDirBranch, "olaris")

	_, _ = git.PlainClone(olarisTmpPath, false, &git.CloneOptions{
		URL: getNuvRepo(),
	},
	)

	// run checkUpdated and check if it creates the latest_check file
	createLatestCheckFile(tmpDirBranch)

	if exists(tmpDirBranch, ".latestcheck") {
		pr("latest_check file created")
	}

	// change latest_check file mtime to 2 seconds ago
	changeLatestCheckTime(tmpDirBranch, -2*time.Second)

	// re-run checkUpdated and check output "Tasks up to date!"
	checkUpdated(tmpDir, 1*time.Second)

	// Output:
	// latest_check file created
	// Checking for updates...
	// Tasks up to date!
}

func Example_checkUpdated_outdated() {
	// clone olaris folder into a temp folder
	tmpDir, err := os.MkdirTemp("", "nuv-test")
	if err != nil {
		pr("failed to create temp dir", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpDirBranch := joinpath(tmpDir, getNuvBranch())
	olarisTmpPath := joinpath(tmpDirBranch, "olaris")

	repo, _ := git.PlainClone(olarisTmpPath, false, &git.CloneOptions{
		URL: getNuvRepo(),
	},
	)

	// run checkUpdated and check if it creates the latest_check file
	createLatestCheckFile(tmpDirBranch)

	if exists(tmpDirBranch, ".latestcheck") {
		pr("latest_check file created")
	}

	// change latest_check file mtime to 2 seconds ago
	changeLatestCheckTime(tmpDirBranch, -2*time.Second)

	// git reset olaris to a previous commit
	resetOneCommit(repo)

	// re-run checkUpdated and check output
	checkUpdated(tmpDir, 1*time.Second)

	// Output:
	// latest_check file created
	// Checking for updates...
	// New tasks available! Use 'nuv -update' to update.
}
