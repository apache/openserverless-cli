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
	"time"

	"github.com/sciabarracom/openwhisk-client-go/whisk"
	"github.com/sciabarracom/openwhisk-wskdeploy/utils"
	"github.com/sciabarracom/openwhisk-wskdeploy/wskderrors"
	"github.com/sciabarracom/openwhisk-wskdeploy/wski18n"
	"github.com/sciabarracom/openwhisk-wskdeploy/wskprint"
)

var Version = "openserverless"

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
		runtimes := []byte(os.Getenv("OPS_RUNTIMES_JSON"))
		err = json.Unmarshal(runtimes, &op)
		if err != nil {
			fmt.Printf("cannot parse this json: ===\n%s\n===\n", runtimes)
			return
		}
	}
	return
}
