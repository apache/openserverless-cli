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

@test "help" {
    export TEST_VAR="envvar"
    run ops sub opts
    # just one as it is a cat of a message
    assert_line "Usage:"
    assert_line "  opts ciao <name>... [-c] [-e envvar]"
    run ops sub opts -h
    assert_line "Usage:"
    run ops sub opts --help
    assert_line "Usage:"
    # do not check the actual version but ensure the output is not the help test
    run ops sub opts --version
    refute_output "Usage:"
}

@test "cmd" {
    run ops sub opts hello
    assert_line "hello!"
}

@test "ciao" {
    run ops sub opts ciao mike
    assert_line "name: mike"
    assert_line "-c: no"

    run ops sub opts ciao mike miri -c
    assert_line "name: mike"
    assert_line "name: miri"
    assert_line "-c: yes"
}

@test "salve sayonara" {
    run ops sub opts salve aaa hi 1 2 --fl=ag
    assert_line "salve name=('aaa') hi x=1 y=2 --fl=ag"
    run ops sub opts sayonara opt1 10 20 --fa
    assert_line "sayonara=true opt1=true opt2=false x=10 y=20 --fa=true --fb=false"
}

@test "errors" {
    export TEST_VAR
    run ops sub opts salve
    assert_line "Usage:"
    assert_failure
    run ops sub opts salve opt4
    assert_line "Usage:"
    assert_failure
}


@test "shortening" {
    run ops s o c mike miri -c
    assert_line "ciao:"
    assert_line "name: mike"
    assert_line "name: miri"
    assert_line "-c: yes"

    run ops s o sal aaa hi 1 2 --fl=ag
    assert_line "salve name=('aaa') hi x=1 y=2 --fl=ag"
}

@test "bad shortening" {
    run ops f
    assert_failure
    assert_line --partial "error: ambiguous command: f."
}