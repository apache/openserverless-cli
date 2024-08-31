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

@test "ops -random help" {
    run ops -random -h
    assert_success
    assert_line --partial "ops -random [options]"

    run ops -random --help
    assert_success
    assert_line --partial "ops -random [options]"
}

@test "ops -random" {
    run ops -random
    assert_success
    refute_line "Usage:"
}

@test "ops -random --int" {
    run ops -random --int
    assert_line "Usage:"

    run ops -random --int aa
    assert_line "Usage:"

    run ops -random --int 1 2 3
    assert_line "Usage:"

    run ops -random --int 1 aa
    assert_failure

    run ops -random --int -1
    assert_failure

    run ops -random --int 0
    assert_failure

    run ops -random --int 10 20
    assert_failure

    run ops -random --int 10 1
    assert_success
}

@test "ops -random -str" {
    run ops -random --str
    assert_line "Usage:"

    run ops -random --str aa
    assert_line "Usage:"

    run ops -random --str 1 2 3
    assert_line "Usage:"

    run ops -random --str -1
    assert_failure

    run ops -random --str 1 aa
    assert_success
}

@test "ops -random -u" {
    run ops -random -u
    assert_success

    run ops -random --uuid
    assert_success
}
    