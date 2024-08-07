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

@test "ops -login help" {
    run ops -login
    assert_line "Usage:"
	assert_line "ops -login <apihost> [<user>]"
    assert_line "error: missing apihost"
	
    run ops -login -h
    assert_line "Usage:"
    assert_line "ops -login <apihost> [<user>]"
}

@test "ops -login with OPS_PASSWORD env does not prompt for password" {
    export OPS_PASSWORD=1234
    run ops -login nuvolaris.dev
    assert_line --partial "Logging in https://nuvolaris.dev"
    refute_line "Enter Password:"
}

@test "ops -login with OPS_USER env defines username" {
    export OPS_PASSWORD=1234
    export OPS_USER=foo
    run ops -login  http://localhost
    assert_failure
    assert_line "Logging in http://localhost as foo"
}

@test "ops -login with OPS_USER and OPS_PASSWORD env" {
    export OPS_PASSWORD=1234
    export OPS_USER=foo
    run ops -login localhost
    assert_line "Logging in http://localhost as foo"
    refute_line "Enter Password:"
}

@test "ops -login with OPS_APIHOST env" {
    export OPS_APIHOST=localhost
    export OPS_PASSWORD=1234
    unset OPS_USER
    run ops -login
    assert_failure
    assert_line "Logging in http://localhost as nuvolaris"
}

@test "ops -login with OPS_APIHOST and OPS_USER env" {
    export OPS_APIHOST=localhost
    export OPS_USER=foo
    export OPS_PASSWORD=1234
    run ops -login
    assert_line "Logging in http://localhost as foo"
}
