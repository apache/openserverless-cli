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
}

@test "first run auto setups" {
    run rm -rf ~/.nuv
    run nuv
    assert_success
    assert_output --partial "Welcome to nuv! Setting up..."
    run ls ~/.nuv
    assert_success
}

@test "wrong branch fails to setup" {
    run rm -rf ~/.nuv
    export NUV_BRANCH=wrong
    run nuv
    assert_failure
    assert_output --partial "Welcome to nuv! Setting up..."
    assert_output --partial "failed to clone olaris on branch 'wrong'"
}

@test "correct branch setups" {
    run rm -rf ~/.nuv
    export NUV_BRANCH=3.0.0-testing
    run nuv
    assert_success
    assert_output --partial "Welcome to nuv! Setting up..."
    run ls ~/.nuv
    assert_success
    assert_output --partial "3.0.0-testing"
}
