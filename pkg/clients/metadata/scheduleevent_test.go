/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (C) 2018 Canonical Ltd
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
package metadata

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

// Test adding a schedule event using the client
func TestAddScheduleEvent(t *testing.T) {
	se := models.ScheduleEvent{
		Id:          "1234",
		Name:        "Test name for schedule event",
		Schedule:    "Test name of owning schedule",
		Addressable: models.Addressable{},
		Parameters:  "{\"VDS-CurrentTemperature\": \"98.6\"}",
		Service:     "Test device service name",
	}

	addingScheduleEventId := se.Id.Hex()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPost {
			t.Errorf("expected http method is %s, active http method is : %s", http.MethodPost, r.Method)
		}

		if r.URL.EscapedPath() != clients.ApiScheduleEventRoute {
			t.Errorf("expected uri path is %s, actual uri path is %s", clients.ApiScheduleEventRoute, r.URL.EscapedPath())
		}

		w.Write([]byte(addingScheduleEventId))

	}))

	defer ts.Close()

	url := ts.URL + clients.ApiScheduleEventRoute

	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiScheduleEventRoute,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault}
	sc := NewScheduleEventClient(params, MockEndpoint{})

	receivedScheduleEventId, err := sc.Add(&se)
	if err != nil {
		t.Error(err.Error())
	}

	if receivedScheduleEventId != addingScheduleEventId {
		t.Errorf("expected schedule event id : %s, actual schedule event id : %s", receivedScheduleEventId, addingScheduleEventId)
	}
}

func TestNewScheduleEventClientWithConsul(t *testing.T) {
	scheduleEventUrl := "http://localhost:48081" + clients.ApiScheduleEventRoute
	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiScheduleEventRoute,
		UseRegistry: true,
		Url:         scheduleEventUrl,
		Interval:    clients.ClientMonitorDefault}

	sc := NewScheduleEventClient(params, MockEndpoint{})

	r, ok := sc.(*ScheduleEventRestClient)
	if !ok {
		t.Error("sc is not of expected type")
	}

	time.Sleep(25 * time.Millisecond)
	if len(r.url) == 0 {
		t.Error("url was not initialized")
	} else if r.url != scheduleEventUrl {
		t.Errorf("unexpected url value %s", r.url)
	}
}
