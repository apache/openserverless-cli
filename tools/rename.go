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

func Rename() (int, error) {
	if len(os.Args) != 3 {
		fmt.Println(MarkdownHelp("rename"))
		return 0, nil
	}

	source := os.Args[1]
	destination := os.Args[2]

	err := os.Rename(source, destination)
	if err != nil {
		return 1, err
	}

	fmt.Printf("renamed %s -> %s\n", source, destination)
	return 0, nil
}
