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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"

	"github.com/pelletier/go-toml"
)

func TestPostCertExists(t *testing.T) {
	fileName := "./testdata/configuration.toml"
	contents, err := os.ReadFile(fileName)
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
		_, _ = w.Write([]byte(`"edgex-kong already associated with existing certificate"`))
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}
	configuration.KongURL = config.KongUrlInfo{
		Server:          host,
		ApplicationPort: port,
	}

	mockLogger := logger.MockLogger{}
	service := NewService(NewRequestor(true, 10, "", mockLogger), mockLogger, configuration)
	mockCertPair := bootstrapConfig.CertKeyPair{Cert: "test-certificate", Key: "test-private-key"}
	e := service.postCert(mockCertPair)
	if e == nil {
		t.Errorf("failed on testing existing certificate on proxy - failed to catch bad request")
		return
	}
	if e.reason != CertExisting {
		t.Errorf("failed on testing existing certificate on proxy - failed to catch message of existing cert")
	}
}

func TestPostCertHttpError(t *testing.T) {
	fileName := "./testdata/configuration.toml"
	contents, err := os.ReadFile(fileName)
	require.NoError(t, err)

	configuration := &config.ConfigurationStruct{}
	err = toml.Unmarshal(contents, configuration)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	// Setting a mal-formatted hostname and port to force error.
	configuration.KongURL = config.KongUrlInfo{
		Server:          "{}",
		ApplicationPort: -1,
	}

	mockLogger := logger.MockLogger{}
	service := NewService(NewRequestor(true, 10, "", mockLogger), mockLogger, configuration)
	mockCertPair := bootstrapConfig.CertKeyPair{Cert: "test-certificate", Key: "test-private-key"}
	e := service.postCert(mockCertPair)
	require.Error(t, e)
	if e.reason != CertExisting {
		assert.Contains(t, e.Error(), "/admin/certificates")
	}
}

func TestInit(t *testing.T) {
	fileName := "./testdata/configuration.toml"
	contents, err := os.ReadFile(fileName)
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

		_, err := io.ReadAll(r.Body)
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
		Server:          host,
		ApplicationPort: port,
	}

	configuration.KongAuth.JWTFile = "./testdata/test-jwt"

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
			expectNumRoutes:    7,
		},
		{
			name:               "add one unique route",
			proxyRouteEnvValue: "testService.http://edgex-testService:12345",
			expectNumRoutes:    8,
		},
		{
			name:               "add one unique route with URL containing IP address",
			proxyRouteEnvValue: "testService.http://127.0.0.1:12345",
			expectNumRoutes:    8,
		},
		{
			name:               "add two unique routes",
			proxyRouteEnvValue: "testService1.http://edgex-testService1:12345, testService2.http://edgex-testService2:12346",
			expectNumRoutes:    9,
		},
		{
			name:               "add two unique routes:one with URL containing IP address; the other one with hostname",
			proxyRouteEnvValue: "testService1.http://127.0.0.1:12345, testService2.http://edgex-testService2:12346",
			expectNumRoutes:    9,
		},
		{
			name:               "add one duplicate route",
			proxyRouteEnvValue: "CoreData.http://edgex-core-data:59880",
			expectNumRoutes:    7,
		},
		{
			name:               "add one unique, one duplicate route",
			proxyRouteEnvValue: "testServcie.https://edgex-test-servcie1:12345, CoreData.http://edgex-core-data:59880",
			expectNumRoutes:    8,
		},
		{
			name:               "add one unique, multiple duplicate routes",
			proxyRouteEnvValue: "testServcie.https://edgex-test-servcie1:12345, CoreData.http://edgex-core-data:59880, Command.https://edgex-core-command:59882",
			expectNumRoutes:    8,
		},
		{
			name:               "add two unique, multiple duplicate routes",
			proxyRouteEnvValue: "testService1.http://127.0.0.1:12345, testService2.http://edgex-testService2:12346, CoreData.http://edgex-core-data:59880, Command.https://edgex-core-command:59882",
			expectNumRoutes:    9,
		},
		// invalid syntax tests:
		// the bad one is not added into kong route pool
		{
			name:               "bad spec without dot . in the definition for route",
			proxyRouteEnvValue: "testServcie=https://edgex-test-servcie1:12345",
			expectNumRoutes:    7,
		},
		{
			name:               "bad URL without full quallified URL",
			proxyRouteEnvValue: "testServcie.edgex-test-servcie1:12345",
			expectNumRoutes:    7,
		},
		{
			name:               "empty service name",
			proxyRouteEnvValue: ".https://edgex-test-servcie1:12345",
			expectNumRoutes:    7,
		},
	}
	for _, tt := range tests {
		currentTest := tt
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(AddProxyRoutesEnv, currentTest.proxyRouteEnvValue)
			service := NewService(NewRequestor(true, 10, "", mockLogger), mockLogger, configuration)

			err = service.Init()

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

		if r.URL.EscapedPath() != "/admin/services" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		body, err := io.ReadAll(r.Body)
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
		Server:          host,
		ApplicationPort: port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.KongURL.ApplicationPort = -1

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
			tk := &models.KongService{
				Name:     tt.serviceId,
				Protocol: "http",
				Host:     "test",
				Port:     80,
			}
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

		relativePath := "/admin/services/test/routes"
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
		Server:          host,
		ApplicationPort: port,
	}

	cfgInvalidPort := cfgOK
	cfgInvalidPort.KongURL.ApplicationPort = -1

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
			err = svc.initKongRoutes(&models.KongRoute{}, tt.path)
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
			_, _ = w.Write([]byte(`{"data": [ {"id": "test-id-1"}, {"id": "test-id-2"}]}`))
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
		Server:          host,
		ApplicationPort: port,
	}

	cfgWrongPort := cfgOK
	cfgWrongPort.KongURL.ApplicationPort = -1

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

		if r.URL.EscapedPath() == "/admin/badjson" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"this is" invalid JSON]`))
			return
		}

		if r.URL.EscapedPath() != "/admin/testservice" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data": [ {"id": "test-id-1"}, {"id": "test-id-2"}]}`))
	}))
	defer ts.Close()

	host, port, err := parseHostAndPort(ts, t)
	if err != nil {
		t.Error(err.Error())
		return
	}

	cfgOK := config.ConfigurationStruct{}
	cfgOK.KongURL = config.KongUrlInfo{
		Server:          host,
		ApplicationPort: port,
	}

	cfgWrongPort := cfgOK
	cfgWrongPort.KongURL.ApplicationPort = 123

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
