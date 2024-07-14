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

@test "nuv sub invoked with prefix" {
    run nuv sub
    # just one as it is a cat of a message
    assert_line "* opts:         opts test"
    assert_line "* simple:       simple"

    run nuv s
    assert_line "* opts:         opts test"
    assert_line "* simple:       simple"
}

@test "nuv sub simple invoked with prefixes" {
    run nuv sub simple
    assert_line "task: [simple] echo simple"
    assert_line "simple"

    run nuv s simple
    assert_line "task: [simple] echo simple"
    assert_line "simple"

    run nuv s s
    assert_line "task: [simple] echo simple"
    assert_line "simple"
}
