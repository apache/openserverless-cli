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
	"bytes"
	"fmt"
	"io"
	"testing"
)

type fakeRandomGenerator struct {
	w          io.Writer
	fakeFloat  float64
	fakeString string
	fakeInt    int
	fakeUUID   string
}

func (r fakeRandomGenerator) GenerateFloat01() {
	fmt.Fprint(r.w, r.fakeFloat)
}

func (r fakeRandomGenerator) GenerateString(len int, chars string) {
	fmt.Fprint(r.w, r.fakeString)
}

func (r fakeRandomGenerator) GenerateInteger(min, max int) {
	fmt.Fprint(r.w, r.fakeInt)
}

func (r fakeRandomGenerator) GenerateUUID() error {
	fmt.Fprint(r.w, r.fakeUUID)
	return nil
}

func TestRandom(t *testing.T) {
	var output bytes.Buffer
	randomGen = fakeRandomGenerator{
		w:          &output,
		fakeFloat:  0.5,
		fakeString: "fakeString",
		fakeInt:    15,
		fakeUUID:   "fakeUUID",
	}

	tests := []struct {
		name string
		args []string
		want string
		err  error
	}{
		{"float", []string{}, "0.5", nil},
		{"uuid", []string{"--uuid"}, "fakeUUID", nil},
		{"uuid", []string{"-u"}, "fakeUUID", nil},
		{"int with max", []string{"--int", "10"}, "15", nil},
		{"int with min and max", []string{"--int", "20", "10"}, "15", nil},
		{"string with length", []string{"--str", "10"}, "fakeString", nil},
		{"string with length and chars", []string{"--str", "10", "abc"}, "fakeString", nil},
		//
		{"int with invalid max", []string{"--int", "0"}, "", fmt.Errorf("invalid max value: 0. Must be greater than 0")},
		{"int with negative max", []string{"--int", "-1"}, "", fmt.Errorf("invalid max value: -1. Must be greater than 0")},
		{"int with min > max", []string{"--int", "10", "20"}, "", fmt.Errorf("invalid min value: 20. Must be less than max value: 10")},
		{"string with invalid length", []string{"--str", "0"}, "", fmt.Errorf("invalid length value: 0. Must be greater than 0")},
		{"string with negative length", []string{"--str", "-1"}, "", fmt.Errorf("invalid length value: -1. Must be greater than 0")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RandTool(tt.args...)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("RandTool() error = %v, wantErr %v", err, tt.err)
			}

			got := output.String()
			if got != tt.want {
				t.Errorf("RandTool() = %v, want %v", got, tt.want)
			}
			output.Reset()
		})
	}
}
