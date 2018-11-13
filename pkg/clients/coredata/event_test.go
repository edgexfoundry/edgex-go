/*******************************************************************************
 * Copyright 1995-2018 Hitachi Vantara Corporation. All rights reserved.
 *
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
 *******************************************************************************/
package coredata

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

const (
	TestId           = "5aae1f4fe4b0d019b26a56b8"
	TestEventDevice1 = "device1"
	TestEventDevice2 = "device2"
)

func TestMarkPushed(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPut {
			t.Errorf("expected http method is PUT, active http method is : %s", r.Method)
		}

		url := clients.ApiEventRoute + "/id/" + TestId
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}
	}))

	defer ts.Close()

	url := ts.URL + clients.ApiEventRoute

	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault}

	ec := NewEventClient(params, mockEventEndpoint{})

	err := ec.MarkPushed(TestId, context.Background())

	if err != nil {
		t.FailNow()
	}
}

func TestGetEvents(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodGet {
			t.Errorf("expected http method is GET, active http method is : %s", r.Method)
		}

		if r.URL.EscapedPath() != clients.ApiEventRoute {
			t.Errorf("expected uri path is %s, actual uri path is %s", clients.ApiEventRoute, r.URL.EscapedPath())
		}

		w.Write([]byte("[" +
			"{" +
			"\"Device\" : \"" + TestEventDevice1 + "\"" +
			"}," +
			"{" +
			"\"Device\" : \"" + TestEventDevice2 + "\"" +
			"}" +
			"]"))

	}))

	defer ts.Close()

	url := ts.URL + clients.ApiEventRoute

	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault}

	ec := NewEventClient(params, mockEventEndpoint{})

	eArr, err := ec.Events(context.Background())
	if err != nil {
		t.FailNow()
	}

	if len(eArr) != 2 {
		t.Errorf("expected event array's length is 2, actual array's length is : %d", len(eArr))
	}

	e1 := eArr[0]
	if e1.Device != TestEventDevice1 {
		t.Errorf("expected first events's device is : %s, actual device is : %s", TestEventDevice1, e1.Device)
	}

	e2 := eArr[1]
	if e2.Device != TestEventDevice2 {
		t.Errorf("expected second events's device is : %s, actual device is : %s ", TestEventDevice2, e2.Device)
	}
}

func TestNewEventClientWithConsul(t *testing.T) {
	deviceUrl := "http://localhost:48080" + clients.ApiEventRoute
	params := types.EndpointParams{
		ServiceKey:  internal.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: true,
		Url:         deviceUrl,
		Interval:    clients.ClientMonitorDefault}

	ec := NewEventClient(params, mockEventEndpoint{})

	r, ok := ec.(*EventRestClient)
	if !ok {
		t.Error("ec is not of expected type")
	}

	time.Sleep(25 * time.Millisecond)
	if len(r.url) == 0 {
		t.Error("url was not initialized")
	} else if r.url != deviceUrl {
		t.Errorf("unexpected url value %s", r.url)
	}
}

type mockEventEndpoint struct {
}

func (e mockEventEndpoint) Monitor(params types.EndpointParams, ch chan string) {
	switch params.ServiceKey {
	case internal.CoreDataServiceKey:
		url := fmt.Sprintf("http://%s:%v%s", "localhost", 48080, params.Path)
		ch <- url
		break
	default:
		ch <- ""
	}
}
