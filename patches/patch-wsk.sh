#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
HERE=$PWD
STAG="1.2.0"
DTAG="v$STAG-openserverless"
cd ../tools/wsk
git reset --hard
git checkout "$STAG" -B openserverless
cp $HERE/i18n_resources.go wski18n/i18n_resources.go
sed -i '/wski18n\/i18n_resources.go/d' .gitignore
git add wski18n/i18n_resources.go
find . \( -name \*.go -o -name go.mod \) | while read file 
do echo $file 
   sed -i 's!apache/openwhisk-cli/!sciabarracom/openwhisk-cli/!' $file
   sed -i 's!apache/openwhisk-cli"!sciabarracom/openwhisk-cli"!' $file
   sed -i 's!apache/openwhisk-cli !sciabarracom/openwhisk-cli !' $file
   sed -i 's!apache/openwhisk-cli$!sciabarracom/openwhisk-cli!' $file
   sed -i 's!apache/openwhisk-wskdeploy!sciabarracom/openwhisk-wskdeploy!' $file
done
go clean -modcache -cache -testcache -fuzzcache

sed -i '/openwhisk-wskdeploy/d' go.mod
sed -i '/openwhisk-client-go/d' go.mod

go get github.com/sciabarracom/openwhisk-wskdeploy@$DTAG
go get github.com/apache/openwhisk-client-go@$STAG

go mod tidy
git commit -m "patching sh for ops" -a
git push origin-auth openserverless -f

git tag -d $DTAG
git tag $DTAG
git push origin-auth --delete $DTAG
git push origin-auth $DTAG
git push origin-auth openserverless


# misterious bug here - it is building apache/openwhisk-cli instead
#echo $CUR
#go clean -modcache -cache -testcache -fuzzcache
#CUR=$(git rev-parse $DTAG)
#GOBIN=$HERE go install github.com/sciabarracom/openwhisk-cli@$DTAG

