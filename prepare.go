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
	"log"
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
	repoURL := getOpsRepo()
	branch := getOpsBranch()
	opsDir, err := homedir.Expand("~/.ops")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(opsDir, 0755); err != nil {
		return "", err
	}

	opsBranchDir := joinpath(opsDir, branch)
	localDir, err := homedir.Expand(joinpath(opsBranchDir, "olaris"))
	if err != nil {
		return "", err
	}
	debug("localDir", localDir)

	// Updating existing tools
	if exists(opsBranchDir, "olaris") {
		trace("Updating olaris in", opsBranchDir)
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

		fmt.Println("Tasks updated successfully")
		touchLatestCheckFile(joinpath(opsBranchDir, LATESTCHECK))
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
		os.RemoveAll(opsBranchDir)
		warn(fmt.Sprintf("failed to clone olaris on branch '%s'", branch))
		return "", err
	}

	fmt.Println("Tasks downloaded successfully")

	createLatestCheckFile(opsBranchDir)

	// clone
	return localDir, nil
}

func pullTasks(force, silent bool) (string, error) {
	// download from github
	localDir, err := downloadTasksFromGitHub(force, silent)
	debug("localDir", localDir)
	if err != nil {
		return "", fmt.Errorf("cannot update tasks because: %s\nremove the folder ~/.ops and run ops -update", err.Error())
	}

	err = ensurePrereq(localDir)
	if err != nil {
		log.Fatalf("cannot download prerequisites: %v", err)
	}

	// validate OpsVersion semver against opsroot.json
	opsRoot, err := readOpsRootFile(localDir)
	if err != nil {
		return "", err
	}

	// check if the version is up to date
	opsVersion, err := semver.NewVersion(OpsVersion)
	if err != nil {
		// in development mode, we don't have a valid semver version
		warn("Unable to validate ops version", OpsVersion, ":", err)
		return localDir, nil
	}

	opsRootVersion, err := semver.NewVersion(opsRoot.Version)
	if err != nil {
		warn("Unable to validate opsroot.json version", opsRoot.Version, ":", err)
		return localDir, nil
	}

	// check if the version is up to date, if not warn the user
	if opsVersion.LessThan(opsRootVersion) {
		fmt.Println()
		fmt.Printf("Your ops version (%v) is older than the required version (%v).\n", opsVersion, opsRootVersion)
		if err := autoCLIUpdate(); err != nil {
			return "", err
		}
	}

	err = checkOperatorVersion(opsRoot.Config)
	if err == nil {
		fmt.Println()
		fmt.Println("New operator version detected!")
		fmt.Println("Current deployed operator can be updated with: ops update operator")
	}

	return localDir, nil
}

// locateOpsRoot locate the folder where starts execution
// it can be a parent folder of the current folder or it can be downloaded
// from github - it should contain a file opsfile.yml and a file opstools.yml in the root
func locateOpsRoot(cur string) (string, error) {
	cur, err := filepath.Abs(cur)
	if err != nil {
		return "", err
	}

	// search the root from here
	search := locateOpsRootSearch(cur)
	if search != "" {
		trace("found searching up:", search)
		return search, nil
	}

	// is there  olaris folder?
	olaris := joinpath(cur, "olaris")
	if exists(cur, "olaris") && exists(olaris, OPSFILE) && exists(olaris, OPSROOT) {
		trace("found sub olaris:", olaris)
		return olaris, nil
	}

	// is there an olaris folder in ~/.ops ?
	opsOlarisDir := fmt.Sprintf("~/.ops/%s/olaris", getOpsBranch())
	olaris, err = homedir.Expand(opsOlarisDir)
	if err == nil && exists(olaris, OPSFILE) && exists(olaris, OPSROOT) {
		trace("found sub", opsOlarisDir, ":", olaris)
		return olaris, nil
	}

	return "", fmt.Errorf("we cannot find opsfiles, download them with ops -update")
}

// locateOpsRootSearch search for `opsfiles.yml`
// and goes up looking for a folder with also `opsroot.json`
func locateOpsRootSearch(cur string) string {
	debug("locateOpsRootSearch:", cur)
	// exits opsfile.yml? if not, go up until you find it
	if !exists(cur, OPSFILE) {
		return ""
	}
	if exists(cur, OPSROOT) {
		return cur
	}
	parent := parent(cur)
	if parent == "" {
		return ""
	}
	return locateOpsRootSearch(parent)
}

func autoCLIUpdate() error {
	cli := os.Getenv("OPS_CMD")
	trace("autoCLIUpdate", cli)
	cmd := exec.Command(cli, "util", "update-cli")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkOperatorVersion(opsRootConfig map[string]interface{}) error {
	trace("checkOperatorVersion")
	images := opsRootConfig["images"].(map[string]interface{})
	operator := images["operator"].(string)
	opVer := strings.Split(operator, ":")[1]

	cmd := exec.Command(os.Getenv("OPS_CMD"), "util", "check-operator-version", opVer)
	return cmd.Run()
}

func setOpsOlarisHash(olarisDir string) error {
	trace("setOpsOlarisHash", olarisDir)
	r, err := git.PlainOpen(olarisDir)
	if err != nil {
		return err
	}
	h, err := r.Head()
	if err != nil {
		return err
	}
	debug("olaris hash", h.Hash().String())
	os.Setenv("OPS_OLARIS", h.Hash().String())
	trace("OPS_OLARIS", os.Getenv("OPS_OLARIS"))
	return nil
}
