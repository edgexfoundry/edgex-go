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
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/support/notifications/clients"
	"github.com/edgexfoundry/edgex-go/support/notifications/models"
	"github.com/gorilla/mux"
)

func transmissionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	case http.MethodPost:
		var t models.Transmission
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&t)

		// Problem Decoding Transmission
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding transmission: " + err.Error())
			return
		}

		loggingClient.Info("Posting Transmission: " + t.String())
		id, err := dbc.AddTransmission(&t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id.Hex()))

		break
	}
}

func transmissionBySlugHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	resendLimit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbc.TransmissionsByNotificationSlug(vars["slug"], resendLimit)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(t, w)
	}
}

func transmissionByStartEndHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	// Problem converting start
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the start to an integer")
		return
	}
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the end to an integer")
		return
	}
	resendLimit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbc.TransmissionsByStartEnd(start, end, resendLimit)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(t, w)
	}
}

func transmissionByStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	// Problem converting start
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the start to an integer")
		return
	}
	resendLimit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:

		t, err := dbc.TransmissionsByStart(start, resendLimit)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(t, w)
	}
}

func transmissionByEndHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	// Problem converting start
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the end to an integer")
		return
	}
	resendLimit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbc.TransmissionsByEnd(end, resendLimit)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(t, w)
	}
}

func transmissionByEscalatedHandler(w http.ResponseWriter, r *http.Request) {
	transmissionByStatusHandler(w, r, models.Trxescalated)
}

func transmissionByFailedHandler(w http.ResponseWriter, r *http.Request) {
	transmissionByStatusHandler(w, r, models.Failed)
}

func transmissionByStatusHandler(w http.ResponseWriter, r *http.Request, status models.TransmissionStatus) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	resendLimit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbc.TransmissionsByStatus(resendLimit, status)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(t, w)
	}
}

func transmissionByAgeSentHandler(w http.ResponseWriter, r *http.Request) {
	transmissionByAgeStatusHandler(w, r, models.Sent)
}

func transmissionByAgeEscalatedHandler(w http.ResponseWriter, r *http.Request) {
	transmissionByAgeStatusHandler(w, r, models.Trxescalated)
}

func transmissionByAgeAcknowledgedHandler(w http.ResponseWriter, r *http.Request) {
	transmissionByAgeStatusHandler(w, r, models.Acknowledged)
}

func transmissionByAgeFailedHandler(w http.ResponseWriter, r *http.Request) {
	transmissionByAgeStatusHandler(w, r, models.Failed)
}

func transmissionByAgeStatusHandler(w http.ResponseWriter, r *http.Request, status models.TransmissionStatus) {

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

		err := dbc.DeleteTransmission(age, status)
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
