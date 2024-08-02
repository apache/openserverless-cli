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

@test "ops -wsk action is wrapped in ops action" {
    run ops action
    assert_line 'work with actions'
    assert_line 'Available Commands:'
}

@test "ops -wsk activation is wrapped in ops activation" {
    run ops activation
    assert_line 'work with activations'
    assert_line 'Available Commands:'
}

@test "ops -wsk package is wrapped in ops package" {
    run ops package
    assert_line 'work with packages'
    assert_line 'Available Commands:'
}

@test "ops -wsk rule is wrapped in ops rule" {
    run ops rule
    assert_line 'work with rules'
    assert_line 'Available Commands:'
}

@test "ops -wsk trigger is wrapped in ops trigger" {
    run ops trigger
    assert_line 'work with triggers'
    assert_line 'Available Commands:'
}

@test "ops -wsk action -r is wrapped in ops invoke" {
    run ops invoke --apihost http://localhost:3233
    assert_line 'error: Invalid argument(s). An action name is required.'

    run ops invoke --apihost http://localhost:3233 --help
    assert_line 'invoke action'
}

@test "ops -wsk activation result is wrapped in ops result" {
    run ops result --apihost http://localhost:3233
    assert_line 'error: Invalid argument(s). An activation ID is required.'

    run ops result --apihost http://localhost:3233 --help
    assert_line 'get the result of an activation'
}

@test "ops -wsk activation logs is wrapped in ops logs" {
    run ops logs --apihost http://localhost:3233
    assert_line 'error: Invalid argument(s). An activation ID is required.'

    run ops logs --apihost http://localhost:3233 --help
    assert_line 'get the logs of an activation'
}

@test "ops -wsk action get --url is wrapped in ops url" {
    run ops url --apihost http://localhost:3233
    assert_line 'error: Invalid argument(s). An action name is required.'

    run ops url --apihost http://localhost:3233 --help
    assert_line 'get action'
}
