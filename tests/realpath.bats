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

@test "-realpath help message" {
    run nuv -realpath
    assert_line "Usage: nuv -realpath <path>"

    run nuv -realpath -h
    assert_line "Usage: nuv -realpath <path>"
}

@test "-realpath with relative path" {
    WD=$(pwd)
    run nuv -realpath ./test
    assert_success
    assert_output "${WD}/test"
}

@test "-realpath with absolute path" {
    WD=$(pwd)
    run nuv -realpath "${WD}/test"
    assert_success
    assert_output "${WD}/test"

    run nuv -realpath "${WD}/./test"
    assert_success
    assert_output "${WD}/test"
}