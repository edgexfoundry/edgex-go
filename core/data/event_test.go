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
	"fmt"
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
	req, _ := http.NewRequest("GET", "/api/v1/event/" + params.EventId1.Hex(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Error("value expected, status code " + strconv.Itoa(w.Code) + " " + req.Method + " " + req.URL.Path)
	} else {
		event := models.Event{}
		json.Unmarshal([]byte(w.Body.String()), &event)
		if event.ID.Hex() != params.EventId1.Hex() {
			t.Error("value mismatch. expected " + params.EventId1.Hex() + " received " + event.ID.Hex())
		}
	}

	fmt.Println(w.Body.String())
}
