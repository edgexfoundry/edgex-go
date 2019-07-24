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

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestCheckServiceStatus(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/" {
			t.Errorf("expected request to /, got %s instead", r.URL.EscapedPath())
		}
	}))
	defer ts.Close()

	svc := NewService(&http.Client{})
	err := svc.checkServiceStatus(ts.URL)
	if err != nil {
		t.Errorf("failed to check service status")
		t.Errorf(err.Error())
	}
}

func TestInitKongService(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/services/" {
			t.Errorf("expected request to /services, got %s instead", r.URL.EscapedPath())
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

	tk := &KongService{"test", "test", 80, "http"}
	svc := NewService(&http.Client{})
	err = svc.initKongService(tk)
	if err != nil {
		t.Errorf("failed to initialize service")
		t.Errorf(err.Error())
	}
}

func TestInitKongRoutes(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	path := "test"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s instead", r.Method)
		}

		relativePath := fmt.Sprintf("/services/%s/routes", path)
		if r.URL.EscapedPath() != relativePath {
			t.Errorf("expected request to /services, got %s instead", r.URL.EscapedPath())
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

	svc := NewService(&http.Client{})
	kr := &KongRoute{}
	err = svc.initKongRoutes(kr, path)
	if err != nil {
		t.Errorf("failed to initialize route")
		t.Errorf(err.Error())
	}
}

func TestInitACL(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/plugins/" {
			t.Errorf("expected request to /plugins/, got %s instead", r.URL.EscapedPath())
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

	svc := NewService(&http.Client{})

	err = svc.initACL("test", "testgroup")
	if err != nil {
		t.Errorf("failed to initialize acl")
		t.Errorf(err.Error())
	}
}

func TestGetSvcIDs(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [ {"id": "test-id-1"}, {"id": "test-id-2"}]}`))
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/test" {
			t.Errorf("expected request to /test, got %s instead", r.URL.EscapedPath())
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

	svc := NewService(&http.Client{})

	coll, err := svc.getSvcIDs("test")
	if err != nil {
		t.Errorf("failed to get service IDs")
		t.Errorf(err.Error())
	}
	if coll.Section[0].ID != "test-id-1" {
		t.Errorf("failed to get service ID test-id-1")
	}
}
