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

	"github.com/mattn/go-isatty"
	goja "github.com/nuvolaris/goja/gojamain"
)

func jsToolMain() error {
	flag := flag.NewFlagSet("js", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Println(`nuv -js

Usage: 
  nuv -js FILE.js

Interpret and run Javascript code.`)
	}

	// Parse command line flags
	err := flag.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// Check if input is from terminal and not from a pipe
	isTerminal := isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
	// if no input file and no input from pipe, print help message
	if isTerminal && flag.NArg() == 0 {
		flag.Usage()
		return nil
	}

	return goja.GojaMain()
}
