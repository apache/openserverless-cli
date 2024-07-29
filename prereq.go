package openserverless

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

var PrereqSeenMap = map[string]string{}

// Define the Go structs
type Prereq struct {
	Version int                   `yaml:"version"`
	Tasks   map[string]PrereqTask `yaml:"tasks"`
}

type PrereqTask struct {
	Description *string           `yaml:"description,omitempty"` // Make description optional
	Vars        map[string]string `yaml:"vars,omitempty"`
}

// execute prereq task
func execPrereqTask(bindir string, name string) error {
	me, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{
		"-task",
		"-d", bindir,
		"-t", PREREQ,
		name,
	}
	if taskDryRun {
		fmt.Printf("invoking prereq for %s\n", name)
		return nil
	}
	trace("Exec:", me, args)
	err = exec.Command(me, args...).Run()
	if err != nil {
		return err
	}
	return nil
}

// load prerequisites in current dir
func loadPrereq() (Prereq, error) {
	var prereq Prereq = Prereq{}
	dir, err := os.Getwd()
	if err != nil {
		return prereq, err
	}

	if !exists(dir, PREREQ) {
		return prereq, fmt.Errorf("not found %s", dir)
	}
	trace("found prereq.yml in ", dir)

	dat, err := os.ReadFile(joinpath(dir, PREREQ))
	if err != nil {
		return prereq, err
	}

	err = yaml.Unmarshal(dat, &prereq)
	if err != nil {
		return prereq, err
	}

	return prereq, err
}

// ensure there is a bindir for downloading prerequisites
// read it from OPS_BIN and create it
// otherwise setup one in ~/nuv/<os>-<arch>/bin
// and sets OPS_BIN
func EnsureBindir() (string, error) {
	var err error = nil
	bindir := os.Getenv("OPS_BIN")
	if bindir == "" {
		bindir, err = homedir.Expand(fmt.Sprintf("~/.nuv/%s-%s/bin", runtime.GOOS, runtime.GOARCH))
		os.Setenv("OPS_BIN", bindir)
	}
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(bindir, 0755)
	if err != nil {
		return "", err
	}
	trace("bindir", bindir)
	return bindir, nil
}

// create a mark of current version touching <name>-<version> and remove all the other files starting with <name>-
func touchAndClean(dir string, name string, version string) error {

	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file starts with the prefix
		if !info.IsDir() && strings.HasPrefix(info.Name(), name+"-") {
			trace("Removing file:", path)
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	err = touch(dir, name+"-"+version)
	if err != nil {
		return err
	}
	return nil

}

// download a prerequisite
func downloadPrereq(name string, task PrereqTask) error {
	version, ok := task.Vars["VERSION"]
	if !ok {
		return nil
	}

	bindir, err := EnsureBindir()
	if err != nil {
		return err
	}

	// check if file and version exists
	vname := name + "-" + version
	trace("checking", vname, version)
	if exists(bindir, name) {
		// check if there is an inconsistency in the versions
		if exists(bindir, vname) {
			trace("already downloaded", vname)
			return nil
		}
		oldver, seen := PrereqSeenMap[name]
		if seen {
			if oldver == version {
				trace("same version again", vname)
				return nil
			}
			return fmt.Errorf("WARNING: %s prerequisite found twice with different versions!\nPrevious version: %s, ignoring %s", name, oldver, version)
		}
	}

	if taskDryRun {
		fmt.Printf("downloading %s %s\n", name, version)
		touch(bindir, name)
		touchAndClean(bindir, name, version)
	} else {
		execPrereqTask(bindir, name)
		// check if file and version exists
		if !exists(bindir, name) {
			return fmt.Errorf("failed to download %s %s", name, version)
		}
	}

	PrereqSeenMap[name] = version
	return touchAndClean(bindir, name, version)
}

// ensure prereq are satified looking at the prereq.yml
func ensurePrereq() error {
	// skip prereq - useful for tests
	if os.Getenv("OPS_NO_PREREQ") != "" {
		return nil
	}
	trace("ensurePrereq")
	prereq, err := loadPrereq()
	for task := range prereq.Tasks {
		err = downloadPrereq(task, prereq.Tasks[task])
		if err != nil {
			fmt.Printf("error in prereq %s: %v\n", task, err)
		}
	}
	return nil
}
