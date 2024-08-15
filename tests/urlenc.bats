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

@test "ops -urlenc help" {
    run ops -urlenc -h
    assert_success
    assert_line --partial "ops -urlenc [-e] [-s <string>] [parameters]"

    run ops -urlenc --help
    assert_success
    assert_line --partial "ops -urlenc [-e] [-s <string>] [parameters]"
}

@test "ops -urlenc" {
  run ops -urlenc
  assert_success
  refute_output

  run ops -urlenc a=1 b=2
  assert_success
  assert_line "a%3D1&b%3D2"

  run ops -urlenc -s "|" a=1 b=2
  assert_success
  assert_line "a%3D1|b%3D2"
}

@test "ops -urlenc string from env" {
  run env TEST=ops ops -urlenc -e TEST
  assert_line "ops"
}
