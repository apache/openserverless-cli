#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
HERE=$PWD
SHVER=$(git ls-remote https://github.com/sciabarracom/sh | awk '/refs\/heads\/openserverless/{print $1}')
STAG="v3.38.0"
cd ../tools/task
git checkout "$STAG" -B openserverless
mkdir -p cmd/taskmain
cat cmd/task/task.go \
| sed -e 's/package main/package taskmain/' \
| sed -e 's/func main() {/func Task(_args []string) { os.Args = _args/' \
| tee cmd/taskmain/task.go
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
git push origin-auth openserverless -f
VER=$(git rev-parse HEAD)
GOBIN=$HERE go install github.com/sciabarracom/task/v3/cmd/task@$VER

