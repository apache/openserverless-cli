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
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func validateTool() error {
	flag := flag.NewFlagSet("validate", flag.ContinueOnError)

	flag.Usage = func() {
		fmt.Println(`Usage:
nuv -validate [-e] [-m | -n | -r <regex>] <value> [<message>]

Check if a value is valid according to the given constraints.
If -e is specified, the value is retrieved from the environment variable with the given name.

Options:`)
		flag.PrintDefaults()
	}

	helpFlag := flag.Bool("h", false, "Print this help message.")
	envFlag := flag.Bool("e", false, "Retrieve value from the environment variable with the given name.")
	mailFlag := flag.Bool("m", false, "Check if the value is a valid email address.")
	numberFlag := flag.Bool("n", false, "Check if the value is a number.")
	regexFlag := flag.String("r", "", "Check if the value matches the given regular expression.")

	err := flag.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	if *helpFlag {
		flag.Usage()
		return nil
	}

	if flag.NArg() != 1 && flag.NArg() != 2 {
		flag.Usage()
		return errors.New("invalid number of arguments")
	}

	arg := flag.Arg(0)
	customErr := flag.Arg(1)
	if customErr == "" {
		customErr = "validation failed"
	}
	value := arg

	if *envFlag {
		value = os.Getenv(arg)
		if value == "" {
			return fmt.Errorf("variable '%s' not set", arg)
		}
	}

	if *mailFlag {
		if !isValidEmail(value) {
			return fmt.Errorf(customErr)
		}
		return nil
	}

	if *numberFlag {
		if !isValidNumber(value) {
			return fmt.Errorf(customErr)
		}
		return nil
	}

	if *regexFlag != "" {
		valid, err := isValidByRegex(value, *regexFlag)
		if err != nil {
			return err
		}

		if !valid {
			return fmt.Errorf(customErr)
		}
	}

	return nil
}

func isValidNumber(number string) bool {
	_, err := strconv.ParseFloat(number, 64)
	return err == nil
}

func isValidEmail(email string) bool {
	// Regular expression pattern for email validation
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Create a regular expression object
	regExp := regexp.MustCompile(pattern)

	// Use the regular expression to match the email string
	return regExp.MatchString(email)
}

func isValidByRegex(value string, regex string) (bool, error) {
	// Create a regular expression object
	regExp, err := regexp.Compile(regex)
	if err != nil {
		return false, err
	}

	// Use the regular expression to match the email string
	return regExp.MatchString(value), nil
}
