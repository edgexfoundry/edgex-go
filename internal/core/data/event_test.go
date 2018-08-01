/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package data

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db/memory"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

var testEvent models.Event
var testRoutes *mux.Router

func TestMain(m *testing.M) {
	testEvent.Device = "test device"
	testEvent.Origin = 123456789
	dbClient = &memory.MemDB{}
	testEvent.ID, _ = dbClient.AddEvent(&testEvent)
	testRoutes = LoadRestRoutes()
	LoggingClient = logger.NewMockClient()
	os.Exit(m.Run())
}

//Test methods
func TestGetEventHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/event", nil)
	w := httptest.NewRecorder()
	configuration.ReadMaxLimit = 1
	testRoutes.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Error("value expected, status code " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}

	if len(w.Body.String()) == 0 {
		t.Error("response was empty " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}

	events := []models.Event{}
	json.Unmarshal([]byte(w.Body.String()), &events)
	for e := range events {
		testEventWithoutReadings(events[e], t)
	}
	configuration.ReadMaxLimit = 0
}

func TestGetEventHandlerMaxExceeded(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/event", nil)
	w := httptest.NewRecorder()
	configuration.ReadMaxLimit = 0
	testRoutes.ServeHTTP(w, req)

	if w.Code != 413 {
		t.Error("413 exceeded, status code " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}
}

func TestGetEventByIdHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/event/"+testEvent.ID.Hex(), nil)
	w := httptest.NewRecorder()

	testRoutes.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Error("value expected, status code " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}

	if len(w.Body.String()) == 0 {
		t.Error("response was empty " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
		return
	}

	event := models.Event{}
	json.Unmarshal([]byte(w.Body.String()), &event)
	testEventWithoutReadings(event, t)
}

func testEventWithoutReadings(event models.Event, t *testing.T) {
	if event.ID.Hex() != testEvent.ID.Hex() {
		t.Error("eventId mismatch. expected " + testEvent.ID.Hex() + " received " + event.ID.Hex())
	}

	if event.Device != testEvent.Device {
		t.Error("device mismatch. expected " + testEvent.Device + " received " + event.Device)
	}

	if event.Origin != testEvent.Origin {
		t.Error("origin mismatch. expected " + strconv.FormatInt(testEvent.Origin, 10) + " received " + strconv.FormatInt(event.Origin, 10))
	}
}
