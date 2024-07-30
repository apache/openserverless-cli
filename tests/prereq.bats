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
    export NUV_BRANCH="0.1.0-testing"
    rm -Rvf ~/.nuv/
    cd prereq
}

@test "ops prereq" {
    run ops
    assert_line --partial "ensuring prerequisite 7zz" 
    assert_line --partial "ensuring prerequisite coreutils" 
    assert_line --partial "ensuring prerequisite bun" 
    assert_line --partial "ensuring prerequisite kubectl" 
    assert_line --partial "ensuring prerequisite kind" 
    assert_line --partial "ensuring prerequisite k3sup" 
    assert_line --partial "info" 
    run ops
    refute_line  "ensuring prerequisite 7zz" 
    refute_line  "ensuring prerequisite coreutils" 
    refute_line  "ensuring prerequisite bun" 
    refute_line  "ensuring prerequisite kubectl" 
    refute_line  "ensuring prerequisite kind" 
    refute_line  "ensuring prerequisite k3sup" 
    assert_line --partial "info" 
}


@test "download others" {
    skip "todo"
    for o in linux darwin
    do for a in arm64 amd64
       do  
          #o=linux a=amd64
          run env OS=$o ARCH=$a ops 
          assert_line --partial "ensuring prerequisite 7zz" 
          assert test -e ~/.nuv/$o-$a/bin/7zz
       done
    done 
    run env OS=windows ops
    assert test -e ~/.nuv/windows-*/bin/7zz.exe
}

