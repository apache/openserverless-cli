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
    ops -reset force
    cd ..
}

@test "help no download"  {

    #assert_line --partial "-reset complete"

    run ops
    assert_line "Welcome to ops, the all-mighty, extensibile apache OPenServerless CLI Tool." 

    run ops -h
    assert_line "Tools (use -<tool> -h for help):"
    refute_line unknown tool -t
    run ops -help
    assert_line "Tools (use -<tool> -h for help):"

    run ops -v 
    assert_line --partial "0.1.0" 
    run ops -version 
    assert_line --partial "0.1.0"
 
    run ops action --help
    assert_line  --partial "ops action [command]"
}


@test "help update" {
    run ops -update 2>/dev/null
    assert_success
    assert_line "Cloning tasks..."
    assert_line  "Tasks downloaded successfully"
    assert_line --partial "ensuring prerequisite coreutils"
}

@test "help with download" {

    run ops -update
    assert_success

    run ops -t
    assert_success
    assert_line --partial "OpenServerless Tasks:" 
    refute_line "Usage of experiments:" 
    run ops -tasks
    assert_line --partial "OpenWhisk Tasks:" 
    refute_line "Usage of experiments:" 

    run ops -i
    assert_success
    assert_line --partial OPS_VERSION: 0.1.0 
    assert_line "OPS_BRANCH: 0.1.0"
    run ops -info
    assert_line --partial OPS_VERSION: 0.1.0

    run ops -u
    assert_success
    assert_line "Updating tasks..."
    run ops -update 
    assert_success
    assert_line "Updating tasks..."

    run ops -l
    assert_failure
    assert_line "ops -login <apihost> [<user>]"
    run ops -login 
    assert_failure
    assert_line "ops -login <apihost> [<user>]"

    run ops -c
    assert_success
    assert_line "ops -config [options] [KEY | KEY=VALUE [KEY=VALUE ...]]"
    run ops -config
    assert_success
    assert_line "ops -config [options] [KEY | KEY=VALUE [KEY=VALUE ...]]"

}

