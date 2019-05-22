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
	"fmt"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding transmission: " + err.Error())
			return
		}

		LoggingClient.Info("Posting Transmission: " + t.String())
		id, err := dbClient.AddTransmission(t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))

		break
	}
}

func transmissionBySlugHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbClient.GetTransmissionsByNotificationSlug(vars["slug"], limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		encode(t, w)
	}
}

func transmissionBySlugAndStartEndHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(fmt.Sprintf("failed to parse start %s %s", vars["start"], err.Error()))
		return
	}
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(fmt.Sprintf("failed to parse end %s %s", vars["end"], err.Error()))
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(fmt.Sprintf("failed to parse limit %s %s", vars["limit"], err.Error()))
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbClient.GetTransmissionsByNotificationSlugAndStartEnd(slug, start, end, limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the start to an integer")
		return
	}
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the end to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbClient.GetTransmissionsByStartEnd(start, end, limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the start to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:

		t, err := dbClient.GetTransmissionsByStart(start, limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the end to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbClient.GetTransmissionsByEnd(end, limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
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
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		t, err := dbClient.GetTransmissionsByStatus(limitNum, status)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Transmission not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the age to an integer")
		return
	}

	switch r.Method {
	case http.MethodDelete:

		err := dbClient.DeleteTransmission(age, status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}
