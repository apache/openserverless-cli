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

@test "-jj help" {
    run nuv -jj -h
    assert_line "usage: jj [-v value] [-purnOD] [-i infile] [-o outfile] keypath"
    assert_success
}

@test "-jj get string example" {
    run echo "$(echo '{"name":{"first":"Tom","last":"Smith"}}' | nuv -jj name.last)"
    assert_output "Smith"
    assert_success
}

@test "-jj get block of JSON" {
    run echo "$(echo '{"name":{"first":"Tom","last":"Smith"}}' | nuv -jj name)"
    assert_line '{"first":"Tom","last":"Smith"}'
    assert_success
}

@test "-jj get raw string value" {
    run echo "$(echo '{"name":{"first":"Tom","last":"Smith"}}' | nuv -jj -r name.last)"
    assert_line '"Smith"'
    assert_success
}

@test "-jj get array value by index" {
    run echo "$(echo '{"friends":["Tom","Jane","Carol"]}' | nuv -jj friends.1)"
    assert_line "Jane"
    assert_success
}
