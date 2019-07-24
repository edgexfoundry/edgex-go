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

type testDeleteRequestor struct {
	ProxyBaseURL  string
	SecretBaseURL string
}

func (tr *testDeleteRequestor) GetProxyBaseURL() string {
	return tr.ProxyBaseURL
}

func (tr *testDeleteRequestor) GetSecretSvcBaseURL() string {
	return tr.SecretBaseURL
}

func (tr *testDeleteRequestor) GetHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}
func TestDelete(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	path := "services"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/1" {
			t.Errorf("expected request to /1, got %s instead", r.URL.EscapedPath())
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

	rc := NewResource("1", &http.Client{})
	err = rc.Remove(path)
	if err != nil {
		t.Errorf("failed to delete resource")
		t.Errorf(err.Error())
	}

}
