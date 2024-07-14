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

@test "nuv -login help" {
    run nuv -login
    assert_line "Usage:"
	assert_line "nuv login <apihost> [<user>]"
    assert_line "error: missing apihost"
	
    run nuv -login -h
    assert_line "Usage:"
    assert_line "nuv login <apihost> [<user>]"
}

@test "nuv -login with NUV_PASSWORD env does not prompt for password" {
    export NUV_PASSWORD=1234
    run nuv -login localhost
    refute_line "Enter Password:"
}

@test "nuv -login with NUV_USER env defines username" {
    export NUV_PASSWORD=1234
    export NUV_USER=foo
    run nuv -login localhost
    assert_line "Logging in as foo to localhost"
}

@test "nuv -login with NUV_USER and NUV_PASSWORD env" {
    export NUV_PASSWORD=1234
    export NUV_USER=foo
    run nuv -login localhost
    assert_line "Logging in as foo to localhost"
    refute_line "Enter Password:"
}

@test "nuv -login with NUV_APIHOST env" {
    export NUV_APIHOST=localhost
    export NUV_PASSWORD=1234
    run nuv -login
    assert_line "Logging in as nuvolaris to localhost"
}

@test "nuv -login with NUV_APIHOST and NUV_USER env" {
    export NUV_APIHOST=localhost
    export NUV_USER=foo
    export NUV_PASSWORD=1234
    run nuv -login
    assert_line "Logging in as foo to localhost"
}

@test "nuv -login with NUV_APIHOST, user is now first argument" {
    export NUV_APIHOST=localhost
    export NUV_PASSWORD=1234
    run nuv -login hello
    assert_line "Logging in as hello to localhost"
    refute_line "Enter Password:"
}