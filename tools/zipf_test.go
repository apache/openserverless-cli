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
	"testing"
)

func TestGenerateOutputFileName(t *testing.T) {
	t.Run("NoMainFiles", func(t *testing.T) {
		// Test case where there are no main files in the folder
		result, err := generateOutputFileName("testdata/no_main_files/")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := "no_main_files.zip"
		if result != expected {
			t.Errorf("Expected: %s, but got: %s", expected, result)
		}
	})

	t.Run("JavaScriptMainFile", func(t *testing.T) {
		// Test case where there is a JavaScript main file
		result, err := generateOutputFileName("testdata/js_main/")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := "js_main.js.zip"
		if result != expected {
			t.Errorf("Expected: %s, but got: %s", expected, result)
		}
	})

	// t.Run("PythonMainFile", func(t *testing.T) {
	// 	// Test case where there is a Python main file
	// 	result, err := generateOutputFileName("testdata/py_main/")
	// 	if err != nil {
	// 		t.Errorf("Unexpected error: %v", err)
	// 	}
	// 	expected := "py_main.py.zip"
	// 	if result != expected {
	// 		t.Errorf("Expected: %s, but got: %s", expected, result)
	// 	}
	// })

}
