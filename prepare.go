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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mitchellh/go-homedir"
)

func downloadTasksFromGitHub(force bool, silent bool) (string, error) {
	debug("Download tasks from github")
	repoURL := getNuvRepo()
	branch := getNuvBranch()
	nuvDir, err := homedir.Expand("~/.nuv")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(nuvDir, 0755); err != nil {
		return "", err
	}

	nuvBranchDir := joinpath(nuvDir, branch)
	localDir, err := homedir.Expand(joinpath(nuvBranchDir, "olaris"))
	if err != nil {
		return "", err
	}
	debug("localDir", localDir)

	// Updating existing tools
	if exists(nuvBranchDir, "olaris") {
		trace("Updating olaris in", nuvBranchDir)
		fmt.Println("Updating tasks...")
		r, err := git.PlainOpen(localDir)
		if err != nil {
			return "", err
		}
		// Get the working directory for the repository
		w, err := r.Worktree()
		if err != nil {
			return "", err
		}

		// Pull the latest changes from the origin remote and merge into the current branch
		// Clone the repo if not existing
		ref := plumbing.NewBranchReferenceName(branch)
		err = w.Pull(&git.PullOptions{
			RemoteName:    "origin",
			ReferenceName: ref,
			SingleBranch:  true,
		})
		if err != nil {
			if err.Error() == "already up-to-date" {
				fmt.Println("Tasks are already up to date!")
				return localDir, nil
			}
			return "", err
		}

		fmt.Println("Nuvfiles updated successfully")
		touchLatestCheckFile(joinpath(nuvBranchDir, LATESTCHECK))
		return localDir, nil
	}

	// Clone the repo if not existing
	ref := plumbing.NewBranchReferenceName(branch)
	cloneOpts := &git.CloneOptions{
		URL:           repoURL,
		Progress:      os.Stderr,
		ReferenceName: ref, // Specify the branch to clone
	}

	fmt.Println("Cloning tasks...")
	_, err = git.PlainClone(localDir, false, cloneOpts)
	if err != nil {
		os.RemoveAll(nuvBranchDir)
		warn(fmt.Sprintf("failed to clone olaris on branch '%s'", branch))
		return "", err
	}

	fmt.Println("Nuvfiles downloaded successfully")

	createLatestCheckFile(nuvBranchDir)

	// clone
	return localDir, nil
}

func pullTasks(force, silent bool) (string, error) {
	// download from github
	localDir, err := downloadTasksFromGitHub(force, silent)
	debug("localDir", localDir)
	if err != nil {
		return "", err
	}

	// validate NuvVersion semver against nuvroot.json
	nuvRoot, err := readNuvRootFile(localDir)
	if err != nil {
		return "", err
	}

	// check if the version is up to date
	nuvVersion, err := semver.NewVersion(NuvVersion)
	if err != nil {
		// in development mode, we don't have a valid semver version
		warn("Unable to validate nuv version", NuvVersion, ":", err)
		return localDir, nil
	}

	nuvRootVersion, err := semver.NewVersion(nuvRoot.Version)
	if err != nil {
		warn("Unable to validate nuvroot.json version", nuvRoot.Version, ":", err)
		return localDir, nil
	}

	// check if the version is up to date, if not warn the user
	if nuvVersion.LessThan(nuvRootVersion) {
		fmt.Println()
		fmt.Printf("Your nuv version (%v) is older than the required version in nuvroot.json (%v).\n", nuvVersion, nuvRootVersion)
		fmt.Println("Attempting to update nuv...")
		if err := autoCLIUpdate(); err != nil {
			return "", err
		}
	}

	err = checkOperatorVersion(nuvRoot.Config)
	if err == nil {
		fmt.Println()
		fmt.Println("New operator version detected!")
		fmt.Println("Current deployed operator can be updated with: nuv update operator")
	}

	return localDir, nil
}

// locateNuvRoot locate the folder where starts execution
// it can be a parent folder of the current folder or it can be downloaded
// from github - it should contain a file nuvfile.yml and a file nuvtools.yml in the root
func locateNuvRoot(cur string) (string, error) {
	cur, err := filepath.Abs(cur)
	if err != nil {
		return "", err
	}

	// search the root from here
	search := locateNuvRootSearch(cur)
	if search != "" {
		trace("found searching up:", search)
		return search, nil
	}

	// is there  olaris folder?
	olaris := joinpath(cur, "olaris")
	if exists(cur, "olaris") && exists(olaris, NUVFILE) && exists(olaris, NUVROOT) {
		trace("found sub olaris:", olaris)
		return olaris, nil
	}

	// is there an olaris folder in ~/.nuv ?
	nuvOlarisDir := fmt.Sprintf("~/.nuv/%s/olaris", getNuvBranch())
	olaris, err = homedir.Expand(nuvOlarisDir)
	if err == nil && exists(olaris, NUVFILE) && exists(olaris, NUVROOT) {
		trace("found sub", nuvOlarisDir, ":", olaris)
		return olaris, nil
	}

	// is there an olaris folder in NUV_BIN?
	nuvBin := os.Getenv("NUV_BIN")
	if nuvBin != "" {
		olaris = joinpath(nuvBin, "olaris")
		if exists(olaris, NUVFILE) && exists(olaris, NUVROOT) {
			trace("found sub NUV_BIN olaris:", olaris)
			return olaris, nil
		}
	}

	return "", fmt.Errorf("we cannot find nuvfiles, download them with nuv -update")
}

// locateNuvRootSearch search for `nuvfiles.yml`
// and goes up looking for a folder with also `nuvroot.json`
func locateNuvRootSearch(cur string) string {
	debug("locateNuvRootSearch:", cur)
	// exits nuvfile.yml? if not, go up until you find it
	if !exists(cur, NUVFILE) {
		return ""
	}
	if exists(cur, NUVROOT) {
		return cur
	}
	parent := parent(cur)
	if parent == "" {
		return ""
	}
	return locateNuvRootSearch(parent)
}

func autoCLIUpdate() error {
	trace("autoCLIUpdate")
	cmd := exec.Command("nuv", "update", "cli")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkOperatorVersion(nuvRootConfig map[string]interface{}) error {
	trace("checkOperatorVersion")
	images := nuvRootConfig["images"].(map[string]interface{})
	operator := images["operator"].(string)
	opVer := strings.Split(operator, ":")[1]

	cmd := exec.Command("nuv", "util", "check-operator-version", opVer)
	return cmd.Run()
}

func setNuvOlarisHash(olarisDir string) error {
	trace("setNuvOlarisHash", olarisDir)
	r, err := git.PlainOpen(olarisDir)
	if err != nil {
		return err
	}
	h, err := r.Head()
	if err != nil {
		return err
	}
	debug("olaris hash", h.Hash().String())
	os.Setenv("NUV_OLARIS", h.Hash().String())
	trace("NUV_OLARIS", os.Getenv("NUV_OLARIS"))
	return nil
}
