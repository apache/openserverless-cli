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

@test "nuv -wsk action is wrapped in nuv action" {
    run nuv action
    assert_line 'work with actions'
    assert_line 'Available Commands:'
}

@test "nuv -wsk activation is wrapped in nuv activation" {
    run nuv activation
    assert_line 'work with activations'
    assert_line 'Available Commands:'
}

@test "nuv -wsk package is wrapped in nuv package" {
    run nuv package
    assert_line 'work with packages'
    assert_line 'Available Commands:'
}

@test "nuv -wsk rule is wrapped in nuv rule" {
    run nuv rule
    assert_line 'work with rules'
    assert_line 'Available Commands:'
}

@test "nuv -wsk trigger is wrapped in nuv trigger" {
    run nuv trigger
    assert_line 'work with triggers'
    assert_line 'Available Commands:'
}

@test "nuv -wsk action -r is wrapped in nuv invoke" {
    run nuv invoke --apihost http://localhost:3233
    assert_line 'error: Invalid argument(s). An action name is required.'

    run nuv invoke --apihost http://localhost:3233 --help
    assert_line 'invoke action'
}

@test "nuv -wsk activation result is wrapped in nuv result" {
    run nuv result --apihost http://localhost:3233
    assert_line 'error: Invalid argument(s). An activation ID is required.'

    run nuv result --apihost http://localhost:3233 --help
    assert_line 'get the result of an activation'
}

@test "nuv -wsk activation logs is wrapped in nuv logs" {
    run nuv logs --apihost http://localhost:3233
    assert_line 'error: Invalid argument(s). An activation ID is required.'

    run nuv logs --apihost http://localhost:3233 --help
    assert_line 'get the logs of an activation'
}

@test "nuv -wsk action get --url is wrapped in nuv url" {
    run nuv url --apihost http://localhost:3233
    assert_line 'error: Invalid argument(s). An action name is required.'

    run nuv url --apihost http://localhost:3233 --help
    assert_line 'get action'
}
