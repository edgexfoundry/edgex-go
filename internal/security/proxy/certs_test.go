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
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/mock"
)

func createRequestorMockHttpOK() Requestor {
	response := &http.Response{StatusCode: http.StatusOK}
	req := &mocks.Requestor{}
	req.On("Do", mock.Anything).Return(response)
	return req
}

func TestLoad(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"cert": "test-certificate", "key": "test-private-key"}}`))
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		return
	}

	cfgOK := ConfigurationStruct{}
	cfgOK.SecretService = SecretServiceInfo{
		Server: host,
		Port:   port,
	}

	validCertPath := "testCertPath"
	validTokenPath := "testdata/test-resp-init.json"

	tests := []struct {
		name        string
		config      ConfigurationStruct
		certPath    string
		tokenPath   string
		expectError bool
	}{
		{"LoadOK", cfgOK, validCertPath, validTokenPath, false},
		{"InvalidTokenPath", cfgOK, validCertPath, "invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			cert := NewCertificateLoader(NewRequestor(true), tt.certPath, tt.tokenPath)
			_, err := cert.Load()
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}

func TestGetAccessToken(t *testing.T) {
	r := createRequestorMockHttpOK()
	path := "testdata/test-resp-init.json"
	cs := certificate{r, "", ""}
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
	pairOK := CertPair{"private-cert", "private-key"}
	pairBlank := CertPair{}
	tests := []struct {
		name        string
		pair        CertPair
		expectError bool
	}{
		{"PairOK", pairOK, false},
		{"PairBlank", pairBlank, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := certificate{r, "", ""}
			err := cs.validate(&tt.pair)
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}

func TestRetrieve(t *testing.T) {
	LoggingClient = logger.MockLogger{}

	certPath := "testCertPath"
	token := "token"
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.EscapedPath() == "/badjson" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`["bad:json}"`))
			return
		}

		if r.URL.EscapedPath() != fmt.Sprintf("/%s", certPath) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Header.Get(VaultToken) != token {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"cert": "test-certificate", "key": "test-private-key"}}`))
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		return
	}

	cfgOK := ConfigurationStruct{}
	cfgOK.SecretService = SecretServiceInfo{
		Server: host,
		Port:   port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.SecretService.Port = -1

	tests := []struct {
		name        string
		config      ConfigurationStruct
		certPath    string
		token       string
		expectError bool
	}{
		{"RetrieveOK", cfgOK, certPath, token, false},
		{"InvalidPath", cfgOK, "invalid", token, true},
		{"InvalidJSON", cfgOK, "badjson", token, true},
		{"InvalidPort", cfgInvalidPort, certPath, token, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			cs := certificate{NewRequestor(true), tt.certPath, ""}
			cp, err := cs.retrieve(tt.token)
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil {
				if tt.expectError {
					t.Error("error was expected, none occurred")
				} else if cp.Cert != "test-certificate" || cp.Key != "test-private-key" {
					t.Errorf("failed to parse certificate key pair")
				}
			}
		})
	}
}
