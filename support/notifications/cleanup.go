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

	"github.com/gorilla/mux"
)

func cleanupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	case http.MethodDelete:
		loggingClient.Info("Cleaning up of notifications and transmissions")

		err := dbc.Cleanup()

		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))

	}
}

func cleanupAgeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	vars := mux.Vars(r)
	age, err := strconv.ParseInt(vars["age"], 10, 64)
	// Problem converting age
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the age to an integer")
		return
	}
	switch r.Method {
	case http.MethodDelete:
		loggingClient.Info("Cleaning up of old notifications and transmissions")

		if err = dbc.CleanupOld(age); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))

	}
}
