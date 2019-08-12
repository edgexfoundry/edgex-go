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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestDelete(t *testing.T) {
	LoggingClient = logger.MockLogger{}
	path := "services"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != "/services/1" {
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

	client := &http.Client{}
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
		r           Resource
		expectError bool
	}{
		{"DeleteOK", cfgOK, NewResource("1", client), false},
		{"InvalidResource", cfgOK, NewResource("2", client), true},
		{"WrongPort", cfgWrongPort, NewResource("1", client), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration = &tt.config
			err = tt.r.Remove(path)
			if err != nil && !tt.expectError {
				t.Error(err)
			}

			if err == nil && tt.expectError {
				t.Error("error was expected, none occurred")
			}
		})
	}
}
