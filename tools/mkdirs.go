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
)

func Mkdirs() error {
	flag.Usage = printMkdirsHelp
	// Define command line flags
	helpFlag := flag.Bool("h", false, "Print help message")
	parentFlag := flag.Bool("p", false, "Create parent directories")

	// Parse command line flags
	flag.Parse()

	// Print help message if -h flag is provided
	if *helpFlag {
		flag.Usage()
		return nil
	}

	// Get the list of directories to create from the remaining command line arguments
	dirs := flag.Args()
	if len(dirs) == 0 {
		flag.Usage()
	}

	// Create each directory, with or without parent directories
	for _, dir := range dirs {
		if *parentFlag {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		} else {
			err := os.Mkdir(dir, 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func printMkdirsHelp() {
	fmt.Println("Usage: mkdir [-h] [-p] DIRECTORY...")
	fmt.Println("Create one or more directories.")
}
