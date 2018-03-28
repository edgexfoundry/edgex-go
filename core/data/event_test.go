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
 *
 * @author: Trevor Conn, Dell
 * @version: 0.5.0
 *******************************************************************************/

package data

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/core/data/clients"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
	"github.com/gorilla/mux"
)

var globalMockParams *clients.MockParams
var testRoutes *mux.Router

func TestMain(m *testing.M) {
	globalMockParams = clients.NewMockParams()
	dbc, _ = clients.NewDBClient(clients.DBConfiguration{DbType: clients.MOCK})
	testRoutes = LoadRestRoutes()
	loggingClient = logger.NewMockClient()
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
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/event/"+globalMockParams.EventId.Hex(), nil)
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
	if event.ID.Hex() != globalMockParams.EventId.Hex() {
		t.Error("eventId mismatch. expected " + globalMockParams.EventId.Hex() + " received " + event.ID.Hex())
	}

	if event.Device != globalMockParams.Device {
		t.Error("device mismatch. expected " + globalMockParams.Device + " received " + event.Device)
	}

	if event.Origin != globalMockParams.Origin {
		t.Error("origin mismatch. expected " + strconv.FormatInt(globalMockParams.Origin, 10) + " received " + strconv.FormatInt(event.Origin, 10))
	}
}
