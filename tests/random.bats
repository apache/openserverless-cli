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

@test "nuv -random help" {
    run nuv -random -h
    assert_success
    assert_line "Usage:"

    run nuv -random --help
    assert_success
    assert_line "Usage:"
}

@test "nuv -random" {
    run nuv -random
    assert_success
    refute_line "Usage:"
}

@test "nuv -random --int" {
    run nuv -random --int
    assert_line "Usage:"

    run nuv -random --int aa
    assert_line "Usage:"

    run nuv -random --int 1 2 3
    assert_line "Usage:"

    run nuv -random --int 1 aa
    assert_failure

    run nuv -random --int -1
    assert_failure

    run nuv -random --int 0
    assert_failure

    run nuv -random --int 10 20
    assert_failure

    run nuv -random --int 10 1
    assert_success
}

@test "nuv -random -str" {
    run nuv -random --str
    assert_line "Usage:"

    run nuv -random --str aa
    assert_line "Usage:"

    run nuv -random --str 1 2 3
    assert_line "Usage:"

    run nuv -random --str -1
    assert_failure

    run nuv -random --str 1 aa
    assert_success
}

@test "nuv -random -u" {
    run nuv -random -u
    assert_success

    run nuv -random --uuid
    assert_success
}
    