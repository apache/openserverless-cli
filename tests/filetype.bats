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

@test "filetype usage print" {
    run nuv -filetype
    assert_line "Usage: filetype [-h] [-e] [-m] FILE"

    run nuv -filetype -h
    assert_line "Usage: filetype [-h] [-e] [-m] FILE"
}

@test "filetype -e" {
    run nuv -filetype -e testdata/sample.png
    assert_line "png"

    run nuv -filetype -e testdata/testfiletype.txt
    assert_line "bin"
}

@test "filetype -m" {
    run nuv -filetype -m testdata/sample.png
    assert_line "image/png"

    run nuv -filetype -m testdata/testfiletype.txt
    assert_line "applications/octet-stream"
}

@test "filetype -e -m" {
    run nuv -filetype -e -m testdata/sample.png
    assert_line "png image/png"

    run nuv -filetype -e -m testdata/testfiletype.txt
    assert_line "bin applications/octet-stream"
}

@test "filetype without flags" {
    run nuv -filetype testdata/sample.png
    assert_line "png image/png"

    run nuv -filetype testdata/testfiletype.txt
    assert_line "bin applications/octet-stream"
}
