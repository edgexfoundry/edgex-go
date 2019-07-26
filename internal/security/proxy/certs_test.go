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
 * @author: Tingyu Zeng, Dell
 * @version: 1.1.0
 *******************************************************************************/
package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/proxy/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/mock"
)

func createRequestorMockHttpOK() Requestor {
	response := &http.Response{StatusCode: http.StatusOK}
	req := &mocks.Requestor{}
	req.On("Do", mock.Anything).Return(response)
	return req
}

func TestGetAccessToken(t *testing.T) {
	r := createRequestorMockHttpOK()
	path := "testdata/test-resp-init.json"
	cs := NewCerts(r, "", "")
	s, err := cs.getAccessToken(path)
	if err != nil {
		t.Errorf("failed to parse token file")
		t.Errorf(err.Error())
	}
	if s != "test-token" {
		t.Errorf("incorrect token")
		t.Errorf(s)
	}
}

func TestValidate(t *testing.T) {
	r := createRequestorMockHttpOK()
	cp := &CertPair{"private-cert", "private-key"}
	cs := NewCerts(r, "", "")
	err := cs.validate(cp)
	if err != nil {
		t.Errorf("failed to validate cert collection")
	}
}

func TestRetrieve(t *testing.T) {
	LoggingClient = logger.MockLogger{}

	certPath := "testCertPath"
	token := "token"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"cert": "test-certificate", "key": "test-private-key"}}`))
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != fmt.Sprintf("/%s", certPath) {
			t.Errorf("expected request to /%s, got %s instead", certPath, r.URL.EscapedPath())
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
	Configuration = &ConfigurationStruct{}
	Configuration.SecretService = SecretServiceInfo{
		Server: parsed.Hostname(),
		Port:   port,
	}

	cs := NewCerts(NewRequestor(true), certPath, "")
	cp, err := cs.retrieve(token)
	if err != nil {
		t.Errorf("failed to retrieve cert pair")
		t.Errorf(err.Error())
	}
	if cp.Cert != "test-certificate" || cp.Key != "test-private-key" {
		t.Errorf("failed to parse certificate key pair")
	}
	Configuration = &ConfigurationStruct{}
}
