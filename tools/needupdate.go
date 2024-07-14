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

package tools

import (
	"flag"
	"fmt"

	"github.com/Masterminds/semver"
)

func needUpdateTool(args []string) error {

	flag := flag.NewFlagSet("needupdate", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Println(`Check if a semver version A > semver version B.
Exits with 0 if greater, 1 otherwise.

Usage:
  nuv -needupdate <versionA> <versionB>
  
Options:`)
		flag.PrintDefaults()
	}

	helpFlag := flag.Bool("help", false, "Print this help")

	err := flag.Parse(args)
	if err != nil {
		return err
	}

	if *helpFlag {
		flag.Usage()
		return nil
	}

	if flag.NArg() != 2 {
		flag.Usage()
		return fmt.Errorf("needupdate requires two arguments")
	}

	a := flag.Arg(0)
	b := flag.Arg(1)

	versionA, err := semver.NewVersion(a)
	if err != nil {
		return fmt.Errorf("invalid semantic version: %s", b)
	}

	versionB, err := semver.NewVersion(b)
	if err != nil {
		return fmt.Errorf("invalid semantic version: %s", b)
	}

	if versionA.GreaterThan(versionB) {
		return nil
	}

	return fmt.Errorf("%s is not greater than %s", a, b)
}
