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
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/BurntSushi/toml"
)

func TestPostCertExists(t *testing.T) {
	fileName := "./testdata/configuration.toml"
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Errorf("could not load configuration file (%s): %s", fileName, err.Error())
		return
	}

	configuration := &config.ConfigurationStruct{}
	err = toml.Unmarshal(contents, configuration)
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
	configuration.KongURL = config.KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	mockLogger := logger.MockLogger{}
	service := NewService(NewRequestor(true, 10, "", mockLogger), mockLogger, configuration)
	mockCertPair := bootstrapConfig.CertKeyPair{Cert: "test-certificate", Key: "test-private-key"}
	e := service.postCert(mockCertPair)
	if e == nil {
		t.Errorf("failed on testing existing certificate on proxy - failed to catch bad request")
	}
	if e.reason != CertExisting {
		t.Errorf("failed on testing existing certificate on proxy - failed to catch message of existing cert")
	}
}

func TestInit(t *testing.T) {
	fileName := "./testdata/configuration.toml"
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Errorf("could not load configuration file (%s): %s", fileName, err.Error())
		return
	}

	configuration := &config.ConfigurationStruct{}
	err = toml.Unmarshal(contents, configuration)
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
	configuration.KongURL = config.KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	mockCertPair := bootstrapConfig.CertKeyPair{Cert: "test-certificate", Key: "test-private-key"}
	mockLogger := logger.MockLogger{}

	origProxyRoutEnv := os.Getenv(AddProxyRoutesEnv)
	defer func() {
		_ = os.Setenv(AddProxyRoutesEnv, origProxyRoutEnv)
	}()

	tests := []struct {
		name               string
		proxyRouteEnvValue string
		expectNumRoutes    int
	}{
		{ // the original number of routes is from configuration file
			name:               "empty env",
			proxyRouteEnvValue: "",
			expectNumRoutes:    8,
		},
		{
			name:               "add one unique route",
			proxyRouteEnvValue: "testService.http://edgex-testService:12345",
			expectNumRoutes:    9,
		},
		{
			name:               "add two unique routes",
			proxyRouteEnvValue: "testService1.http://edgex-testService1:12345, testService2.http://edgex-testService2:12346",
			expectNumRoutes:    10,
		},
		{
			name:               "add one duplicate route",
			proxyRouteEnvValue: "CoreData.http://edgex-core-data:48080",
			expectNumRoutes:    8,
		},
		{
			name:               "add one unique, one duplicate route",
			proxyRouteEnvValue: "testServcie.https://edgex-test-servcie1:12345, CoreData.http://edgex-core-data:48080",
			expectNumRoutes:    9,
		},
		{
			name:               "add one unique, multiple duplicate routes",
			proxyRouteEnvValue: "testServcie.https://edgex-test-servcie1:12345, CoreData.http://edgex-core-data:48080, Command.https://edgex-core-command:48082",
			expectNumRoutes:    9,
		},
		// invalid syntax tests:
		// the bad one is not added into kong route pool
		{
			name:               "bad spec without dot . in the definition for route",
			proxyRouteEnvValue: "testServcie=https://edgex-test-servcie1:12345",
			expectNumRoutes:    8,
		},
		{
			name:               "bad URL without full quallified URL",
			proxyRouteEnvValue: "testServcie.edgex-test-servcie1:12345",
			expectNumRoutes:    8,
		},
		{
			name:               "empty service name",
			proxyRouteEnvValue: ".https://edgex-test-servcie1:12345",
			expectNumRoutes:    8,
		},
	}
	for _, tt := range tests {
		currentTest := tt
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(AddProxyRoutesEnv, currentTest.proxyRouteEnvValue)
			service := NewService(NewRequestor(true, 10, "", mockLogger), mockLogger, configuration)

			err = service.Init(mockCertPair)

			require.NoError(t, err)

			assert.Equal(t, currentTest.expectNumRoutes, len(service.routes), "number of Kong routes not the same")
		})
	}
}

func TestCheckServiceStatus(t *testing.T) {
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
		configuration := &config.ConfigurationStruct{}
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&http.Client{}, logger.MockLogger{}, configuration)
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
	cfgOK := config.ConfigurationStruct{}
	cfgOK.KongURL = config.KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.KongURL.AdminPort = -1

	cfgInvalidHost := cfgOK
	cfgInvalidHost.KongURL.Server = ""

	tests := []struct {
		name        string
		config      config.ConfigurationStruct
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
			tk := &KongService{tt.serviceId, "test", 80, "http"}
			svc := NewService(&http.Client{}, logger.MockLogger{}, &tt.config)
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

	cfgOK := config.ConfigurationStruct{}
	cfgOK.KongURL = config.KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.KongURL.AdminPort = -1

	tests := []struct {
		name        string
		config      config.ConfigurationStruct
		path        string
		expectError bool
	}{
		{"routeOK", cfgOK, "test", false},
		{"InvalidRoute", cfgOK, "invalid", true},
		{"InvalidPort", cfgInvalidPort, "test", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&http.Client{}, logger.MockLogger{}, &tt.config)
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

	cfgOK := config.ConfigurationStruct{}
	cfgOK.KongURL = config.KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.KongURL.AdminPort = -1

	tests := []struct {
		name        string
		config      config.ConfigurationStruct
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
			svc := NewService(&http.Client{}, logger.MockLogger{}, &tt.config)
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

	cfgOK := config.ConfigurationStruct{}
	cfgOK.KongURL = config.KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgWrongPort := cfgOK
	cfgWrongPort.KongURL.AdminPort = -1

	tests := []struct {
		name        string
		config      config.ConfigurationStruct
		expectError bool
	}{
		{"resetOK", cfgOK, false},
		{"InvalidPort", cfgWrongPort, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&http.Client{}, logger.MockLogger{}, &tt.config)
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

	cfgOK := config.ConfigurationStruct{}
	cfgOK.KongURL = config.KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgWrongPort := cfgOK
	cfgWrongPort.KongURL.AdminPort = 123

	validService := "testservice"

	tests := []struct {
		name        string
		config      config.ConfigurationStruct
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
			svc := NewService(&http.Client{}, logger.MockLogger{}, &tt.config)

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

	cfgOK := config.ConfigurationStruct{}
	cfgOK.KongURL = config.KongUrlInfo{
		Server:    host,
		AdminPort: port,
	}

	cfgWrongPort := cfgOK
	cfgWrongPort.KongURL.AdminPort = 123

	tests := []struct {
		name        string
		config      config.ConfigurationStruct
		expectError bool
	}{
		{"jwtOK", cfgOK, false},
		{"InvalidPort", cfgWrongPort, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&http.Client{}, logger.MockLogger{}, &tt.config)
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
