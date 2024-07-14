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
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func echoIfTool() error {
	flags := flag.NewFlagSet("echoif", flag.ContinueOnError)
	flags.Usage = printEchoIfUsage

	showHelp := flags.Bool("h", false, "Show help")

	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *showHelp {
		flags.Usage()
		return nil
	}

	if flags.NArg() != 2 {
		flags.Usage()
		return nil
	}

	a := flags.Arg(0)
	b := flags.Arg(1)

	cmd := exec.Command("sh", "-c", "echo $?")
	exitCode, err := cmd.Output()
	if err != nil {
		return err
	}

	// Trim any newline characters from the exit code
	exitCode = exitCode[:len(exitCode)-1]

	// Check if the exit code is 0 and print the corresponding value
	if string(exitCode) == "0" {
		fmt.Println(a)
	} else {
		fmt.Println(b)
	}
	return nil
}

func echoIfEmptyTool() error {
	flags := flag.NewFlagSet("echoifempty", flag.ContinueOnError)
	flags.Usage = printEchoIfEmptyUsage

	showHelp := flags.Bool("h", false, "Show help")
	if *showHelp {
		flags.Usage()
		return nil
	}
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	if flags.NArg() != 3 {
		flags.Usage()
		return nil
	}

	str := flags.Arg(0)
	if str == "" {
		fmt.Println(flags.Arg(1))
	} else {
		fmt.Println(flags.Arg(2))
	}

	return nil
}

func echoIfExistsTool() error {
	flags := flag.NewFlagSet("echoifexists", flag.ContinueOnError)
	flags.Usage = printEchoIfExistsUsage

	showHelp := flags.Bool("h", false, "Show help")
	if *showHelp {
		flags.Usage()
		return nil
	}
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	if flags.NArg() != 3 {
		flags.Usage()
		return nil
	}

	file := flags.Arg(0)

	_, err := os.Stat(file)
	if err == nil {
		fmt.Println(flags.Arg(1))
		return nil
	}

	if os.IsNotExist(err) {
		fmt.Println(flags.Arg(2))
		return nil
	}

	return err
}

func printEchoIfUsage() {
	fmt.Println(`Usage: echoif <a> <b>

echoif is a utility that echoes the value of <a> if the exit code of the previous command is 0, 
echoes the value of <b> otherwise`)
}

func printEchoIfEmptyUsage() {
	fmt.Println(`Usage: echoifempty <str> <a> <b>

echoifempty is a utility that echoes the value of <a> if <str> is empty, echoes the value of <b> otherwise`)
}

func printEchoIfExistsUsage() {
	fmt.Println(`Usage: echoifexists <file> <a> <b>

echoifexists is a utility that echoes the value of <a> if <file> exists, echoes the value of <b> otherwise`)
}
