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
	"fmt"
	"os"
	"strings"
)

func Executable() (int, error) {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println(MarkdownHelp("executable"))
		return 0, nil
	}

	file := args[0]

	// Get the current file permissions
	info, err := os.Stat(file)
	if err != nil {
		return 1, err
	}

	// Add execute permissions for the owner
	if GetOS() == "windows" {
		if !strings.HasSuffix(strings.ToLower(file), ".exe") {
			fileexe := file + ".exe"
			err = os.Rename(file, fileexe)
			if err != nil {
				return 1, err
			}
			fmt.Printf("Successfully renamed %s to %s\n", file, fileexe)
			return 0, nil
		} else {
			fmt.Println("Nothing to do")
			return 0, nil
		}
	} else {
		err = os.Chmod(file, info.Mode()|0100)
		if err != nil {
			return 1, err
		}
		fmt.Printf("Successfully added execute permissions to %s\n", file)
		return 0, nil
	}
}
