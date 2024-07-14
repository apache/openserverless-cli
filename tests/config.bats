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

    # create ~/.nuv if it doesn't exist
    mkdir -p ~/.nuv
}

@test "config usage print" {
    run nuv -config
    assert_line "Usage:"
    assert_success

    run nuv -config -h
    assert_line "Usage:"
    assert_success

    run nuv -config --help
    assert_line "Usage:"
    assert_success
}

@test "set simple var in config.json" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY=VALUE
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "key": "VALUE"'
}

@test "set complex var in config.json" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY='{"a": 1}'
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "key": {'
    assert_line '    "a": 1'
    assert_line '  }'
}

@test "set multiple keys in config.json" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY_NESTED=123 KEY_SIMPLE=abc
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "key": {'
    assert_line '    "nested": 123,'
    assert_line '    "simple": "abc"'
    assert_line '  }'
}

@test "replace existing key in config.json" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY=VALUE
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "key": "VALUE"'

    run nuv -config KEY=NEW_VALUE
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "key": "NEW_VALUE"'
}

@test "replace existing key with different type" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY=VALUE
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "key": "VALUE"'

    run nuv -config KEY='{"a": 1}'
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "key": {'
    assert_line '    "a": 1'
    assert_line '  }'

}

@test "add keys to existing config.json" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY=VALUE
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "key": "VALUE"'

    run nuv -config ANOTHER=123
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "another": 123,'
    assert_line '  "key": "VALUE"'
}

@test "merge object keys" {
    run rm -f ~/.nuv/config.json
    run nuv -config NESTED_KEY=123
    assert_success

    run nuv -config NESTED_ANOTHER=456
    assert_success

    run cat ~/.nuv/config.json
    assert_line '  "nested": {'
    assert_line '    "another": 456,'
    assert_line '    "key": 123'
    assert_line '  }'   
}

@test "dump configs" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY=VALUE ANOTHER=123
    assert_success

    run nuv -config --dump
    assert_success
    assert_line 'KEY=VALUE'
    assert_line 'ANOTHER=123'
}

@test "remove config values" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY=VALUE ANOTHER=123
    assert_success

    run nuv -config --remove KEY
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "another": 123'

    run nuv -config --remove ANOTHER
    assert_success
    run cat ~/.nuv/config.json
    assert_line '{}'
}

@test "remove nested values" {
    run rm -f ~/.nuv/config.json
    run nuv -config NESTED_KEY=VALUE NESTED_ANOTHER=123
    assert_success

    run nuv -config --remove NESTED_KEY
    assert_success
    run cat ~/.nuv/config.json
    assert_line '  "nested": {'
    assert_line '    "another": 123'
    assert_line '  }'

    run nuv -config --remove NESTED_ANOTHER
    assert_success
    run cat ~/.nuv/config.json
    assert_line '{}'
}

@test "read single value" {
    run rm -f ~/.nuv/config.json
    run nuv -config KEY=VALUE
    assert_success

    run nuv -config KEY
    assert_success
    assert_line 'VALUE'

    # read nested value
    run nuv -config NESTED_KEY=new_value
    assert_success

    run nuv -config NESTED_KEY
    assert_success
    assert_line 'new_value'

    # read multiple values
    run nuv -config KEY NESTED_KEY
    assert_success
    assert_line 'VALUE'
    assert_line 'new_value'   
}