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

@test "-echoif help message" {
    run ops -echoif
    assert_line "Usage: echoif <a> <b>"

    run ops -echoif -h
    assert_line "Usage: echoif <a> <b>"
}

@test "-echoif echoes a with successful previous command" {
    run echo "ok"
    run ops -echoif "a" "b"
    assert_line "a"
}

@test "-echoif echoes b with failed previous command" {
    run ops failing
    run ops -echoif "a" "b"
    refute_line "b"
}

@test "-echoifempty help message" {
    run ops -echoifempty
    assert_line "Usage: echoifempty <str> <a> <b>"

    run ops -echoifempty -h
    assert_line "Usage: echoifempty <str> <a> <b>"
}

@test "-echoifempty echoes a if string is empty" {
    run ops -echoifempty "" "a" "b"
    assert_line "a"
}

@test "-echoifempty echoes b if string is not empty" {
    run ops -echoifempty "not empty" "a" "b"
    assert_line "b"
}

@test "-echoifexists help message" {
    run ops -echoifexists
    assert_line "Usage: echoifexists <file> <a> <b>"

    run ops -echoifexists -h
    assert_line "Usage: echoifexists <file> <a> <b>"
}

@test "-echoifexists echoes a if file exists" {
    run ops -echoifexists "testdata/_1vars_" "a" "b"
    assert_line "a"
}

@test "-echoifexists echoes b if file does not exist" {
    run ops -echoifexists "testdata/_1vars_not_exists_" "a" "b"
    assert_line "b"
}