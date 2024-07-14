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
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mitchellh/go-homedir"
)

func pluginTool() error {
	flag := flag.NewFlagSet("plugin", flag.ExitOnError)
	flag.Usage = printPluginUsage

	err := flag.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	if flag.NArg() != 1 {
		flag.Usage()
		return errors.New("invalid number of arguments. Expected 1")
	}

	return downloadPluginTasksFromRepo(flag.Arg(0))
}

func printPluginUsage() {
	fmt.Println(`Usage: nuv -plugin <repo-url>

Install/update plugins from a remote repository.
The name of the repository must start with 'olaris-'.`)
}

func downloadPluginTasksFromRepo(repo string) error {
	isNameValid, repoName := checkGitRepo(repo)
	if !isNameValid {
		return fmt.Errorf("plugin repository must be a https url and plugin must start with 'olaris-'")
	}

	pluginDir, err := homedir.Expand("~/.nuv/" + repoName)
	if err != nil {
		return err
	}

	if isDir(pluginDir) {
		fmt.Println("Updating plugin", repoName)

		r, err := git.PlainOpen(pluginDir)
		if err != nil {
			return err
		}
		// Get the working directory for the repository
		w, err := r.Worktree()
		if err != nil {
			return err
		}

		// Pull the latest changes from the origin remote and merge into the current branch
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil {
			if err.Error() == "already up-to-date" {
				fmt.Println("The plugin repo is already up to date!")
				return nil
			}
			return err
		}

		return nil
	}

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	// if not, clone
	cloneOpts := &git.CloneOptions{
		URL:           repo,
		Progress:      os.Stderr,
		ReferenceName: plumbing.NewBranchReferenceName("main"),
	}

	fmt.Println("Downloading plugins:", repoName)
	_, err = git.PlainClone(pluginDir, false, cloneOpts)
	if err != nil {
		return err
	}

	return nil
}

func checkGitRepo(url string) (bool, string) {
	// Remove the ".git" extension if present
	url = strings.TrimSuffix(url, ".git")

	// Extract the repository name from the URL
	parts := strings.Split(url, "/")
	repoName := parts[len(parts)-1]

	// Check if the repository name matches the pattern "https://...olaris-*"
	matchProtocol, _ := regexp.MatchString(`^https://.*$`, url)
	matchName, _ := regexp.MatchString(`^olaris-.*$`, repoName)

	if matchName && matchProtocol {
		return true, repoName
	}
	return false, ""
}

func printPluginsHelp() error {
	plgs, err := newPlugins()
	if err != nil {
		return err
	}
	plgs.print()
	return nil
}

// GetNuvRootPlugins returns the map with all the olaris-*/nuvroot.json files
// in the local and ~/.nuv folders, pointed by their plugin names.
// If the same plugin is found in both folders, the one in the local folder
// is used.
// Useful to build the config map including the plugin configs
func GetNuvRootPlugins() (map[string]string, error) {
	plgs, err := newPlugins()
	if err != nil {
		return nil, err
	}

	nuvRoots := make(map[string]string)
	for _, path := range plgs.local {
		name := getPluginName(path)
		nuvRootPath := joinpath(path, NUVROOT)
		nuvRoots[name] = nuvRootPath
	}

	for _, path := range plgs.nuv {
		name := getPluginName(path)
		// if the plugin is already in the map, skip it
		if _, ok := nuvRoots[name]; ok {
			continue
		}
		nuvRootPath := joinpath(path, NUVROOT)
		nuvRoots[name] = nuvRootPath
	}

	return nuvRoots, nil
}

// findTaskInPlugins returns the path to the plugin containing the task
// or an error if the task is not found
func findTaskInPlugins(plg string) (string, error) {
	plgs, err := newPlugins()
	if err != nil {
		return "", err
	}

	// check that plg is the suffix of a folder name in plgs.local
	for _, path := range plgs.local {
		folder := filepath.Base(path)
		if strings.TrimPrefix(folder, "olaris-") == plg {
			return path, nil
		}
	}

	// check that plg is the suffix of a folder name in plgs.nuv
	for _, path := range plgs.nuv {
		folder := filepath.Base(path)
		if strings.TrimPrefix(folder, "olaris-") == plg {
			return path, nil
		}
	}

	return "", &TaskNotFoundErr{input: plg}
}

// plugins struct holds the list of local and ~/.nuv olaris-* folders
type plugins struct {
	local []string
	nuv   []string
}

func newPlugins() (*plugins, error) {
	localDir := os.Getenv("NUV_ROOT_PLUGIN")
	localOlarisFolders := make([]string, 0)
	nuvOlarisFolders := make([]string, 0)

	// Search in directory (localDir/olaris-*)
	dir := filepath.Join(localDir, "olaris-*")
	olarisFolders, err := filepath.Glob(dir)
	if err != nil {
		return nil, err
	}

	// filter all folders that are do not contain nuvfile.yaml
	for _, folder := range olarisFolders {
		if !isDir(folder) || !exists(folder, NUVFILE) {
			continue
		}
		localOlarisFolders = append(localOlarisFolders, folder)
	}

	// Search in ~/.nuv/olaris-*
	nuvHome, err := homedir.Expand("~/.nuv")
	if err != nil {
		return nil, err
	}

	olarisNuvFolders, err := filepath.Glob(filepath.Join(nuvHome, "olaris-*"))
	if err != nil {
		return nil, err
	}
	for _, folder := range olarisNuvFolders {
		if !isDir(folder) || !exists(folder, NUVFILE) {
			continue
		}
		nuvOlarisFolders = append(nuvOlarisFolders, folder)
	}

	return &plugins{
		local: localOlarisFolders,
		nuv:   nuvOlarisFolders,
	}, nil
}

func (p *plugins) print() {
	if len(p.local) == 0 && len(p.nuv) == 0 {
		debug("No plugins installed")
		// fmt.Println("No plugins installed. Use 'nuv -plugin' to add new ones.")
		return
	}

	fmt.Println("Plugins:")
	if len(p.local) > 0 {
		for _, plg := range p.local {
			plgName := getPluginName(plg)
			fmt.Printf("  %s (local)\n", plgName)
		}
	}

	if len(p.nuv) > 0 {
		for _, plg := range p.nuv {
			plgName := getPluginName(plg)
			fmt.Printf("  %s (nuv)\n", plgName)
		}
	}
}

// getPluginName returns the plugin name from the plugin path, removing the
// olaris- prefix
func getPluginName(plg string) string {
	// remove olaris- prefix
	plgName := strings.TrimPrefix(filepath.Base(plg), "olaris-")
	return plgName

}
