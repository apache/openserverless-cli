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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkdownHelp(t *testing.T) {
	for _, s := range ToolList {
		if s.HasDoc {
			opt := MarkdownHelp(s.Name)
			if opt == "" {
				t.Fatalf("Tool %s doesn't have valid help", s.Name)
			}
		}
	}
}

func TestGetMarkDownSuccess(t *testing.T) {
	t.Helper()
	_, err := GetMarkDown("base64")
	require.NoError(t, err)
}

func TestGetMarkDownError(t *testing.T) {
	t.Helper()
	_, err := GetMarkDown("notexistenttool")
	require.Error(t, err)
}

func ExampleMergeToolsList() {
	mergedList := MergeToolsList([]string{})
	fmt.Println(len(mergedList))
	//Output: 25
}
