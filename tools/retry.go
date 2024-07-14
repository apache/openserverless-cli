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
	"os"
	"os/exec"
	"time"

	"github.com/cenkalti/backoff/v4"
)

func ExpBackoffRetry(args []string) error {
	// Define command line flags
	flag.Usage = printRetryHelp

	var helpFlag bool
	var triesFlag int
	var maxFlag int
	var verboseFlag bool

	flag.BoolVar(&helpFlag, "help", false, "Print help message")
	flag.BoolVar(&helpFlag, "h", false, "Print help message")
	flag.IntVar(&triesFlag, "tries", 10, "Set max retries")
	flag.IntVar(&triesFlag, "t", 10, "Set max retries")
	flag.IntVar(&maxFlag, "max", 60, "Maximum time to run")
	flag.IntVar(&maxFlag, "m", 60, "Maximum time to run")
	flag.BoolVar(&verboseFlag, "verbose", false, "Verbose output")
	flag.BoolVar(&verboseFlag, "v", false, "Verbose output")

	// Parse command line flags
	os.Args = args
	flag.Parse()

	// Print help message if -h flag is provided
	if helpFlag {
		flag.Usage()
		return nil
	}

	// Remaining command line arguments
	rest := flag.Args()
	if len(rest) == 0 {
		flag.Usage()
		return nil
	}

	if verboseFlag {
		fmt.Printf("Retry Parameters: max time=%d seconds, retries=%d times\n", maxFlag, triesFlag)
	}
	runCmd := func(args []string) error {
		cmd := exec.Command(rest[0], rest[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	err := retry(runCmd, rest, maxFlag, triesFlag)

	return err
}

func retry(fn func([]string) error, args []string, maxTime int, maxRetries int) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(maxTime) * time.Second

	err := backoff.Retry(func() error {
		err := fn(args)
		if err != nil {
			return err
		}
		return nil
	}, backoff.WithMaxRetries(b, uint64(maxRetries)))

	if err != nil {
		return fmt.Errorf("failure after %d retries or %d seconds.\n%s", maxRetries, maxTime, err.Error())
	}

	return nil
}

func printRetryHelp() {
	fmt.Print(
		`Usage:
nuv -retry [options] task [task options]
-h, --help	Print help message
-t, --tries=#	Set max retries: Default 10
-m, --max=secs	Maximum time to run (set to 0 to disable): Default 60 seconds
-v, --verbose	Verbose output
`)
}
