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
	"syscall"

	"golang.org/x/term"
)

type PasswordReader interface {
	ReadPassword() (string, error)
}

type StdInPasswordReader struct{}

func (r StdInPasswordReader) ReadPassword() (string, error) {
	pwd, error := term.ReadPassword(int(syscall.Stdin))
	return string(pwd), error
}

var pwdReader PasswordReader = StdInPasswordReader{}

func AskPassword() (string, error) {
	pwd, err := pwdReader.ReadPassword()
	if err != nil {
		return "", err
	}
	if len(pwd) == 0 {
		return "", errors.New("password is empty")
	}
	return pwd, nil
}
