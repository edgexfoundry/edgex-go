/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright 2018 Joan Duran
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
 *******************************************************************************/

package general

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

const (
	TestUnexpectedMsgFormatStr = "unexpected result, active: '%s' but expected: '%s'"
)

type mockGeneralEndpoint struct {
}

func (e mockGeneralEndpoint) Monitor(params types.EndpointParams, ch chan string) {
}

func TestGetConfig(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodGet {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodGet)
		}
		if r.URL.EscapedPath() != clients.ApiConfigRoute {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), clients.ApiConfigRoute)
		}
	}))

	defer ts.Close()

	url := ts.URL

	params := types.EndpointParams{
		ServiceKey:  internal.SystemManagementAgentServiceKey,
		Path:        "/",
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault,
	}

	mc := NewGeneralClient(params, mockGeneralEndpoint{})

	responseJSON, err := mc.FetchConfiguration(context.Background())
	if err != nil {
		t.Errorf("Fetched this for its configuration: {%v}", responseJSON)
	}
}

func TestGetMetrics(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ 'status' : 'OK' }"))
		if r.Method != http.MethodGet {
			t.Errorf(TestUnexpectedMsgFormatStr, r.Method, http.MethodGet)
		}
		if r.URL.EscapedPath() != clients.ApiMetricsRoute {
			t.Errorf(TestUnexpectedMsgFormatStr, r.URL.EscapedPath(), clients.ApiMetricsRoute)
		}
	}))

	defer ts.Close()

	url := ts.URL

	params := types.EndpointParams{
		ServiceKey:  internal.SystemManagementAgentServiceKey,
		Path:        "/",
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault,
	}

	mc := NewGeneralClient(params, mockGeneralEndpoint{})

	responseJSON, err := mc.FetchMetrics(context.Background())
	if err != nil {
		t.Errorf("Fetched this for its configuration: {%v}", responseJSON)
	}
}
