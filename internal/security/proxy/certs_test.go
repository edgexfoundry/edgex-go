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
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/proxy/mocks"
	"github.com/stretchr/testify/mock"
)

type testRequestor struct {
	SecretSvcBaseURL string
}

func (tr *testRequestor) GetProxyBaseURL() string {
	return "test"
}

func (tr *testRequestor) GetSecretSvcBaseURL() string {
	return tr.SecretSvcBaseURL
}

func (tr *testRequestor) GetHTTPClient() *http.Client {
	return &http.Client{}
}

type testCertCfg struct {
	CertPath string
}

func (tc *testCertCfg) GetCertPath() string {
	return tc.CertPath
}

func (tc *testCertCfg) GetTokenPath() string {
	return "test"
}

func createRequestorMockHttpOK() Requestor {
	response := &http.Response{StatusCode:http.StatusOK}
	req := &mocks.Requestor{}
	req.On("Do", mock.Anything).Return(response)
	return req
}

func TestGetSecret(t *testing.T) {
	r := createRequestorMockHttpOK()
	path := "testdata/test-resp-init.json"
	cs := NewCerts(r, "", "")
	s, err := cs.getSecret(path)
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
	certPath := "testCertPath"
	token := "token"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	//r := createRequestorMockHttpOK()
	cs := NewCerts(&http.Client{}, ts.URL, certPath)
	cp, err := cs.retrieve(token)
	if err != nil {
		t.Errorf("failed to retrieve cert pair")
		t.Errorf(err.Error())
	}
	if cp.Cert != "test-certificate" || cp.Key != "test-private-key" {
		t.Errorf("failed to parse certificate key pair")
	}
}
