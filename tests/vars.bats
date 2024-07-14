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


@test "simple" {
    run nuv sub vars simple
    # use top level
    assert_line 'eV=e2 pV=p2'

    run env V=e1 nuv sub vars simple V=p1
    # override external
    assert_line 'eV=e1 pV=p1'
}


@test "inner" {
    run nuv sub vars inner
    # use inner
    assert_line 'eV=e3 pV=p3'

    run nuv sub vars inner V=p1
    # no override var for inner var
    assert_line 'eV=e3 pV=p3'

    run env V=e1 nuv sub vars inner
    # env override inner env
    assert_line 'eV=e1 pV=p3'

    run env V=e1 nuv sub vars inner V=p1
    # no override var for inner var
    # env override inner env
    assert_line 'eV=e1 pV=p3'
}


@test "override" {
    run nuv sub vars prio
    # vars see each other does not see env
    # env sees parameters
    assert_line 'pOP=p2 pOE= eOE= eOP=p2'

    run env V=e1 nuv sub vars prio
    assert_line 'pOP=p2 pOE=e1 eOE=e1 eOP=p2'

    run nuv sub vars prio V=p1
    assert_line 'pOP=p1 pOE= eOE= eOP=p1'

    run env V=e1 nuv sub vars prio V=p1
    assert_line 'pOP=p1 pOE=e1 eOE=e1 eOP=p1'
}

@test "saved" {
  # local override envfile
  # external override envfile

  run nuv sub vars env
  assert_line  'E=3 EE=2'

  run env E=1 nuv sub vars env
  assert_line 'E=1 EE=2'

  # saved locally
  run nuv sub vars saved
  assert_line 'S=3 SS=2 SSS=4 overriden SS=2'
}

@test "saved-vars" {
    run nuv sub vars clean
    run nuv sub vars v1v2
    assert_line 'V1=x V2='
    run nuv sub vars save1 V1=a
    run nuv sub vars v1v2
    assert_line 'V1=a V2='
    run nuv sub vars save1 V1=b V2=c
    run nuv sub vars v1v2
    assert_line 'V1=b V2=c'
    run nuv sub vars save2 V2=d
    run nuv sub vars v1v2
    assert_line 'V1=b V2=d'
    run nuv sub vars v1v2 V1=a
    assert_line 'V1=a V2=d'
    run nuv sub vars v1v2 V2=c
    assert_line 'V1=b V2=c'
    run nuv sub vars clean
}





nuv sub vars save2 V1=a

V1= V2=
nuv sub vars v1v2



