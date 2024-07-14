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

@test "-awk" {
    run nuv -awk
    assert_line "usage: goawk [-F fs] [-v var=value] [-f progfile | 'prog'] [file ...]"
    assert_failure
}

@test "-awk -h" {
    run nuv -awk -h
    assert_line "usage: goawk [-F fs] [-v var=value] [-f progfile | 'prog'] [file ...]"
    assert_line "Standard AWK arguments:"
    assert_line "Additional GoAWK features:"
    assert_line "GoAWK debugging arguments:"
}

@test "-awk print $ 1 file" {
    run nuv -awk '{print $1}' testdata/awk_test.txt
    assert_line "This"
    assert_line "This"
    assert_line "This"
    assert_success
}

@test "-awk replace" {
    run echo "$(echo "Hello World" | nuv -awk '{$2="Nuvolaris"; print $0}')"
    assert_line "Hello Nuvolaris"
    assert_success
}

@test "file not found" {
    run nuv -awk '{print $1}' testdata/no_tests.txt
    assert_line "file \"testdata/no_tests.txt\" not found"
    assert_failure
}
