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

	"github.com/h2non/filetype"
)

func Filetype() error {
	flag.Usage = printFiletypeHelp

	// Define command line flags
	helpFlag := flag.Bool("h", false, "Print help message")
	extensionFlag := flag.Bool("e", false, "Show file standard extension")
	mimeFlag := flag.Bool("m", false, "Show file mime type")

	// Parse command line flags
	flag.Parse()

	// Print help message if -h flag is provided
	if *helpFlag {
		flag.Usage()
		return nil
	}

	// Get the file path from the remaining command line arguments
	files := flag.Args()
	if len(files) != 1 {
		flag.Usage()
		return nil
	}

	file := files[0]
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	kind, err := filetype.Match(fileContent)
	if err != nil {
		return err
	}

	var extension string
	if kind == filetype.Unknown {
		extension = "bin"
	} else {
		extension = kind.Extension
	}

	var mime string
	if kind == filetype.Unknown {
		mime = "applications/octet-stream"
	} else {
		mime = kind.MIME.Value
	}

	// if both flags missing or both present, print ext and mime
	if (!*extensionFlag && !*mimeFlag) || (*extensionFlag && *mimeFlag) {
		fmt.Println(extension, mime)
		return nil
	}

	if *extensionFlag {
		fmt.Println(extension)
	} else {
		fmt.Println(mime)
	}

	return nil
}

func printFiletypeHelp() {
	fmt.Println("Usage: filetype [-h] [-e] [-m] FILE")
	fmt.Println("Show extensione and MIME type of a file.")
	fmt.Println(" -h  shows this help")
	fmt.Println(" -e  show file standard extension")
	fmt.Println(" -m  show file mime type")
}
