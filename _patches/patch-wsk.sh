#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
HERE=$PWD
STAG="1.2.0"
cd ../tools/wsk
#git reset --hard
git checkout "$STAG" -B openserverless
cp $HERE/i18n_resources.go wski18n/i18n_resources.go
sed -i '/wski18n\/i18n_resources.go/d' .gitignore
git add wski18n/i18n_resources.go
sed -i -e 's/"wsk"/"ops -wsk"/' commands/wsk.go
find . \( -name \*.go -o -name go.mod \) | while read file 
do echo $file 
   sed -i 's!apache/openwhisk-cli/!sciabarracom/openwhisk-cli/!' $file
   sed -i 's!apache/openwhisk-cli"!sciabarracom/openwhisk-cli"!' $file
   sed -i 's!apache/openwhisk-cli !sciabarracom/openwhisk-cli !' $file
   sed -i 's!apache/openwhisk-cli$!sciabarracom/openwhisk-cli!' $file
   sed -i 's!apache/openwhisk-wskdeploy!sciabarracom/openwhisk-wskdeploy!' $file
done
sed -i '/openwhisk-wskdeploy/d' go.mod
DEPLOYVER=$(git ls-remote https://github.com/sciabarracom/openwhisk-wskdeploy | awk '/refs\/heads\/openserverless/{print $1}')
go get github.com/sciabarracom/openwhisk-wskdeploy@$DEPLOYVER
go mod tidy
git commit -m "patching sh for ops" -a
git push origin-auth openserverless -f
go clean -modcache -cache -testcache -fuzzcache
VER=$(git rev-parse HEAD)
GOBIN=$HERE go install github.com/sciabarracom/openwhisk-cli@$VER

