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

package auth

import (
	"errors"
	"testing"
)

type stubPasswordReader struct {
	Password    string
	ReturnError bool
}

func (pr stubPasswordReader) ReadPassword() (string, error) {
	if pr.ReturnError {
		return "", errors.New("stubbed error")
	}
	return pr.Password, nil
}

func TestAskPassword(t *testing.T) {
	t.Run("error: returns error when password reader returns error", func(t *testing.T) {
		oldPwdReader := pwdReader
		pwdReader = stubPasswordReader{ReturnError: true}

		result, err := AskPassword()

		pwdReader = oldPwdReader
		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "stubbed error" {
			t.Errorf("Expected error to be 'stubbed error', got %s", err.Error())
		}

		if result != "" {
			t.Errorf("Expected empty string, got %s", result)
		}
	})

	t.Run("success: returns password correctly", func(t *testing.T) {
		oldPwdReader := pwdReader
		pwdReader = stubPasswordReader{Password: "a password", ReturnError: false}

		result, err := AskPassword()

		pwdReader = oldPwdReader

		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}
		if result != "a password" {
			t.Errorf("Expected 'a password', got %s", result)
		}
	})
}
