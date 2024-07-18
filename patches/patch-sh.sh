#!/bin/sh
cd "$(dirname $0)"
# any suggestion how to avoid this rename and use just replaces in go.mod is welcome
go clean -cache -modcache -testcache
HERE=$PWD
STAG="v3.8.0"
DTAG="$STAG-openserverless"
cd ../tools/sh
git checkout "$STAG" -B openserverless
cp $HERE/builtin_extra.go interp/builtin_extra.go
git add interp/builtin_extra.go
sed  -i -e 's!) builtinCode(!) _builtinCode(!' -e 's!func isBuiltin!func _isBuiltin!' interp/builtin.go
find . \( -name \*.go -o -name go.mod \) | while read file 
do echo $file 
   sed -i 's!mvdan.cc/sh!github.com/sciabarracom/sh!' $file 
done
go clean -modcache -cache -testcache -fuzzcache
go mod tidy
git commit -m "patching sh for ops" -a
git push origin-auth openserverless -f
git tag -d $DTAG
git tag $DTAG
git push origin-auth --delete $DTAG
git push origin-auth $DTAG
GOBIN=$HERE go install github.com/sciabarracom/sh/v3/cmd/gosh@$DTAG

#echo now push it git push origin-auth openserverless

