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
go clean -cache -modcache -testcache
HERE=$PWD
STAG="v3.8.0"
DTAG="v3.8.3"
cd sh
git reset --hard
git checkout "$STAG" -B openserverless
cp $HERE/builtin_extra.go interp/builtin_extra.go
git add interp/builtin_extra.go
sed  -i -e 's!) builtinCode(!) builtinCode_orig(!' -e 's!func isBuiltin!func isBuiltin_orig!' interp/builtin.go
find . \( -name \*.go -o -name go.mod \) | while read file 
do echo $file 
   sed -i 's!mvdan.cc/sh!github.com/sciabarracom/sh!' $file 
done
go clean -modcache -cache -testcache -fuzzcache
go mod tidy

#echo >>go.mod "replace github.com/sciabarracom/sh/v3/interp => ./interp" 
#echo >>go.mod "replace github.com/sciabarracom/sh/v3/syntax => ./syntax"
#go build ./cmd/gosh
#
#exit

git commit -m "patching sh for ops" -a
git tag $DTAG
git push -f origin-auth openserverless --tags
VER=$(git rev-parse --short HEAD)
GOBIN=$HERE go install github.com/sciabarracom/sh/v3/cmd/gosh@$VER
OPS_CMD=$(which ops) OPS_TOOLS="jq awk" OPS_COREUTILS="ls cp mv" TRACE=1 ./gosh
