/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package secretstore

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestGenerate(t *testing.T) {
	tokenPath := "testdata/test-resp-init.json"
	gk := NewGokeyGenerator(tokenPath)
	cr := NewCred(&http.Client{}, tokenPath, gk, "", logger.MockLogger{})

	realm1 := "service1"
	realm2 := "service2"

	p1, err := cr.GeneratePassword(realm1)
	if err != nil {
		t.Errorf("failed to create credential")
		t.Errorf(err.Error())
	}
	p2, err := cr.GeneratePassword(realm2)
	if err != nil {
		t.Errorf("failed to create credential")
		t.Errorf(err.Error())
	}
	if p1 == p2 {
		t.Errorf("error: different master key and realm combination need to generate different passwords")
	}
}

func TestRetrieveCred(t *testing.T) {
	credPath := "testCredPath"
	token := "token"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"username": "test-user", "password": "test-password"}}`))
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != fmt.Sprintf("/%s", credPath) {
			t.Errorf("expected request to /%s, got %s instead", credPath, r.URL.EscapedPath())
		}

		if r.Header.Get(VaultToken) != token {
			t.Errorf("expected request header for %s is %s, got %s instead", VaultToken, token, r.Header.Get(VaultToken))
		}
	}))
	defer ts.Close()

	parsed, err := url.Parse(ts.URL)
	if err != nil {
		t.Errorf("unable to parse test server URL %s", ts.URL)
		return
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Errorf("parsed port number cannot be converted to int %s", parsed.Port())
		return
	}

	configuration := &config.ConfigurationStruct{
		SecretService: secretstoreclient.SecretServiceInfo{
			Server: parsed.Hostname(),
			Port:   port,
			Scheme: "https",
		},
	}

	mockLogger := logger.MockLogger{}
	cr := NewCred(
		NewRequester(true, configuration.SecretService.CaFilePath, mockLogger),
		"",
		NewGokeyGenerator(""),
		configuration.SecretService.GetSecretSvcBaseURL(),
		mockLogger)
	pair, err := cr.retrieve(token, credPath)
	if err != nil {
		t.Errorf("failed to retrieve credential pair")
		t.Errorf(err.Error())
	}
	if pair.User != "test-user" || pair.Password != "test-password" {
		t.Errorf("failed to parse credential pair")
	}
}
