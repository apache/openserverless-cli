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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/nuvolaris/nuv/config"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

// setupMockServer sets up a new mock HTTP server with the given test data and expected response
func setupMockServer(t *testing.T, expectedLogin, expectedPass, expectedRes string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		login, ok := requestBody["login"]
		require.True(t, ok, "expected login field in request body")
		require.Equal(t, expectedLogin, login, "expected login %s, got %s", expectedLogin, login)

		password, ok := requestBody["password"]
		require.True(t, ok, "expected password field in request body")
		require.Equal(t, expectedPass, password, "expected password %s, got %s", expectedPass, password)

		_, _ = w.Write([]byte(expectedRes))
	}))

	return server
}
func TestLoginCmd(t *testing.T) {
	keyring.MockInit()

	t.Run("error: returns error when empty password", func(t *testing.T) {
		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "",
			ReturnError: false,
		}
		os.Args = []string{"login", "fakeApiHost", "fakeUser"}
		res, err := LoginCmd()
		pwdReader = oldPwdReader
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "password is empty" {
			t.Errorf("Expected error to be 'password is empty', got %s", err.Error())
		}
		if res != nil {
			t.Errorf("Expected response to be nil, got %v", res)
		}
	})

	t.Run("with only apihost, add received credentials", func(t *testing.T) {
		mockServer := setupMockServer(t, "nuvolaris", "a password", "{\"AUTH\": \"test\"}")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}

		os.Args = []string{"login", mockServer.URL}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader

		require.NoError(t, err)
		require.NotNil(t, loginResult)

		// cred, err := keyring.Get(nuvSecretServiceName, "AUTH")
		nuvHome, err := homedir.Expand("~/.nuv")
		require.NoError(t, err)

		configMap, err := config.NewConfigMapBuilder().
			WithConfigJson(filepath.Join(nuvHome, "config.json")).
			Build()
		require.NoError(t, err)

		v, err := configMap.Get("AUTH")
		require.NoError(t, err)
		require.Equal(t, "test", v)
	})

	t.Run("with apihost and user adds received credentials to secret store", func(t *testing.T) {
		mockServer := setupMockServer(t, "a user", "a password", "{ \"AUTH\": \"testauth\", \"fakeCred\": \"test\"}")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}

		os.Args = []string{"login", mockServer.URL, "a user"}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader
		require.NoError(t, err)
		require.NotNil(t, loginResult)

		// cred, err := keyring.Get(nuvSecretServiceName, "fakeCred")
		// require.NoError(t, err)
		// require.Equal(t, "test", cred)

		nuvHome, err := homedir.Expand("~/.nuv")
		require.NoError(t, err)

		configMap, err := config.NewConfigMapBuilder().
			WithConfigJson(filepath.Join(nuvHome, "config.json")).
			Build()
		require.NoError(t, err)

		v, err := configMap.Get("AUTH")
		require.NoError(t, err)
		require.Equal(t, "testauth", v)

		v, err = configMap.Get("FAKECRED")
		require.NoError(t, err)
		require.Equal(t, "test", v)
	})

	t.Run("error when response body is invalid", func(t *testing.T) {
		mockServer := setupMockServer(t, "a user", "a password", "invalid json")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}
		os.Args = []string{"login", mockServer.URL, "a user"}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader
		require.Error(t, err)
		require.Equal(t, "failed to decode response from login request", err.Error())
		require.Nil(t, loginResult)
	})

	t.Run("error when response body is missing AUTH token", func(t *testing.T) {
		mockServer := setupMockServer(t, "a user", "a password", "{\"fakeCred\": \"test\"}")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}
		os.Args = []string{"login", mockServer.URL, "a user"}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader
		require.Error(t, err)
		require.Equal(t, "missing AUTH token from login response", err.Error())
		require.Nil(t, loginResult)
	})
}

func Test_doLogin(t *testing.T) {
	mockServer := setupMockServer(t, "a user", "a password", "{\"fakeCred\": \"test\"}")
	defer mockServer.Close()

	cred, err := doLogin(mockServer.URL, "a user", "a password")
	require.NoError(t, err)
	require.NotNil(t, cred)
	require.Equal(t, "test", cred["fakeCred"])
}

func Test_storeCredentials(t *testing.T) {
	keyring.MockInit()

	fakeCreds := make(map[string]string)
	fakeCreds["AUTH"] = "fakeValue"
	fakeCreds["REDIS_URL"] = "fakeValue"
	fakeCreds["MONGODB"] = "fakeValue"

	err := storeCredentials(fakeCreds)
	require.NoError(t, err)
	require.NotNil(t, fakeCreds)

	for k := range fakeCreds {
		cred, err := keyring.Get(nuvSecretServiceName, k)
		require.NoError(t, err)
		require.Equal(t, fakeCreds[k], cred)
	}
}
