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
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const defaultCharRange = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type randomGenerator interface {
	GenerateFloat01()
	GenerateString(length int, chars string)
	GenerateInteger(min, max int)
	GenerateUUID() error
}

type randomGeneratorImpl struct {
	rand *rand.Rand
}

var randomGen randomGenerator = randomGeneratorImpl{
	rand: rand.New(rand.NewSource(time.Now().UnixNano())),
}

func (r randomGeneratorImpl) GenerateFloat01() {
	fmt.Println(r.rand.Float64())
}

func (r randomGeneratorImpl) GenerateString(length int, chars string) {
	var buf bytes.Buffer

	for i := 0; i < length; i++ {
		randIndex := r.rand.Intn(len(chars))
		randChar := chars[randIndex]
		buf.WriteByte(randChar)
	}

	fmt.Println(buf.String())
}

func (r randomGeneratorImpl) GenerateInteger(min, max int) {
	fmt.Println(r.rand.Intn(max-min) + min)
}

func (r randomGeneratorImpl) GenerateUUID() error {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	fmt.Println(uuid.String())
	return nil
}

func RandTool(args ...string) error {

	flag := flag.NewFlagSet("random", flag.ExitOnError)

	flag.Usage = func() {
		fmt.Println(`Generate random numbers, strings and uuids
		
Usage:
	nuv -random [options]

Options:
	-h, --help  shows this help
	-u, --uuid  generates a random uuid v4
	--int  <max> [min] generates a random non-negative integer between min and max (default min=0)
	--str  <len> [<characters>] generates an alphanumeric string of length <len> from the set of <characters> provided (default <characters>=a-zA-Z0-9)`)
	}

	var helpFlag bool
	var intFlag int
	var strFlag int
	var uuidFlag bool

	// Define command line flags
	flag.BoolVar(&helpFlag, "h", false, "Show help message")
	flag.BoolVar(&helpFlag, "help", false, "Show help message")
	flag.IntVar(&intFlag, "int", -1, "Generate a random integer")
	flag.IntVar(&strFlag, "str", -1, "Generate a random string")
	flag.BoolVar(&uuidFlag, "u", false, "Generate a random uuid")
	flag.BoolVar(&uuidFlag, "uuid", false, "Generate a random uuid")

	// Parse command line flags
	err := flag.Parse(args)
	if err != nil {
		return err
	}

	// Print help message if -h flag is provided
	if helpFlag {
		flag.Usage()
		return nil
	}

	if uuidFlag {
		return randomGen.GenerateUUID()
	}

	if isInputFlag(*flag, "int") {
		if flag.NArg() > 1 {
			flag.Usage()
			return errors.New("invalid number of arguments, expected 1 or 2 for --int")
		}

		max := intFlag
		min := 0

		if max <= 0 {
			return fmt.Errorf("invalid max value: %v. Must be greater than 0", max)
		}

		if flag.NArg() == 1 {
			minOpt, err := strconv.Atoi(flag.Arg(0))
			if err != nil {
				return err
			}

			min = minOpt
		}

		if min >= max {
			return fmt.Errorf("invalid min value: %v. Must be less than max value: %v", min, max)
		}

		randomGen.GenerateInteger(min, max)
		return nil
	}

	if isInputFlag(*flag, "str") {
		if flag.NArg() > 1 {
			flag.Usage()
			return errors.New("invalid number of arguments, expected 1 or 2 for --str")
		}

		length := strFlag
		chars := defaultCharRange

		if length <= 0 {
			return fmt.Errorf("invalid length value: %v. Must be greater than 0", length)
		}

		if flag.NArg() == 1 {
			chars = flag.Arg(0)
		}

		randomGen.GenerateString(length, chars)
		return nil
	}

	// Get remaining args
	if flag.NArg() > 0 {
		flag.Usage()
		return errors.New("invalid number of arguments")
	}

	randomGen.GenerateFloat01()
	return nil
}

func isInputFlag(fs flag.FlagSet, flagName string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == flagName {
			found = true
		}
	})
	return found
}
