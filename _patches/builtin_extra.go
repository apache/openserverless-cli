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
	return IsBuiltin1(name)
}

func (r *Runner) builtinCode(ctx context.Context, pos syntax.Pos, name string, args []string) int {

	if isCoreutil(name) {
		cmd := []string{"coreutils"}
		if name != "coreutils" {
			cmd = append(cmd, name)
		}
		args = append(cmd, args...)
		fmt.Printf("%v\n", args)
		r.exitShell(ctx, 1)
		r.exec(ctx, args)
		return r.exit
	}

	if isTools(name) {
		args = append([]string{"ops", "-" + name}, args...)
		fmt.Printf("%v\n", args)
		r.exitShell(ctx, 1)
		r.exec(ctx, args)
		return r.exit

	}
	return r.BuiltinCode1(ctx, pos, name, args)
}
