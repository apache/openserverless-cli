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

# setup() {
#     load 'test_helper/bats-support/load'
#     load 'test_helper/bats-assert/load'
#     export NO_COLOR=1
# }

# @test "-scan help msg" {
#     run nuv -scan --help
#     assert_line --partial "Usage:"

#     run nuv -scan
#     assert_line --partial "Usage:"
#     assert_failure
# }

# @test "-scan stops if actions folder not present" {
#     NUV_PWD="./olaris" run nuv -scan nuv -js
#     assert_line --partial "actions folder not found"
#     assert_failure
# }

# @test "-scan -js argv.js without -g" {
#     NUV_PWD="./testdata" run nuv -scan nuv -js testdata/js_test_argv.js
#     WD=$(pwd)
#     assert_line --partial "goja,testdata/js_test_argv.js,$WD/testdata/actions"
#     assert_line --partial "goja,testdata/js_test_argv.js,$WD/testdata/actions/subfolder"
#     assert_line --partial "goja,testdata/js_test_argv.js,$WD/testdata/actions/subfolder/subsub"
#     assert_success
# }

# @test "-scan -js argv.js with -g *" {
#     NUV_PWD="./testdata" run nuv -scan -glob "*" nuv -js testdata/js_test_argv.js
#     WD=$(pwd)
#     assert_line --partial "goja,testdata/js_test_argv.js,$WD/testdata/actions"
#     assert_line --partial "goja,testdata/js_test_argv.js,$WD/testdata/actions/subfolder"
#     assert_line --partial "goja,testdata/js_test_argv.js,$WD/testdata/actions/subfolder/subsub,hello.js,hello.py"
#     assert_success
# }

# @test "-scan --dry-run" {
#     NUV_PWD="./testdata" run nuv -scan -glob "*" --dry-run nuv -scan -js testdata/js_test_argv.js
#     WD=$(pwd)
#     assert_line --partial "nuv -scan -js testdata/js_test_argv.js $WD/testdata/actions"
#     assert_line --partial "nuv -scan -js testdata/js_test_argv.js $WD/testdata/actions/subfolder"
#     assert_line --partial "nuv -scan -js testdata/js_test_argv.js $WD/testdata/actions/subfolder/subsub hello.js hello.py"
#     assert_success
# }

# @test "-scan glob pattern" {
#     NUV_PWD="./testdata" run nuv -scan --dry-run -glob "*.js" nuv -js testdata/js_test_argv.js
#     WD=$(pwd)
#     assert_line --partial "nuv -js testdata/js_test_argv.js $WD/testdata/actions"
#     assert_line --partial "nuv -js testdata/js_test_argv.js $WD/testdata/actions/subfolder"
#     assert_line --partial "nuv -js testdata/js_test_argv.js $WD/testdata/actions/subfolder/subsub hello.js"
#     assert_success
# }
