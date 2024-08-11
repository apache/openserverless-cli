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

@test "legacy docops without markdown"  {
    # show the legacy docopts.txt
    run ops docopts legacy
    assert_success
    assert_line --partial "legacy simple [-f]"
    run ops docopts legacy simple
    assert_line --partial "simple without flag"    
    run ops docopts legacy simple -f
    assert_line "simple with flag"
}

@test "docops with markdown"  {
    run ops docopts
    assert_success
    assert_line  --partial "--------"
    assert_line  --partial "Synopsis"
    assert_line  --partial "simple [-f]"

    run env OPS_NO_DOCOPTS=1 ops docopts simple _f=false
    assert_line --partial "Warning: ignoring docopts.md"
    assert_line "simple without flag"    
    run env OPS_NO_DOCOPTS=1 ops docopts simple _f=true
    assert_line --partial "Warning: ignoring docopts.md"
    assert_line "simple with flag"    

    run ops docopts simple
    assert_line "simple without flag"

    run ops docopts simple -f
    assert_line "simple with flag"
  
    run ops docopts simple -g
    assert_failure
    assert_line "Usage:"
    assert_line --partial "docopts simple [-f]"

}

@test "docops with both"  {
    # show the docopts.txt with a warning
    run ops docopts both
    assert_success
    refute_line "This should be ignored"
    assert_line --partial "both simple [-f]"
    assert_line "Warning: both docopts.txt and docopts.md are present, docopts.txt ignored."

    run ops docopts both simple -f
    assert_success
    assert_line "simple with flag"
}

