#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
go clean -cache -modcache -testcache
HERE=$PWD
STAG="v3.8.0"
DTAG="v3.8.2"
cd sh
git checkout "$STAG" -B openserverless
cp $HERE/builtin_extra.go interp/builtin_extra.go
git add interp/builtin_extra.go
sed  -i -e 's!) builtinCode(!) BuiltinCode1(!' -e 's!func isBuiltin!func IsBuiltin1!' interp/builtin.go
find . \( -name \*.go -o -name go.mod \) | while read file 
do echo $file 
   sed -i 's!mvdan.cc/sh!github.com/sciabarracom/sh!' $file 
done
go clean -modcache -cache -testcache -fuzzcache
go mod tidy
git commit -m "patching sh for ops" -a
git tag $DTAG
git push -f origin-auth openserverless --tags
VER=$(git rev-parse --short HEAD)
GOBIN=$HERE go install github.com/sciabarracom/sh/v3/cmd/gosh@$VER

