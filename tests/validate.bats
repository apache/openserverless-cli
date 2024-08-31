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

@test "ops validate help" {
    run ops -validate
    assert_failure
    assert_line --partial "ops -validate [-e] [-m | -n | -r <regex>] <value> [<message>]"
    assert_line "invalid number of arguments"

    run ops -validate -h
    assert_success
    assert_line --partial "ops -validate [-e] [-m | -n | -r <regex>] <value> [<message>]"
}

@test "ops validate email" {
    run ops -validate -m example@email.com
    assert_success

    run ops -validate -m example
    assert_line "validation failed"
    assert_failure
}

@test "ops validate number" {
    run ops -validate -n 123
    assert_success

    run ops -validate -n 123.456
    assert_success

    run ops -validate -n abc
    assert_line "validation failed"
    assert_failure
}

@test "ops validate with custom regex" {
    run ops -validate -r '^[a-z]+$' abc
    assert_success

    run ops -validate -r '^[a-z]+$' 123
    assert_line "validation failed"
    assert_failure
}

@test "ops validate on env vars" {
    run ops -validate -e -n TEST_ENV_VAR
    assert_line "variable 'TEST_ENV_VAR' not set"
    assert_failure

    export TEST_ENV_VAR=123
    run ops -validate -e -n TEST_ENV_VAR
    assert_success

    run ops -validate -e -m TEST_ENV_VAR
    assert_line "validation failed"
    assert_failure

    export TEST_ENV_VAR=example@gmail.com
    run ops -validate -e -m TEST_ENV_VAR
    assert_success

    export TEST_ENV_VAR=abc
    run ops -validate -e -r '^[a-z]+$' TEST_ENV_VAR
    assert_success

    export TEST_ENV_VAR=123
    run ops -validate -e -r '^[a-z]+$' TEST_ENV_VAR
    assert_line "validation failed"
    assert_failure
}

@test "ops validate with custom error message" {
    run ops -validate -m example@email.com "custom error message"
    assert_success

    run ops -validate -m abc "custom error message"
    assert_line "custom error message"
    assert_failure

    run ops -validate -n abc "custom error message"
    assert_line "custom error message"
    assert_failure

    run ops -validate -r '^[a-z]+$' 123 "custom error message"
    assert_line "custom error message"
    assert_failure
}
