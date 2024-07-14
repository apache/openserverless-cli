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
//

package tools

import (
	"fmt"
	"os"

	"github.com/apache/openwhisk-client-go/whisk"
	goi18n "github.com/nicksnyder/go-i18n/i18n"
	"github.com/nuvolaris/openwhisk-cli/commands"
	"github.com/nuvolaris/openwhisk-cli/wski18n"
)

// prepare wsk execution
var cliDebug = os.Getenv("WSK_CLI_DEBUG") // Useful for tracing init() code

var T goi18n.TranslateFunc

func init() {
	if len(cliDebug) > 0 {
		whisk.SetDebug(true)
	}

	T = wski18n.T

	// Rest of CLI uses the Properties struct, so set the build time there
	commands.Properties.CLIVersion = "TODO"
}

// Wsk invokes wsk subcommand
func Wsk(command []string, args ...string) error {

	// set wsk variable
	//setWskEnvVariable(false)

	//debug("WSK_CONFIG_FILE=" + os.Getenv("WSK_CONFIG_FILE"))

	// compose the command
	os.Args = command
	os.Args = append(os.Args, args...)
	// cleanup
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			fmt.Println(T("Application exited unexpectedly"))
		}
	}()
	return commands.Execute()
}
