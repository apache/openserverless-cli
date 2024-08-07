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
    cd testdata
    ops -reset force
}

@test "first run auto setups" {
    run ops
    assert_success
    assert_output --partial "Welcome to ops"
    run ls ~/.ops
    assert_success
}

@test "wrong branch fails to setup" {
    export OPS_BRANCH=wrong
    run ops -info
    assert_failure
    assert_output --partial "Welcome to ops! Setting up..."
    assert_output --partial "failed to clone olaris on branch 'wrong'"
}

@test "correct branch setups" {
    export OPS_BRANCH=0.1.0
    run ops -info
    assert_success
    assert_output --partial "Welcome to ops! Setting up..."
    assert_output --partial "OPS_BRANCH: 0.1.0"
    run ls ~/.ops
    assert_success
}
