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
	"flag"
	"fmt"
	"time"

	"golang.org/x/exp/maps"
)

func printDateFmtUsage() {
	fmts := maps.Keys(dateFormats)

	fmt.Printf(`Usage:
nuv -datefmt [options] [arguments]

Print date with different formats. If no time stamp or date strings are given, uses current time.

-h, --help		print this help info
-t, --timestamp		unix timestamp to format (default: current time)
-s, --str 	  	date string to format
--if			input format to use with input date string (via --str)
-f, --of		output format to use (default: UnixDate)

Possible formats: %v
`, fmts)
}

var dateFormats = map[string]string{
	"Layout":      time.Layout,
	"ANSIC":       time.ANSIC,
	"UnixDate":    time.UnixDate,
	"RubyDate":    time.RubyDate,
	"RFC822":      time.RFC822,
	"RFC822Z":     time.RFC822Z,
	"RFC850":      time.RFC850,
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
	"RFC3339":     time.RFC3339,
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	// Handy time stamps.
	"Stamp":      time.Stamp,
	"StampMilli": time.StampMilli,
	"StampMicro": time.StampMicro,
	"StampNano":  time.StampNano,
	// hack! for some reason on GitHub Actions those constants are missing - replaced with their value
	"DateTime":     time.DateTime,
	"DateOnly":     time.DateOnly,
	"TimeOnly":     time.TimeOnly,
	"ms":           "Milliseconds",
	"Milliseconds": "Milliseconds",
}

var (
	helpFlag      bool
	timestampFlag int64
	strFlag       string
	iFmtFlag      string
	oFmtFlag      string
)

func DateFmtTool() error {
	flag.Usage = printDateFmtUsage

	flag.BoolVar(&helpFlag, "h", false, "print this help info")
	flag.BoolVar(&helpFlag, "help", false, "print this help info")
	flag.Int64Var(&timestampFlag, "t", time.Now().Unix(), "unix timestamp to convert")
	flag.Int64Var(&timestampFlag, "timestamp", time.Now().Unix(), "unix timestamp to convert")
	flag.StringVar(&strFlag, "s", "", "date string to convert")
	flag.StringVar(&strFlag, "str", "", "date string to convert")
	flag.StringVar(&iFmtFlag, "if", "", "input format to use")
	flag.StringVar(&oFmtFlag, "of", "UnixDate", "output format to use")
	flag.StringVar(&oFmtFlag, "f", "UnixDate", "output format to use")

	flag.Parse()

	if helpFlag {
		flag.Usage()
		return nil
	}

	ofmt, ok := dateFormats[oFmtFlag]
	if !ok {
		return fmt.Errorf("error: invalid output format: %s", oFmtFlag)
	}

	if strFlag != "" && iFmtFlag != "" {
		ifmt, ok := dateFormats[iFmtFlag]
		if !ok {
			return fmt.Errorf("error: invalid input format: %s", iFmtFlag)
		}

		currentDate, err := time.Parse(ifmt, strFlag)
		if err != nil {
			return err
		}

		fmt.Println(applyFormat(currentDate, ofmt))

		return nil
	}

	if strFlag != "" && iFmtFlag == "" {
		return fmt.Errorf("error: both --str and --if must be provided. Only str given: %s", strFlag)
	}
	if strFlag == "" && iFmtFlag != "" {
		return fmt.Errorf("error: both --str and --if must be provided. Only input format given: %s", iFmtFlag)
	}

	date := time.Unix(timestampFlag, 0)
	fmt.Println(applyFormat(date, ofmt))
	return nil
}

func applyFormat(date time.Time, ofmt string) string {
	switch ofmt {
	case "Milliseconds":
		return fmt.Sprintf("%d", date.UnixMilli())
	default:
		return date.Format(ofmt)
	}
}
