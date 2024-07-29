package openserverless

import (
	"fmt"
	"os"
)

func Example_execPrereqTask() {
	fmt.Println(execPrereqTask("bin", "bun"))
	// Output:
	// invoking prereq for bun
	// <nil>
}

func Example_loadPrereq() {
	//downloadPrereq("")
	dir := joinpath(workDir, "tests")
	_, err := loadPrereq(dir)
	fmt.Println(npath(err.Error()))
	prq, err := loadPrereq(joinpath(dir, "prereq"))
	//fmt.Println(prq)
	fmt.Println(err, *prq.Tasks["bun"].Description, prq.Tasks["bun"].Vars["VERSION"])
	// Output:
	// not found /work/tests
	// <nil> bun v1.11.20
}

func Example_ensureBindir() {
	bindir, _ := EnsureBindir()
	os.RemoveAll(bindir)
	_, err1 := os.Stat(bindir)
	bindir1, _ := EnsureBindir()
	_, err2 := os.Stat(bindir)
	fmt.Printf("no dir:%s\ncreated: %t\nyes dir: %v\n", after(":", err1.Error()), bindir1 == bindir, err2)
	// Output:
	// no dir: no such file or directory
	// created: true
	// yes dir: <nil>
}

func Example_touchAndClean() {
	bindir, _ := EnsureBindir()
	os.RemoveAll(bindir)
	bindir, _ = EnsureBindir()
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
	bindir, _ := EnsureBindir()
	os.RemoveAll(bindir)

	prqdir := joinpath(joinpath(workDir, "tests"), "prereq")
	prq, _ := loadPrereq(prqdir)
	fmt.Println("1", downloadPrereq("bun", prq.Tasks["bun"]))
	fmt.Println("2", downloadPrereq("bun", prq.Tasks["bun"]))

	prq, _ = loadPrereq(joinpath(prqdir, "sub"))
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
	bindir, _ := EnsureBindir()
	os.RemoveAll(bindir)
	dir := joinpath(joinpath(workDir, "tests"), "prereq")
	fmt.Println(ensurePrereq(dir))
	fmt.Println(ensurePrereq(joinpath(dir, "sub")))
	// Output:
	// downloading bun v1.11.20
	// downloading coreutils 0.0.27
	// <nil>
	// error in prereq bun: WARNING: bun prerequisite found twice with different versions!
	// Previous version: v1.11.20, ignoring v1.11.21
	// <nil>
}
