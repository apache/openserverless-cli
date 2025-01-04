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
    export OPS_BRANCH="0.1.0"
    rm -Rvf _bin >/dev/null
    export COUNT=1
}

@test "OS=linux ARCH=amd64" {
    OS=linux ARCH=amd64
    mkdir -p _bin/$OS-$ARCH
    run env __OS=$OS __ARCH=$ARCH task -t prereq/olaris/prereq.yml -d _bin/$OS-$ARCH all
    find _bin/$OS-$ARCH -type f | xargs file | grep ELF | grep x86-64 | wc -l | xargs | tee _count
    assert_equal "$(cat _count)" $COUNT
}

@test "OS=linux ARCH=arm64" {
    OS=linux ARCH=arm64
    mkdir -p _bin/$OS-$ARCH
    run env __OS=$OS __ARCH=$ARCH task -t prereq/olaris/prereq.yml -d _bin/$OS-$ARCH  all
    find _bin/$OS-$ARCH -type f | xargs file | grep ELF | grep ARM | wc -l | xargs | tee _count
    assert_equal "$(cat _count)" $COUNT
}

@test "OS=darwin ARCH=amd64"  {
    OS=darwin ARCH=amd64
    mkdir -p _bin/$OS-$ARCH
    run env __OS=$OS __ARCH=$ARCH task -t prereq/olaris/prereq.yml -d _bin/$OS-$ARCH all
    find _bin/$OS-$ARCH -type f | xargs file | head -n 1 | grep Mach-O | grep x86_64 | wc -l | xargs | tee _count
    assert_equal "$(cat _count)" $COUNT
}

@test "OS=darwin ARCH=arm64"  {
    OS=darwin ARCH=arm64
    mkdir -p _bin/$OS-$ARCH
    run env __OS=$OS __ARCH=$ARCH task -t prereq/olaris/prereq.yml -d _bin/$OS-$ARCH all
    find _bin/$OS-$ARCH -type f | xargs file | head -n 1 | grep Mach-O | grep arm64 | wc -l | xargs | tee _count
    assert_equal "$(cat _count)" $COUNT
}

@test "OS=windows ARCH=amd64"  {
    OS=windows ARCH=amd64
    mkdir -p _bin/$OS-$ARCH
    run env __OS=$OS __ARCH=$ARCH task -t prereq/olaris/prereq.yml -d _bin/$OS-$ARCH __OS=$OS __ARCH=$ARCH all
    find _bin/$OS-$ARCH -type f | xargs file | grep PE32  | wc -l | xargs | tee _count
    assert_equal "$(cat _count)" $COUNT
}
