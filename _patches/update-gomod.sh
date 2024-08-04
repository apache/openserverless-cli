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

cd "$(dirname $0)"/..
#git ls-remote https://github.com/sciabarracom/task | awk '/refs\/heads\/openserverless/{print $1}' >_task.hash
#git ls-remote https://github.com/sciabarracom/openwhisk-cli | awk '/refs\/heads\/openserverless/{print $1}' >_wsk.hash
#go get github.com/sciabarracom/task/v3@$(cat _task.hash)
#go get github.com/sciabarracom/openwhisk-cli@$(cat _wsk.hash)
#go mod tidy


go clean -cache -modcache
git ls-remote https://github.com/sciabarracom/sh | awk '/refs\/heads\/openserverless/{print $1}' | tee _sh.hash
go get github.com/sciabarracom/sh/v3@$(cat _sh.hash)

git ls-remote https://github.com/sciabarracom/task | awk '/refs\/heads\/openserverless/{print $1}' | tee _task.hash
go get github.com/sciabarracom/task/v3@$(cat _task.hash)

git ls-remote https://github.com/sciabarracom/openwhisk-wskdeploy | awk '/refs\/heads\/openserverless/{print $1}' | tee _wskdeploy.hash
go get github.com/sciabarracom/openwhisk-wskdeploy@$(cat _wskdeploy.hash)

git ls-remote https://github.com/sciabarracom/openwhisk-cli | awk '/refs\/heads\/openserverless/{print $1}' | tee _wsk.hash

go get github.com/sciabarracom/openwhisk-cli@$(cat _wsk.hash)

go mod tidy
