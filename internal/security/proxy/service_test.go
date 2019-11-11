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
 *******************************************************************************/
package proxy

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type mockCertificateLoader struct{}

func (m mockCertificateLoader) Load() (*CertPair, error) {
	return &CertPair{"test-certificate", "test-private-key"}, nil
}

func TestPostCertExists(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	fileName := "./testdata/configuration.toml"
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Errorf("could not load configuration file (%s): %s", fileName, err.Error())
		return
	}

	Configuration = &ConfigurationStruct{}
	err = toml.Unmarshal(contents, Configuration)
	if err != nil {
		t.Errorf("unable to parse configuration file (%s): %s", fileName, err.Error())
		return
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"edgex-kong already associated with existing certificate"`))
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}
	Configuration.KongURL = KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	mock := mockCertificateLoader{}
	service := NewService(NewRequestor(true, 10))
	err, existed := service.postCert(mock)
	if err != nil {
		t.Errorf(err.Error())
	}
	if existed == false {
		t.Errorf("failed on testing existing certificate on proxy")
	}
}

func TestInit(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	fileName := "./testdata/configuration.toml"
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Errorf("could not load configuration file (%s): %s", fileName, err.Error())
		return
	}

	Configuration = &ConfigurationStruct{}
	err = toml.Unmarshal(contents, Configuration)
	if err != nil {
		t.Errorf("unable to parse configuration file (%s): %s", fileName, err.Error())
		return
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if !strings.Contains(r.URL.EscapedPath(), ServicesPath) &&
			!strings.Contains(r.URL.EscapedPath(), CertificatesPath) &&
			!strings.Contains(r.URL.EscapedPath(), PluginsPath) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}
	Configuration.KongURL = KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	mock := mockCertificateLoader{}
	service := NewService(NewRequestor(true, 10))
	err = service.Init(mock)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestCheckServiceStatus(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.EscapedPath() != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{"checkOK", ts.URL, false},
		{"InvalidURL", "invalid", true},
		{"WrongPort", "http://127.0.0.1:0", true},
		{"WrongPath", ts.URL + "/test", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&http.Client{})
			err := svc.checkServiceStatus(tt.url)
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}

func TestInitKongService(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.EscapedPath() != "/services" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		recvd := string(body)
		if recvd == "host=test&name=conflict&port=80&protocol=http" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if recvd != "host=test&name=test&port=80&protocol=http" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}
	cfgOK := ConfigurationStruct{}
	cfgOK.KongURL = KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.KongURL.AdminPort = -1

	cfgInvalidHost := cfgOK
	cfgInvalidHost.KongURL.Server = ""

	tests := []struct {
		name        string
		config      ConfigurationStruct
		serviceId   string
		expectError bool
	}{
		{"serviceOK", cfgOK, "test", false},
		{"service409", cfgOK, "conflict", false},
		{"InvalidPort", cfgInvalidPort, "test", true},
		{"InvalidService", cfgOK, "invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			tk := &KongService{tt.serviceId, "test", 80, "http"}
			svc := NewService(&http.Client{})
			err = svc.initKongService(tk)

			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}

func TestInitKongRoutes(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		relativePath := "/services/test/routes"
		if r.URL.EscapedPath() != relativePath {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	cfgOK := ConfigurationStruct{}
	cfgOK.KongURL = KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.KongURL.AdminPort = -1

	tests := []struct {
		name        string
		config      ConfigurationStruct
		path        string
		expectError bool
	}{
		{"routeOK", cfgOK, "test", false},
		{"InvalidRoute", cfgOK, "invalid", true},
		{"InvalidPort", cfgInvalidPort, "test", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			svc := NewService(&http.Client{})
			err = svc.initKongRoutes(&KongRoute{}, tt.path)
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}

func TestInitACL(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		recvd := string(body)
		if recvd == "config.whitelist=testgroup&name=conflict" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if recvd != "config.whitelist=testgroup&name=test" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	cfgOK := ConfigurationStruct{}
	cfgOK.KongURL = KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.KongURL.AdminPort = -1

	tests := []struct {
		name        string
		config      ConfigurationStruct
		aclName     string
		whitelist   string
		expectError bool
	}{
		{"aclOK", cfgOK, "test", "testgroup", false},
		{"aclConflict", cfgOK, "conflict", "testgroup", false},
		{"aclInvalid", cfgOK, "invalid", "testgroup", true},
		{"InvalidPort", cfgInvalidPort, "test", "testgroup", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			svc := NewService(&http.Client{})
			err = svc.initACL(tt.aclName, tt.whitelist)
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}

func TestResetProxy(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.URL.EscapedPath(), ServicesPath) &&
			!strings.Contains(r.URL.EscapedPath(), CertificatesPath) &&
			!strings.Contains(r.URL.EscapedPath(), PluginsPath) &&
			!strings.Contains(r.URL.EscapedPath(), RoutesPath) &&
			!strings.Contains(r.URL.EscapedPath(), ConsumersPath) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": [ {"id": "test-id-1"}, {"id": "test-id-2"}]}`))
		} else if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	cfgOK := ConfigurationStruct{}
	cfgOK.KongURL = KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgWrongPort := cfgOK
	cfgWrongPort.KongURL.AdminPort = -1

	tests := []struct {
		name        string
		config      ConfigurationStruct
		expectError bool
	}{
		{"resetOK", cfgOK, false},
		{"InvalidPort", cfgWrongPort, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			svc := NewService(&http.Client{})
			err := svc.ResetProxy()
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}

func TestGetSvcIDs(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.EscapedPath() == "/badjson" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"this is" invalid JSON]`))
			return
		}

		if r.URL.EscapedPath() != "/testservice" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": [ {"id": "test-id-1"}, {"id": "test-id-2"}]}`))
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	cfgOK := ConfigurationStruct{}
	cfgOK.KongURL = KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgWrongPort := cfgOK
	cfgWrongPort.KongURL.AdminPort = 123

	validService := "testservice"

	tests := []struct {
		name        string
		config      ConfigurationStruct
		serviceId   string
		expectError bool
	}{
		{"GetOK", cfgOK, validService, false},
		{"InvalidService", cfgOK, "invalid", true},
		{"InvalidUrl", cfgWrongPort, validService, true},
		{"BadJSONResponse", cfgOK, "badjson", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			svc := NewService(&http.Client{})

			coll, err := svc.getSvcIDs(tt.serviceId)
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil {
				if tt.expectError {
					t.Error("error was expected, none occurred")
				} else if coll.Section[0].ID != "test-id-1" {
					t.Errorf("failed to get service ID test-id-1")
				}
			}
		})
	}
}

func TestInitJWTAuth(t *testing.T) {
	LoggingClient = logger.MockLogger{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if !strings.Contains(r.URL.EscapedPath(), PluginsPath) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	cfgOK := ConfigurationStruct{}
	cfgOK.KongURL = KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgWrongPort := cfgOK
	cfgWrongPort.KongURL.AdminPort = 123

	tests := []struct {
		name        string
		config      ConfigurationStruct
		expectError bool
	}{
		{"jwtOK", cfgOK, false},
		{"InvalidPort", cfgWrongPort, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			svc := NewService(&http.Client{})
			err := svc.initJWTAuth()
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}

func parseHostAndPort(server *httptest.Server, t *testing.T) (host string, port int, err error) {
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Errorf("unable to parse test server URL %s", server.URL)
		return
	}
	port, err = strconv.Atoi(parsed.Port())
	if err != nil {
		t.Errorf("parsed port number cannot be converted to int %s", parsed.Port())
		return
	}
	host = parsed.Hostname()
	return
}
