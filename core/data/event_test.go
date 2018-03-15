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
	"strconv"
	"testing"

	"github.com/tsconn23/edgex-go/core/data/clients"
	"github.com/tsconn23/edgex-go/core/domain/models"
)


//Test methods
func TestGetEventByIdHandler(t *testing.T) {
	params := clients.NewMockParams()
	dbc, _ = clients.NewDBClient(clients.DBConfiguration{DbType: clients.MOCK})

	r := LoadRestRoutes()
	req, _ := http.NewRequest("GET", "/api/v1/event/" + params.EventId.Hex(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

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
	if event.ID.Hex() != params.EventId.Hex() {
		t.Error("eventId mismatch. expected " + params.EventId.Hex() + " received " + event.ID.Hex())
	}

	if event.Device != params.Device {
		t.Error("device mismatch. expected " + params.Device + " received " + event.Device)
	}

	if event.Origin != params.Origin {
		t.Error("origin mismatch. expected " + strconv.FormatInt(params.Origin,10) + " received " + strconv.FormatInt(event.Origin, 10))
	}
	//fmt.Println(w.Body.String())
}
