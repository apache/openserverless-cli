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
    export LANG=en_US.UTF-8
    export LANGUAGE=en_US.UTF-8
    export LC_ALL=en_US.UTF-8
    rm -rf ~/.ops
    ops -update
    cd ~/.ops/$OPS_BRANCH/olaris
    ops -info
}

@test "ops -update on olaris with old commit updates correctly" {
    run git reset --hard HEAD~1
    run git status
    assert_line --partial "Your branch is behind"

    run ops -update
    assert_line --partial "Tasks updated successfully"
    assert_success

    run git status
    assert_line --partial "Your branch is up to date with" 
}
