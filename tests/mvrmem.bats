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
    rm -f empty_file
}

@test "-rename -remove" {
    run nuv -rename
    assert_line "Usage: rename <source> <destination>"

    run nuv -rename missing 
    assert_line "Usage: rename <source> <destination>"

    run nuv -rename missing something
    assert_failure
    assert_line "rename missing something: no such file or directory" 

    touch something
    run nuv -rename something somethingelse
    assert_success
    assert_line "renamed something -> somethingelse"

    run nuv -rename somethingelse /dev/null
    assert_failure
    assert_line "rename somethingelse /dev/null: invalid cross-device link"

    run nuv -remove
    assert_line "Usage: remove <filename>"

    run nuv -remove missing
    assert_failure
    assert_line "remove missing: no such file or directory"

    run nuv -remove somethingelse
    assert_success
    assert_line "removed somethingelse"
}    

@test "-empty" {
    run ops -empty
    assert_line "Usage: filename"
    run ops -empty empty_file
    assert test -f empty_file
    assert_success
    run ops -empty empty_file
    assert_failure
    assert_line "file already exists"
    rm empty_file

}