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

@test "datefmt print usage" {
    run nuv -datefmt -h
    assert_success
    assert_line "Usage:"

    run nuv -datefmt --help
    assert_success
    assert_line "Usage:"
}

@test "datefmt with input timestamp and output format" {
    run nuv -datefmt -t 1577836800 -f DateOnly
    assert_success
    assert_output "2020-01-01"
}

@test "datefmt string date with in fmt" {
    run nuv -datefmt -s "2023-01-01" --if DateOnly
    assert_success
    assert_output "Sun Jan  1 00:00:00 UTC 2023"
}

@test "datefmt with errors" {
    # missing input format
    run nuv -datefmt -s "2023-01-01"
    assert_failure
    assert_output "error: both --str and --if must be provided. Only str given: 2023-01-01"

    # missing input string
    run nuv -datefmt --if DateOnly
    assert_failure
    assert_output "error: both --str and --if must be provided. Only input format given: DateOnly"

    # wrong input format
    run nuv -datefmt -s "2023-01-01" --if NotAFormat
    assert_failure
    assert_output "error: invalid input format: NotAFormat"

    # wrong output format
    run nuv -datefmt -t 1577836800 -f NotAFormat
    assert_failure
    assert_output "error: invalid output format: NotAFormat"
} 