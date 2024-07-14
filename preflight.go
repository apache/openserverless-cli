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
	"fmt"
	"os/exec"
)

type preflight struct {
	err error
}

type checkFn func(pd *preflight)

func (p *preflight) check(f checkFn) {
	if p.err != nil {
		// skip if previous check failed
		return
	}
	f(p)
}

// preflightChecks performs preflight checks:
// - curl is installed
// - ssh is installed
func preflightChecks() error {
	trace("preflight checks")
	preflight := preflight{}

	preflight.check(checkInstalled("curl"))
	preflight.check(checkInstalled("ssh"))
	// preflight.check(checkInstalled("grep"))

	return preflight.err
}

func checkInstalled(check string) checkFn {
	return func(p *preflight) {
		trace("check installed", check)
		_, err := exec.LookPath(check)
		if err != nil {
			p.err = fmt.Errorf("%s not installed but is required", check)
		}
	}
}
