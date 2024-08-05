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

// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

// gosh is a proof of concept shell built on top of [interp].
package tools

import (
	"context"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"

	"github.com/sciabarracom/sh/v3/interp"
	"github.com/sciabarracom/sh/v3/syntax"
)

func Sh() (int, error) {
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
		if path == "-h" || path == "--help" {
			fmt.Println("Sh is the mvdan shell using the ops environment.\nUsage: [<script>|-h|--help]\nWithout args starts an interactive shell otherwise execute the script.")
			return 0, nil
		}
	}
	err := runAll(path)
	if e, ok := interp.IsExitStatus(err); ok {
		return int(e), err
	}
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func runAll(path string) error {
	r, err := interp.New(interp.StdIO(os.Stdin, os.Stdout, os.Stderr))
	if err != nil {
		return err
	}

	if path != "" {
		return runPath(r, path)
	}
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return runInteractive(r, os.Stdin, os.Stdout, os.Stderr)
	}
	return run(r, os.Stdin, "")
}

func run(r *interp.Runner, reader io.Reader, name string) error {
	prog, err := syntax.NewParser().Parse(reader, name)
	if err != nil {
		return err
	}
	r.Reset()
	ctx := context.Background()
	return r.Run(ctx, prog)
}

func runPath(r *interp.Runner, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return run(r, f, path)
}

func runInteractive(r *interp.Runner, stdin io.Reader, stdout, stderr io.Writer) error {
	parser := syntax.NewParser()
	fmt.Fprintf(stdout, "@ ")
	var runErr error
	fn := func(stmts []*syntax.Stmt) bool {
		if parser.Incomplete() {
			fmt.Fprintf(stdout, "> ")
			return true
		}
		ctx := context.Background()
		for _, stmt := range stmts {
			runErr = r.Run(ctx, stmt)
			if r.Exited() {
				return false
			}
		}
		fmt.Fprintf(stdout, "@ ")
		return true
	}
	if err := parser.Interactive(stdin, fn); err != nil {
		return err
	}
	return runErr
}
