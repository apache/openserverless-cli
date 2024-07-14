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

@test "nuv validate help" {
    run nuv -validate
    assert_line "Usage:"

    run nuv -validate -h
    assert_line "Usage:"
    assert_line "nuv -validate [-e] [-m | -n | -r <regex>] <value> [<message>]"
}

@test "nuv validate email" {
    run nuv -validate -m example@email.com
    assert_success

    run nuv -validate -m example
    assert_line "validation failed"
    assert_failure
}

@test "nuv validate number" {
    run nuv -validate -n 123
    assert_success

    run nuv -validate -n 123.456
    assert_success

    run nuv -validate -n abc
    assert_line "validation failed"
    assert_failure
}

@test "nuv validate with custom regex" {
    run nuv -validate -r '^[a-z]+$' abc
    assert_success

    run nuv -validate -r '^[a-z]+$' 123
    assert_line "validation failed"
    assert_failure
}

@test "nuv validate on env vars" {
    run nuv -validate -e -n TEST_ENV_VAR
    assert_line "variable 'TEST_ENV_VAR' not set"
    assert_failure

    export TEST_ENV_VAR=123
    run nuv -validate -e -n TEST_ENV_VAR
    assert_success

    run nuv -validate -e -m TEST_ENV_VAR
    assert_line "validation failed"
    assert_failure

    export TEST_ENV_VAR=example@gmail.com
    run nuv -validate -e -m TEST_ENV_VAR
    assert_success

    export TEST_ENV_VAR=abc
    run nuv -validate -e -r '^[a-z]+$' TEST_ENV_VAR
    assert_success

    export TEST_ENV_VAR=123
    run nuv -validate -e -r '^[a-z]+$' TEST_ENV_VAR
    assert_line "validation failed"
    assert_failure
}

@test "nuv validate with custom error message" {
    run nuv -validate -m example@email.com "custom error message"
    assert_success

    run nuv -validate -m abc "custom error message"
    assert_line "custom error message"
    assert_failure

    run nuv -validate -n abc "custom error message"
    assert_line "custom error message"
    assert_failure

    run nuv -validate -r '^[a-z]+$' 123 "custom error message"
    assert_line "custom error message"
    assert_failure
}
