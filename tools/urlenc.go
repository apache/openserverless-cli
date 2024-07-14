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
	"net/url"
	"os"
	"strings"
)

func URLEncTool() error {
	// Define command-line flags
	fs := flag.NewFlagSet("urlenc", flag.ContinueOnError)

	separator := fs.String("s", "&", "Separator for concatenating the parameters")
	encodeEnv := fs.Bool("e", false, "Encode parameter values from environment variables")
	help := fs.Bool("h", false, "Show help")

	// Parse command-line flags
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	if *help {
		fs.Usage()
		return nil
	}

	// Get command-line arguments
	args := fs.Args()

	// Encode and concatenate parameters
	params := make([]string, 0)

	for _, arg := range args {
		if *encodeEnv {
			arg = os.Getenv(arg)
		}
		encodedValue := url.QueryEscape(arg)
		params = append(params, encodedValue)
	}

	result := strings.Join(params, *separator)
	fmt.Println(result)
	return nil
}
