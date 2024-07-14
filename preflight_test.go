// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testNoopCheck(p *preflight) {}
func testErrCheck(p *preflight) {
	p.err = errors.New("test error in check")
}

func Test_preflight(t *testing.T) {
	p := preflight{}
	p.check(testNoopCheck)
	require.NoError(t, p.err)

	p.err = errors.New("test error")
	p.check(testNoopCheck)
	require.Error(t, p.err)
	require.Equal(t, "test error", p.err.Error())

	p.err = nil
	p.check(testErrCheck)
	require.Error(t, p.err)
	require.Equal(t, "test error in check", p.err.Error())
}
