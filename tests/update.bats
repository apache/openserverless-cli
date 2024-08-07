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

setup() {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
    export NO_COLOR=1
    ops -reset force
    cd ..
}

@test "ops -update and versions" {
    run ops -update
    assert_success
    assert_line "Tasks downloaded successfully"
    assert_line --partial "ensuring prerequisite coreutils"

    OPS_VERSION=0.0.0 OPS_SKIP_UPDATE_CLI=1 run ops -update
    assert_line --partial "Your ops version (0.0.0) is older than the required version"
    assert_success
    assert_line "Trying to update ops..."
    assert_line --partial "skipping rename"

    OPS_VERSION=10.2.3 run ops -update
    assert_line "Tasks are already up to date!"
    assert_success
}

@test "ops -update with bad version" {
    OPS_VERSION=notsemver run ops -update
    assert_line "Unable to validate ops version notsemver : Invalid Semantic Version"
    assert_success
}

@test "ops -update on branch" {
    OPS_BRANCH=main run ops -update
    assert_line "Tasks downloaded successfully"
    assert test -d ~/.ops/main
    assert_success
}
