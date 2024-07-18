/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package runtimes

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/apache/openwhisk-client-go/whisk"
	"github.com/nuvolaris/openwhisk-wskdeploy/utils"
	"github.com/nuvolaris/openwhisk-wskdeploy/wskderrors"
	"github.com/nuvolaris/openwhisk-wskdeploy/wski18n"
	"github.com/nuvolaris/openwhisk-wskdeploy/wskprint"
)

const (
	NODEJS_FILE_EXTENSION   = "js"
	SWIFT_FILE_EXTENSION    = "swift"
	PYTHON_FILE_EXTENSION   = "py"
	JAVA_FILE_EXTENSION     = "java"
	JAR_FILE_EXTENSION      = "jar"
	CSHARP_FILE_EXTENSION   = "cs"
	PHP_FILE_EXTENSION      = "php"
	ZIP_FILE_EXTENSION      = "zip"
	RUBY_FILE_EXTENSION     = "rb"
	GO_FILE_EXTENSION       = "go"
	RUST_FILE_EXTENSION     = "rs"
	NODEJS_RUNTIME          = "nodejs"
	SWIFT_RUNTIME           = SWIFT_FILE_EXTENSION
	PYTHON_RUNTIME          = "python"
	JAVA_RUNTIME            = JAVA_FILE_EXTENSION
	DOTNET_RUNTIME          = ZIP_FILE_EXTENSION
	PHP_RUNTIME             = PHP_FILE_EXTENSION
	RUBY_RUNTIME            = "ruby"
	GO_RUNTIME              = GO_FILE_EXTENSION
	RUST_RUNTIME            = "rust"
	HTTP_CONTENT_TYPE_KEY   = "Content-Type"
	HTTP_CONTENT_TYPE_VALUE = "application/json; charset=UTF-8"
	RUNTIME_NOT_SPECIFIED   = "NOT SPECIFIED"
	BLACKBOX                = "blackbox"
	HTTPS                   = "https://"
)

// Structs used to denote the OpenWhisk Runtime information
type Limit struct {
	Apm       uint `json:"actions_per_minute"`
	Tpm       uint `json:"triggers_per_minute"`
	ConAction uint `json:"concurrent_actions"`
}

type Runtime struct {
	Deprecated bool   `json:"deprecated"`
	Default    bool   `json:"default"`
	Kind       string `json:"kind"`
}

type SupportInfo struct {
	Github string `json:"github"`
	Slack  string `json:"slack"`
}

type OpenWhiskInfo struct {
	Support  SupportInfo          `json:"support"`
	Desc     string               `json:"description"`
	ApiPath  []string             `json:"api_paths"`
	Runtimes map[string][]Runtime `json:"runtimes"`
	Limits   Limit                `json:"limits"`
}

var FileExtensionRuntimeKindMap map[string]string
var SupportedRunTimes map[string][]string
var DefaultRunTimes map[string]string
var FileRuntimeExtensionsMap map[string]string

func GetRuntimesByUrl(opURL string, pop *OpenWhiskInfo) error {

	// configure transport
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	var netTransport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	var netClient = &http.Client{
		Timeout:   time.Second * utils.DEFAULT_HTTP_TIMEOUT,
		Transport: netTransport,
	}

	req, _ := http.NewRequest("GET", opURL, nil)
	req.Header.Set(HTTP_CONTENT_TYPE_KEY, HTTP_CONTENT_TYPE_VALUE)
	whisk.Debug(whisk.DbgInfo, "trying "+req.URL.String())
	res, err := netClient.Do(req)
	if err != nil {
		// TODO() create an error
		errString := wski18n.T(wski18n.ID_ERR_RUNTIMES_GET_X_err_X,
			map[string]interface{}{"err": err.Error()})
		whisk.Debug(whisk.DbgWarn, errString)
		if utils.Flags.Strict {
			errMessage := wski18n.T(wski18n.ID_ERR_RUNTIME_PARSER_ERROR,
				map[string]interface{}{wski18n.KEY_ERR: err.Error()})
			err = wskderrors.NewRuntimeParserError(errMessage)
		}
		return err
	} else {
		if res != nil {
			defer res.Body.Close()
		}
		b, _ := ioutil.ReadAll(res.Body)
		if b != nil && len(b) > 0 {
			stdout := wski18n.T(wski18n.ID_MSG_UNMARSHAL_NETWORK_X_url_X,
				map[string]interface{}{"url": opURL})
			wskprint.PrintOpenWhiskVerbose(utils.Flags.Verbose, stdout)
			return json.Unmarshal(b, pop)
		}
		return fmt.Errorf("cannot get runtimes")
	}
}

// We could get the openwhisk info from bluemix through running the command
// `curl -k https://openwhisk.ng.bluemix.net`
// hard coding it here in case of network unavailable or failure.
func ParseOpenWhisk(apiHost string) (op OpenWhiskInfo, err error) {
	opURL := apiHost
	_, err = url.ParseRequestURI(opURL)
	if err != nil {
		opURL = HTTPS + opURL
	}

	// trying to download info
	err = GetRuntimesByUrl(opURL+"/api/info", &op)
	if err != nil {
		err = GetRuntimesByUrl(opURL, &op)
	}
	if err != nil {
		stdout := wski18n.T(wski18n.ID_MSG_UNMARSHAL_LOCAL)
		wskprint.PrintOpenWhiskVerbose(utils.Flags.Verbose, stdout)
		runtimes := []byte(os.Getenv("WSK_RUNTIMES_JSON"))
		err = json.Unmarshal(runtimes, &op)
		if err != nil {
			fmt.Printf("cannot parse this json: ===\n%s\n===\n", runtimes)
			return
		}
	}
	return
}

func ConvertToMap(op OpenWhiskInfo) (rt map[string][]string) {
	rt = make(map[string][]string)
	for k, v := range op.Runtimes {
		rt[k] = make([]string, 0, len(v))
		for i := range v {
			if !v[i].Deprecated {
				rt[k] = append(rt[k], v[i].Kind)
			}
			if v[i].Default {
				rt[k] = append(rt[k], strings.Split(v[i].Kind, ":")[0]+":default")
			}
		}
	}
	return
}

func DefaultRuntimes(op OpenWhiskInfo) (rt map[string]string) {
	rt = make(map[string]string)
	for k, v := range op.Runtimes {
		for i := range v {
			if !v[i].Deprecated {
				if v[i].Default {
					rt[k] = v[i].Kind
				}
			}
		}
	}
	return
}

func FileExtensionRuntimes(op OpenWhiskInfo) (ext map[string]string) {
	ext = make(map[string]string)
	for k := range op.Runtimes {
		if strings.Contains(k, NODEJS_RUNTIME) {
			ext[NODEJS_FILE_EXTENSION] = k
		} else if strings.Contains(k, PYTHON_RUNTIME) {
			ext[PYTHON_FILE_EXTENSION] = k
		} else if strings.Contains(k, SWIFT_RUNTIME) {
			ext[SWIFT_FILE_EXTENSION] = k
		} else if strings.Contains(k, PHP_RUNTIME) {
			ext[PHP_FILE_EXTENSION] = k
		} else if strings.Contains(k, JAVA_RUNTIME) {
			ext[JAVA_FILE_EXTENSION] = k
			ext[JAR_FILE_EXTENSION] = k
		} else if strings.Contains(k, RUBY_RUNTIME) {
			ext[RUBY_FILE_EXTENSION] = k
		} else if strings.Contains(k, RUST_RUNTIME) {
			ext[RUST_FILE_EXTENSION] = k
		} else if strings.Contains(k, GO_RUNTIME) {
			ext[GO_FILE_EXTENSION] = k
		} else if strings.Contains(k, DOTNET_RUNTIME) {
			ext[CSHARP_FILE_EXTENSION] = k
			ext[ZIP_FILE_EXTENSION] = k
		} else if strings.Contains(k, RUST_RUNTIME) {
			ext[RUST_FILE_EXTENSION] = k
		}
	}
	return
}

func FileRuntimeExtensions(op OpenWhiskInfo) (rte map[string]string) {
	rte = make(map[string]string)

	for k, v := range op.Runtimes {
		for i := range v {
			if !v[i].Deprecated {
				if strings.Contains(k, NODEJS_RUNTIME) {
					rte[v[i].Kind] = NODEJS_FILE_EXTENSION
				} else if strings.Contains(k, PYTHON_RUNTIME) {
					rte[v[i].Kind] = PYTHON_FILE_EXTENSION
				} else if strings.Contains(k, SWIFT_RUNTIME) {
					rte[v[i].Kind] = SWIFT_FILE_EXTENSION
				} else if strings.Contains(k, PHP_RUNTIME) {
					rte[v[i].Kind] = PHP_FILE_EXTENSION
				} else if strings.Contains(k, JAVA_RUNTIME) {
					rte[v[i].Kind] = JAVA_FILE_EXTENSION
				} else if strings.Contains(k, RUBY_RUNTIME) {
					rte[v[i].Kind] = RUBY_FILE_EXTENSION
				} else if strings.Contains(k, RUST_RUNTIME) {
					rte[v[i].Kind] = RUST_FILE_EXTENSION
				} else if strings.Contains(k, GO_RUNTIME) {
					rte[v[i].Kind] = GO_FILE_EXTENSION
				} else if strings.Contains(k, DOTNET_RUNTIME) {
					rte[v[i].Kind] = CSHARP_FILE_EXTENSION
				} else if strings.Contains(k, RUST_RUNTIME) {
					rte[v[i].Kind] = RUST_FILE_EXTENSION
				}
			}
		}
	}
	return
}

func CheckRuntimeConsistencyWithFileExtension(ext string, runtime string) bool {
	rt := FileExtensionRuntimeKindMap[ext]
	for _, v := range SupportedRunTimes[rt] {
		if runtime == v {
			return true
		}
	}
	return false
}

func CheckExistRuntime(rtname string, runtimes map[string][]string) bool {
	for _, v := range runtimes {
		for i := range v {
			if rtname == v[i] {
				return true
			}
		}
	}
	return false
}

func ListOfSupportedRuntimes(runtimes map[string][]string) (rt []string) {
	for _, v := range runtimes {
		for _, t := range v {
			rt = append(rt, t)
		}
	}
	return
}
