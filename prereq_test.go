package openserverless

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
)

func Example_execPrereqTask() {
	fmt.Println(execPrereqTask("bin", "bun"))
	// Output:
	// executing: ops [-task -d bin -t prereq.yml bun]
	// <nil>
}

func Example_loadPrereq() {
	//downloadPrereq("")
	os.Chdir(workDir)
	os.Chdir("tests")
	_, err := loadPrereq()
	fmt.Println(npath(err.Error()))
	os.Chdir("prereq")
	prq, err := loadPrereq()
	//fmt.Println(prq)
	fmt.Println(err, *prq.Tasks["bun"].Description, prq.Tasks["bun"].Vars["VERSION"])
	// Output:
	// not found /work/tests
	// <nil> bun v1.11.20
}

func Example_ensureBindir() {
	bindir, _ := homedir.Expand("~/.nuv/bin")
	os.RemoveAll(bindir)
	_, err1 := os.Stat(bindir)
	bindir1, _ := ensureBindir()
	_, err2 := os.Stat(bindir)
	fmt.Printf("no dir: %s\ncreated: %t\nyes dir: %v\n", nhpath(err1.Error()), bindir1 == bindir, err2)
	// Output:
	// no dir: stat /home/.nuv/bin: no such file or directory
	// created: true
	// yes dir: <nil>
}

func Example_touchAndClean() {
	bindir, _ := ensureBindir()
	os.RemoveAll(bindir)
	bindir, _ = ensureBindir()
	touch(bindir, "hello")
	err := touchAndClean(bindir, "hello", "1.2.3")
	fmt.Println(err, exists(bindir, "hello"), exists(bindir, "hello-1.2.3"), exists(bindir, "hello-1.2.4"))
	err = touchAndClean(bindir, "hello", "1.2.4")
	fmt.Println(err, exists(bindir, "hello"), exists(bindir, "hello-1.2.3"), exists(bindir, "hello-1.2.4"))
	// Output:
	// <nil> true true false
	// <nil> true false true
}

func Example_downloadPrereq() {
	bindir, _ := homedir.Expand("~/.nuv/bin")
	os.RemoveAll(bindir)
	os.Chdir(workDir)
	os.Chdir("tests")
	os.Chdir("prereq")

	prq, _ := loadPrereq()
	fmt.Println("1", downloadPrereq("bun", prq.Tasks["bun"]))
	fmt.Println("2", downloadPrereq("bun", prq.Tasks["bun"]))

	os.Chdir("other")
	prq, _ = loadPrereq()
	//fmt.Println(prq)
	//fmt.Println(PrereqSeenMap)
	fmt.Println("3", downloadPrereq("bun", prq.Tasks["bun"]))
	// Output:
	// downloading bun v1.11.20
	// 1 <nil>
	// 2 <nil>
	// 3 WARNING: bun prerequisite found twice with different versions!
	// Previous version: v1.11.20, ignoring v1.11.21
}

func Example_ensurePrereq() {
	bindir, _ := homedir.Expand("~/.nuv/bin")
	os.RemoveAll(bindir)
	os.Chdir(workDir)
	os.Chdir("tests")
	os.Chdir("prereq")
	fmt.Println(ensurePrereq())
	os.Chdir("sub")
	fmt.Println(ensurePrereq())
	// Output:
	// downloading bun v1.11.20
	// downloading coreutils 0.0.27
	// <nil>
	// <nil>
	// error in prereq bun: WARNING: bun prerequisite found twice with different versions!
	// Previous version: v1.11.20, ignoring v1.11.21
	// <nil>
}
