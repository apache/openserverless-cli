#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
HERE=$PWD
STAG="1.2.0"
DTAG="$STAG-openserverless"
cd ../tools/wskdeploy
git checkout "$STAG" -B openserverless

find . \( -name \*.go -o -name go.mod \) | while read file 
do echo $file 
   sed -i "s!apache/openwhisk-wskdeploy!sciabarracom/openwhisk-wskdeploy!" $file
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

