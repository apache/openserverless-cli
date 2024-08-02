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

@test "ops sub invoked with prefix" {
    run ops sub
    # just one as it is a cat of a message
    assert_line "* opts:         opts test"
    assert_line "* simple:       simple"

    run ops s
    assert_line "* opts:         opts test"
    assert_line "* simple:       simple"
}

@test "ops sub simple invoked with prefixes" {
    run ops sub simple
    assert_line "ops: [simple] echo simple"
    assert_line "simple"

    run ops s simple
    assert_line "ops: [simple] echo simple"
    assert_line "simple"

    run ops s s
    assert_line "ops: [simple] echo simple"
    assert_line "simple"
}
