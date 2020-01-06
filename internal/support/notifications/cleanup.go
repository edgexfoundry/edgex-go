/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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

package notifications

import (
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/gorilla/mux"
)

func cleanupHandler(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	lc.Info("Cleaning up of notifications and transmissions")
	cleanupHandlerCloser(w, dbClient.Cleanup(), lc)
}

func cleanupAgeHandler(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}
	vars := mux.Vars(r)
	age, err := strconv.Atoi(vars["age"])
	// Problem converting age
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting the age to an integer")
		return
	}

	lc.Info("Cleaning up of notifications and transmissions")
	cleanupHandlerCloser(w, dbClient.CleanupOld(age), lc)
}

func cleanupHandlerCloser(w http.ResponseWriter, err error, lc logger.LoggingClient) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error(err.Error())
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("true"))
}
