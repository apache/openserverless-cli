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

@test "needupdate message" {
    run nuv -needupdate
    assert_line "Usage:"

    run nuv -needupdate -h
    assert_line "Usage:"
}

@test "needupdate returns 0 if a > b" {
    run nuv -needupdate 1.0.0 0.9.0
    assert_success

    run nuv -needupdate 1.0.0-dev.202312121700 1.0.0-dev.202312121600
    assert_success
}

@test "needupdate returns 1 if a < b" {
    run nuv -needupdate 0.9.0 1.0.0
    assert_failure

    run nuv -needupdate 1.0.0-dev.202312121600 1.0.0-dev.202312121700
    assert_failure
}

@test "needupdate returns 1 if a == b" {
    run nuv -needupdate 1.0.0 1.0.0
    assert_failure

    run nuv -needupdate 1.0.0-dev.202312121700 1.0.0-dev.202312121700
    assert_failure
}

@test "needupdate prints in stderr if parse fails" {
    run nuv -needupdate
    assert_failure

    run nuv -needupdate 1.0.0 1.0.wrong
    assert_failure
    assert_line "invalid semantic version: 1.0.wrong"
}
