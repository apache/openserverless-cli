#!/bin/sh
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
HERE=$PWD
SHVER=$(git ls-remote https://github.com/sciabarracom/sh | awk '/refs\/heads\/openserverless/{print $1}')
STAG="v3.38.0"
DTAG="v3.38.10"
cd task
git reset --hard
go clean -cache -modcache
git checkout "$STAG" -B openserverless
mkdir -p cmd/taskmain
cat cmd/task/task.go \
| sed -e 's/package main/package taskmain/' \
| sed -e 's/func main()/func _main()/' \
| tee cmd/taskmain/task.go
sed -i -e 's/func init/func FlagInit/' internal/flags/flags.go
cat   <<EOF >>cmd/taskmain/task.go

var todoFlagInit = true

func Task(_args []string) (int, error) {
	os.Args = _args
	if todoFlagInit {
		flags.FlagInit()
		todoFlagInit = false
	}
	if err := run(); err != nil {
		l := &logger.Logger{
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Verbose: flags.Verbose,
			Color:   flags.Color,
		}
		if err, ok := err.(*errors.TaskRunError); ok && flags.ExitCode {
			l.Errf(logger.Red, "%v\n", err)
			return err.TaskExitCode(), err
		}
		if err, ok := err.(errors.TaskError); ok {
			l.Errf(logger.Red, "%v\n", err)
			return err.Code(), err
		}
		l.Errf(logger.Red, "%v\n", err)
		return errors.CodeUnknown, err
	}
	return errors.CodeOk, nil
}
EOF
#cp $HERE/task.go cmd/taskmain/task.go
git add cmd/taskmain/task.go
find . -name \*.go  | while read file 
do echo $file 
   sed -i "s!go-task/task!sciabarracom/task!" $file
   sed -i 's!mvdan.cc/sh!github.com/sciabarracom/sh!' $file
   sed -i 's!"Taskfile.!"opsfile.!' $file
   sed -i 's!task: !ops: !' $file
done
sed -i -e 's/go-task\/task/sciabarracom\/task/' go.mod
sed -i -e '/mvdan.cc/g' go.mod
go get github.com/sciabarracom/sh/v3@$SHVER
go mod tidy
git commit -m "patching sh for ops" -a
go build
git tag $DTAG
git push origin-auth openserverless -f --tags
VER=$(git rev-parse HEAD)
cd ..
mkdir -p bin
GOBIN=$HERE/bin go install github.com/sciabarracom/task/v3/cmd/task@$VER 

