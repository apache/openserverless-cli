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
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

func realpathTool() error {
	flags := flag.NewFlagSet("realpath", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println("Usage: nuv -realpath <path>")
		fmt.Println()
		fmt.Println("Options:")
		flags.PrintDefaults()
	}

	showHelp := flags.Bool("h", false, "Show help")

	// Parse command-line arguments
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *showHelp {
		flags.Usage()
		return nil
	}

	if flags.NArg() != 1 {
		flags.Usage()
		return errors.New("no path provided")
	}

	path, err := homedir.Expand(flags.Arg(0))
	if err != nil {
		return err
	}

	var absPath string
	if filepath.IsLocal(path) {
		absPath = filepath.Join(os.Getenv("NUV_PWD"), path)
	} else {
		absPath = path
	}

	absPath = filepath.Clean(absPath)
	fmt.Println(absPath)

	return nil
}
