#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
HERE=$PWD
STAG="1.2.0"
DTAG="v1.2.1"
cd wskdeploy
git checkout "$STAG" -B openserverless
cp $HERE/runtimes_extra.go runtimes/runtimes_extra.go
git add runtimes/runtimes_extra.go
sed -i -e 's/func ParseOpenWhisk/func _ParseOpenWhisk/' runtimes/runtimes.go
sed -i -e 's/"unset"/"openserverless"/g' main.go
find . \( -name \*.go -o -name go.mod \) | while read file 
do echo $file 
   sed -i 's!apache/openwhisk-wskdeploy!sciabarracom/openwhisk-wskdeploy!' $file
done
go clean -modcache -cache -testcache -fuzzcache
go mod tidy
git commit -m "patching sh for ops" -a
git tag $DTAG
git push origin-auth openserverless -f --tags
VER=$(git rev-parse HEAD)
GOBIN=$HERE go install github.com/sciabarracom/openwhisk-wskdeploy@$VER

