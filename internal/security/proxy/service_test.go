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
	"time"
)

type testServiceRequestor struct {
	ProxyBaseURL string
}

func (tsr *testServiceRequestor) GetProxyBaseURL() string {
	return tsr.ProxyBaseURL
}

func (tsr *testServiceRequestor) GetSecretSvcBaseURL() string {
	return tsr.ProxyBaseURL
}

func (tsr *testServiceRequestor) GetHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

type testServiceCertCfg struct {
}

func (tsc *testServiceCertCfg) GetCertPath() string {
	return ""
}

func (tsc *testServiceCertCfg) GetTokenPath() string {
	return ""
}

type testServiceConfig struct {
}

func (ts *testServiceConfig) GetProxyAuthMethod() string {
	return ""
}

func (ts *testServiceConfig) GetProxyAuthTTL() int {
	return 0
}

func (ts *testServiceConfig) GetProxyAuthResource() string {
	return "all"
}

func (ts *testServiceConfig) GetProxyACLName() string {
	return ""
}

func (ts *testServiceConfig) GetProxyACLWhiteList() string {
	return ""
}

func (ts *testServiceConfig) GetSecretSvcSNIS() string {
	return ""
}

func (ts *testServiceConfig) GetEdgeXSvcs() map[string]service {
	return nil
}

func TestCheckServiceStatus(t *testing.T) {
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

	svc := Service{&testServiceRequestor{ts.URL}, &testServiceCertCfg{}, &testServiceConfig{}}
	err := svc.checkServiceStatus(ts.URL)
	if err != nil {
		t.Errorf("failed to check service status")
		t.Errorf(err.Error())
	}
}

func TestInitKongService(t *testing.T) {
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

	tk := &KongService{"test", "test", "80", "http"}
	svc := Service{&testServiceRequestor{ts.URL}, &testServiceCertCfg{}, &testServiceConfig{}}
	err := svc.initKongService(tk)
	if err != nil {
		t.Errorf("failed to initialize service")
		t.Errorf(err.Error())
	}
}

func TestInitKongRoutes(t *testing.T) {
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

	svc := Service{&testServiceRequestor{ts.URL}, &testServiceCertCfg{}, &testServiceConfig{}}
	kr := &KongRoute{}
	err := svc.initKongRoutes(kr, path)
	if err != nil {
		t.Errorf("failed to initialize route")
		t.Errorf(err.Error())
	}
}

func TestInitACL(t *testing.T) {
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

	svc := Service{&testServiceRequestor{ts.URL}, &testServiceCertCfg{}, &testServiceConfig{}}

	err := svc.initACL("test", "testgroup")
	if err != nil {
		t.Errorf("failed to initialize acl")
		t.Errorf(err.Error())
	}
}

func TestGetSvcIDs(t *testing.T) {
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

	svc := Service{&testServiceRequestor{ts.URL}, &testServiceCertCfg{}, &testServiceConfig{}}

	coll, err := svc.getSvcIDs("test")
	if err != nil {
		t.Errorf("failed to get service IDs")
		t.Errorf(err.Error())
	}
	if coll.Section[0].ID != "test-id-1" {
		t.Errorf("failed to get service ID test-id-1")
	}
}
