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
	"fmt"
	"os"
)

func Empty() (int, error) {
	if len(os.Args) < 2 {
		fmt.Println("Empty creates an empty file - returns error if it already exists\nUsage: filename")
		return 0, nil
	}
	filename := os.Args[1]

	// Check if the file exists
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return 1, fmt.Errorf("file already exists")
	}

	// Create an empty file
	file, err := os.Create(filename)
	if err != nil {
		return 1, err
	}
	defer file.Close()
	return 0, nil

}
