// Copyright (c) 2024, Michele Sciabarra  <msciabarra@apache.org>
// See LICENSE for licensing information

package interp

import (
	"context"
	"fmt"

	"github.com/sciabarracom/sh/v3/syntax"
)

func isCoreutil(name string) bool {
	switch name {
	case
		"arch", "b2sum", "b3sum", "base32", "base64", "basename", "basenc", "cat", "chgrp", "chmod", "chown", "chroot",
		"cksum", "comm", "cp", "csplit", "cut", "date", "dd", "df", "dir", "dircolors", "dirname", "du", "echo", "env", "expand",
		"expr", "factor", "false", "fmt", "fold", "groups", "hashsum", "head", "hostid", "hostname", "id", "install", "join",
		"kill", "link", "ln", "logname", "ls", "md5sum", "mkdir", "mkfifo", "mknod", "mktemp", "more", "mv", "nice", "nl",
		"nohup", "nproc", "numfmt", "od", "paste", "pathchk", "pinky", "pr", "printenv", "printf", "ptx", "pwd", "readlink",
		"realpath", "rm", "rmdir", "seq", "sha1sum", "sha224sum", "sha256sum", "sha3-224sum", "sha3-256sum", "sha3-384sum",
		"sha3-512sum", "sha384sum", "sha3sum", "sha512sum", "shake128sum", "shake256sum", "shred", "shuf", "sleep", "sort",
		"split", "stat", "stdbuf", "sum", "sync", "tac", "tail", "tee", "test", "timeout", "touch", "tr", "true", "truncate",
		"tsort", "tty", "uname", "unexpand", "uniq", "unlink", "uptime", "users", "vdir", "wc", "who", "whoami", "yes",
		"coreutils":
		return true
	}
	return false

}

func isBuiltin(name string) bool {

	if isCoreutil(name) {
		return true
	}
	return _isBuiltin(name)
}

func (r *Runner) builtinCode(ctx context.Context, pos syntax.Pos, name string, args []string) int {

	if isCoreutil(name) {

		cmd := []string{"coreutils"}
		if name != "coreutils" {
			cmd = append(cmd, name)
		}
		args = append(cmd, args...)
		fmt.Printf("%v\n", args)
		// TODO: Consider unix.Exec, i.e. actually replacing
		// the process. It's in theory what a shell should do,
		// but in practice it would kill the entire Go process
		// and it's not available on Windows.
		r.exitShell(ctx, 1)
		r.exec(ctx, args)
		return r.exit
	}

	return _builtinCode(ctx, pos, name, args)
}
