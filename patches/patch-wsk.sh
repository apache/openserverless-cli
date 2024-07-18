#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
HERE=$PWD
STAG="v3.38.0"
DTAG="$STAG-openserverless"
cd ../tools/task
git checkout "$TAG" -B openserverless
mkdir -p cmd/taskmain
cat cmd/task/task.go \
| sed -e 's/package main/package taskmain/' \
| sed -e 's/func main() {/func Task(_args []string) { os.Args = _args/' \
| tee cmd/taskmain/task.go
cp $HERE/task.go cmd/taskmain/task.go
git add cmd/taskmain/task.go
find . \( -name \*.go -o -name go.mod \) | while read file 
do echo $file 
   sed -i "s!go-task/task!sciabarracom/task!" $file
   sed -i 's!mvdan.sh/cc!sciabarracom/sh!' $file
   sed -i 's!"Taskfile.!"opsfile.!' $file
   sed -i 's!task: !ops: !' $file
done
go clean -modcache -cache -testcache -fuzzcache
go mod tidy
git commit -m "patching sh for ops" -a
git push origin-auth openserverless -f
git tag -d $DTAG
git tag $DTAG
git push origin-auth --delete $DTAG
git push origin-auth $DTAG
GOBIN=$HERE go install github.com/sciabarracom/task/v3/cmd/task@$DTAG

