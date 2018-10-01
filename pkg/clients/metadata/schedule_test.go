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

// Test adding a schedule using the client
func TestAddSchedule(t *testing.T) {
	d := models.Schedule{
		Id:        "1234",
		Name:      "Test name for schedule",
		Start:     "", // defaults to now
		End:       "", // defaults to ZDT MAX
		Frequency: "PT30S",
	}

	addingScheduleId := d.Id.Hex()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPost {
			t.Errorf("expected http method is %s, active http method is : %s", http.MethodPost, r.Method)
		}

		if r.URL.EscapedPath() != clients.ApiScheduleRoute {
			t.Errorf("expected uri path is %s, actual uri path is %s", clients.ApiScheduleRoute, r.URL.EscapedPath())
		}

		w.Write([]byte(addingScheduleId))

	}))

	defer ts.Close()

	url := ts.URL + clients.ApiScheduleRoute

	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiScheduleRoute,
		UseRegistry: false,
		Url:         url,
		Interval:    clients.ClientMonitorDefault}
	sc := NewScheduleClient(params, MockEndpoint{})

	receivedScheduleId, err := sc.Add(&d)
	if err != nil {
		t.Error(err.Error())
	}

	if receivedScheduleId != addingScheduleId {
		t.Errorf("expected schedule id : %s, actual schedule id : %s", receivedScheduleId, addingScheduleId)
	}
}

func TestNewScheduleClientWithConsul(t *testing.T) {
	scheduleUrl := "http://localhost:48081" + clients.ApiScheduleRoute
	params := types.EndpointParams{
		ServiceKey:  internal.CoreMetaDataServiceKey,
		Path:        clients.ApiScheduleRoute,
		UseRegistry: true,
		Url:         scheduleUrl,
		Interval:    clients.ClientMonitorDefault}

	sc := NewScheduleClient(params, MockEndpoint{})

	r, ok := sc.(*ScheduleRestClient)
	if !ok {
		t.Error("sc is not of expected type")
	}

	time.Sleep(25 * time.Millisecond)
	if len(r.url) == 0 {
		t.Error("url was not initialized")
	} else if r.url != scheduleUrl {
		t.Errorf("unexpected url value %s", r.url)
	}
}
