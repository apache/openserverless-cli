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

import "testing"

func Test_validateEmail(t *testing.T) {

	tests := []struct {
		name string
		args string
		want bool
	}{
		{"valid", "example@email.com", true},
		{"invalid", "example", false},
		{"invalid starting with at", "@email.com.", false},
		{"invalid ending with at", "example@", false},
		{"invalid ending with dot", "example.", false},
		{"invalid ending with dot and at", "example@.", false},
		{"invalid ending with dot and at", "example@.com", false},
		{"invalid number", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidEmail(tt.args); got != tt.want {
				t.Errorf("validateEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidNumber(t *testing.T) {

	tests := []struct {
		name string
		args string
		want bool
	}{
		{"valid", "123", true},
		{"valid with decimal", "123.45", true},
		{"invalid", "example", false},
		{"invalid with letters", "123abc", false},
		{"invalid with plus", "123+45", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidNumber(tt.args); got != tt.want {
				t.Errorf("isValidNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidByRegex(t *testing.T) {

	tests := []struct {
		name  string
		regex string
		value string
		want  bool
	}{
		{"valid", "^[a-z]+$", "example", true},
		{"valid with numbers only", "^[0-9]+$", "123", true},
		{"invalid", "^[a-z]+$", "123", false},
		{"invalid", "^[a-z]+$", "example123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isValidByRegex(tt.value, tt.regex)
			if err != nil {
				t.Fatalf("isValidByRegex() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("isValidByRegex() = %v, want %v", got, tt.want)
			}
		})
	}

}
