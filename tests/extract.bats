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
    rm -f  7zip.tar.xz 7zz coreutils.tar.gz coreutils.zip coreutils.exe
}

@test "extract" {
    run ops -extract
    assert_line "Usage: file.(zip|tgz|tar[.gz|.bz2|.xz]) target"
    curl -sL -o7zip.tar.xz https://www.7-zip.org/a/7z2407-linux-x64.tar.xz 

    run ops -extract 7zip.tar.xz missing
    assert_failure
    assert_line "file not found"

    run ops -extract 7zip.tar.xz 7zz
    assert_success
    assert test -x 7zz

    curl -sL -ocoreutils.tar.gz https://github.com/uutils/coreutils/releases/download/0.0.27/coreutils-0.0.27-x86_64-unknown-linux-gnu.tar.gz
    run ops -extract coreutils.tar.gz missing
    assert_failure
    assert_line "file not found"

    
    run ops -extract coreutils.tar.gz coreutils
    assert_success
    assert test -x coreutils

    
    curl -sL -ocoreutils.zip https://github.com/uutils/coreutils/releases/download/0.0.27/coreutils-0.0.27-x86_64-pc-windows-msvc.zip
    run ops -extract coreutils.zip missing
    assert_failure
    assert_line "file not found"

    run ops -extract coreutils.zip coreutils.exe
    assert_success
    assert test -e coreutils.exe
}    

@test "-empty" {
    run ops -empty
    assert_line "Usage: filename"
    run ops -empty empty_file
    assert test -f empty_file
    assert_success
    run ops -empty empty_file
    assert_failure
    assert_line "file already exists"
    rm empty_file

}