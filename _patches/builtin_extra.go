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

// Copyright (c) 2024, Michele Sciabarra  <msciabarra@apache.org>
// See LICENSE for licensing information

package interp

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sciabarracom/sh/v3/syntax"
)

func isCoreutil(name string) bool {
	coreUtils := os.Getenv("OPS_COREUTILS")
	if coreUtils == "" {
		return false
	}
	coreUtils = " coreutils " + coreUtils + " "
	return strings.Contains(coreUtils, " "+name+" ")
}

func isTools(name string) bool {
	tools := os.Getenv("OPS_TOOLS")
	if tools == "" {
		return false
	}
	tools = " " + tools + " "
	return strings.Contains(tools, " "+name+" ")
}

func isBuiltin(name string) bool {
	if isCoreutil(name) {
		return true
	}
	if isTools(name) {
		return true
	}
	return isBuiltin_orig(name)
}

func (r *Runner) builtinCode(ctx context.Context, pos syntax.Pos, name string, args []string) int {

	if isCoreutil(name) {
		cmd := []string{"coreutils"}
		if name != "coreutils" {
			cmd = append(cmd, name)
		}
		args = append(cmd, args...)
		if os.Getenv("TRACE") != "" {
			fmt.Printf("%v\n", args)
		}
		r.exec(ctx, args)
		return r.exit
	}

	if isTools(name) {
		args = append([]string{os.Getenv("OPS_CMD"), "-" + name}, args...)
		if os.Getenv("TRACE") != "" {
			fmt.Printf("%v\n", args)
		}
		r.exec(ctx, args)
		return r.exit
	}
	return r.builtinCode_orig(ctx, pos, name, args)
}
