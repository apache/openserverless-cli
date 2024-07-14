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
    export NUV_BRANCH=3.0.0-testing
    rm -rf ~/.nuv
    run nuv -update
    cd ~/.nuv/3.0.0-testing/olaris
}

@test "nuv -update on olaris with old commit updates correctly" {
    run git reset --hard HEAD~1
    run git status
    assert_line --partial "Your branch is behind 'origin/3.0.0-testing'"

    run nuv -update
    assert_line "Nuvfiles updated successfully"
    assert_success

    run git status
    assert_line "Your branch is up to date with 'origin/3.0.0-testing'."
}
