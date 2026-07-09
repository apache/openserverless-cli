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

func OpsPath() (int, error) {
	flags := flag.NewFlagSet("opspath", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println(MarkdownHelp("opspath"))
	}

	showHelp := flags.Bool("h", false, "Show help")

	// Parse command-line arguments
	if err := flags.Parse(os.Args[1:]); err != nil {
		return 1, err
	}

	if *showHelp {
		flags.Usage()
		return 0, nil
	}

	if flags.NArg() != 1 {
		flags.Usage()
		return 1, errors.New("error: no path provided")
	}

	inputPath := filepath.Clean(filepath.FromSlash(flags.Arg(0))) // ensure correct OS separator

	if filepath.IsAbs(inputPath) {
		fmt.Println(inputPath)
		return 0, nil
	}

	expandedPath, err := homedir.Expand(inputPath)
	if err != nil {
		return 1, err
	}

	if filepath.IsAbs(expandedPath) {
		fmt.Println(expandedPath)
		return 0, nil
	}

	fullPath, err := filepath.Abs(filepath.Join(os.Getenv("OPS_PWD"), expandedPath))
	if err != nil {
		return 1, err
	}

	fmt.Println(fullPath)

	return 0, nil
}
