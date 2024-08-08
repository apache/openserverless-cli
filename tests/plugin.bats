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
    export OPS_BRANCH="$(cat ../branch.txt)"
    run rm -rf ~/.ops/$OPS_BRANCH
}

@test "ops prints 'Plugins:'" {
    run ops -t
    assert_line 'Plugins:'
    assert_line "  plugin (local)"
}

@test "ops skips invalid plugin folders (without opsfile.yml)" {
    run mkdir olaris-test2
    run ops -t
    refute_line "  test2 (local)"
    run rm -rf olaris-test2
}

@test "ops help on sub cmds plugin" {
    run ops plugin sub
    assert_line '* opts:         opts test'
    assert_line '* simple:       simple'
}

@test "ops exec sub simple plugin cmd" {
    run ops plugin sub simple
    assert_line 'simple'
}

@test "original ops sub simple still works" {
    run ops sub simple
    assert_line 'simple'
}

@test "config in plugin opsroot is added with prefix" {
    run ops -config -d
    assert_line 'PLUGIN_KEY=value'
    assert_line 'PLUGIN_ANOTHER_KEY=a plugin value'
}

@test "other plugin without olaris is shown" {
    cd testdata
    run ops -update
    run ops -t
    assert_line 'Plugins:'
    assert_line "  other (local)"
}

@test "other sub simple prints simple" {
    cd testdata
    run ops -update
    run ops other sub simple
    assert_line 'simple'
}

@test "other tool runs ops tool" {
    cd testdata
    run ops -update
    run ops other tool
    assert_line 'hello'
}

@test "other command runs ops command" {
    cd testdata
    run ops -update
    run ops other command
    assert_success

}

# Plugin Tool Tests

@test "ops -plugin with wrong name" {
    run ops -plugin https://github.com/giusdp/olari
    assert_line "error: plugin repository must be a https url and plugin must start with 'olaris-'"
    assert_failure

    run ops -plugin olaris-test
    assert_line "error: plugin repository must be a https url and plugin must start with 'olaris-'"
    assert_failure
}

@test "ops -plugin with correct plugin repo" {
    run ops -plugin https://github.com/sciabarracom/olaris-test.git
    assert_success

    run ops -t
    assert_line 'Plugins:'
    assert_line "  plugin (local)"
    assert_line "  test (ops)"

    run rm -rf ~/.ops/olaris-test
}

@test "ops -plugin on existing plugin will update it" {
    run ops -plugin https://github.com/giusdp/olaris-test.git
    assert_success

    run ops -plugin https://github.com/giusdp/olaris-test.git
    assert_success
    assert_line "Updating plugin olaris-test"
    assert_line "The plugin repo is already up to date!"

    run rm -rf ~/.ops/olaris-test
}
