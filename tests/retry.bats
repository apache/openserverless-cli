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

@test "ops -retry help" {
    run ops -retry
    assert_success
    assert_line "Usage:"

    run ops -retry -h
    assert_success
    assert_line "Usage:"

    run ops -retry --help
    assert_success
    assert_line "Usage:"
}

@test "ops -retry fail" {
    run ops -retry -t 0 ops failing
    assert_line "error: failure after 0 retries or 60 seconds."
    assert_failure

    run ops -retry -t 0 -v ops failing
    assert_line "Retry Parameters: max time=60 seconds, retries=0 times."
    assert_line "error: failure after 0 retries or 60 seconds."
    assert_failure

    run ops -retry -t 5 -m 2 ops failing
    assert_line "error: failure after 5 retries or 2 seconds."
    assert_failure
}

@test "ops -retry succeed" {
    run ops -retry -t 1 -m 5 ops fail_then_succeed
    assert_success

    run ops -retry -t 1 -m 5 -v ops fail_then_succeed
    assert_success
}
