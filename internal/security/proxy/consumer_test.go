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
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type testConsumerRequestor struct {
	ProxyBaseURL string
}

func (tc *testConsumerRequestor) GetProxyBaseURL() string {
	return tc.ProxyBaseURL
}

func (tc *testConsumerRequestor) GetSecretSvcBaseURL() string {
	return tc.ProxyBaseURL
}

func (tc *testConsumerRequestor) GetHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

type testConsumerConfig struct {
	ProxyBaseURL string
}

func (te *testConsumerConfig) GetProxyServerName() string {
	return te.ProxyBaseURL
}

func (te *testConsumerConfig) GetProxyServerPort() string {
	return "8001"
}

func (te *testConsumerConfig) GetProxyApplicationPortSSL() string {
	return "8443"
}

func (te *testConsumerConfig) GetProxyAuthMethod() string {
	return "jwt"
}

func (te *testConsumerConfig) GetProxyAuthResource() string {
	return "all"
}

func TestCreate(t *testing.T) {
	LoggingClient = logger.MockLogger{}

	name := "testuser"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "PUT" {
			t.Errorf("expected PUT request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/consumers/testuser" {
			t.Errorf("expected request to /consumer, got %s instead", r.URL.EscapedPath())
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
	Configuration.KongURL = KongUrlInfo{
		Server:    parsed.Hostname(),
		AdminPort: port,
	}

	co := NewConsumer(name, &http.Client{})
	err = co.Create("test")
	if err != nil {
		t.Errorf("failed to creat consumer testuser")
		t.Errorf(err.Error())
	}
}

func TestAssociateWithGroup(t *testing.T) {
	LoggingClient = logger.MockLogger{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/consumers/testuser/acls" {
			t.Errorf("expected request to /consumers/testuser/acls, got %s instead", r.URL.EscapedPath())
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
	Configuration.KongURL = KongUrlInfo{
		Server:    parsed.Hostname(),
		AdminPort: port,
	}

	co := NewConsumer("testuser", &http.Client{})
	err = co.AssociateWithGroup("groupname")
	if err != nil {
		t.Errorf("failed to associate consumer with group")
		t.Errorf(err.Error())
	}
}

func TestCreateJWTToken(t *testing.T) {
	LoggingClient = logger.MockLogger{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"consumer_id": "test", "created_at": 1442426001000,"id": "test", "key": "test-key","secret": "test-secret"}`))
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/consumers/testuser/jwt" {
			t.Errorf("expected request to /consumers/testuser/jwt, got %s instead", r.URL.EscapedPath())
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
	Configuration.KongURL = KongUrlInfo{
		Server:    parsed.Hostname(),
		AdminPort: port,
	}

	co := NewConsumer("testuser", &http.Client{})
	_, err = co.createJWTToken()
	if err != nil {
		t.Errorf("failed to creat JWT token for consumer")
		t.Errorf(err.Error())
	}
}

/*func TestCreateOAuth2Token(t *testing.T) {
	t.Skip()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token_type": "oauth2", "access_token": "test", "expires_in": 1442426001000}`))
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/consumers/testuser/oauth2" {
			t.Errorf("expected request to /consumers/testuser/oauth2, got %s instead", r.URL.EscapedPath())
		}
	}))
	defer ts.Close()

	co := Consumer{"testuser", &testConsumerRequestor{ts.URL}, &testConsumerConfig{ts.URL}}
	_, err := co.createOAuth2Token()
	if err != nil {
		t.Errorf("failed to creat OAuth2 token for consumer")
		t.Errorf(err.Error())
	}
}*/
