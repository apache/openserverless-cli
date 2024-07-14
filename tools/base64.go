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
	"encoding/base64"
	"flag"
	"fmt"
)

func base64Tool() error {

	var (
		helpFlag bool
		encode   string
		decode   string
	)

	flag.Usage = func() {
		fmt.Println(`Usage: nuv -base64 [options] <string>

Options:
	-h, --help             Display this help message
	-e, --encode <string>  Encode a string to base64
	-d, --decode <string>  Decode a base64 string`)
	}

	flag.BoolVar(&helpFlag, "h", false, "Display this help message")
	flag.BoolVar(&helpFlag, "help", false, "Display this help message")
	flag.StringVar(&encode, "e", "", "Encode a string to base64")
	flag.StringVar(&encode, "encode", "", "Encode a string to base64")
	flag.StringVar(&decode, "d", "", "Decode a base64 string")
	flag.StringVar(&decode, "decode", "", "Decode a base64 string")

	flag.Parse()

	if helpFlag {
		flag.Usage()
		return nil
	}

	if encode != "" {
		encodedString := base64.StdEncoding.EncodeToString([]byte(encode))
		fmt.Println(encodedString)
		return nil
	}

	if decode != "" {
		decodedBytes, err := base64.StdEncoding.DecodeString(decode)
		if err != nil {
			return err
		}
		decodedString := string(decodedBytes)
		fmt.Println(decodedString)
		return nil
	}

	flag.Usage()
	return nil
}
