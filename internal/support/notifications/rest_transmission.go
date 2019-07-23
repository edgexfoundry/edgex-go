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

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

func addTransmission(t models.Transmission) (id string, err error) {
	LoggingClient.Info("Posting Transmission: " + t.String())
	id, err = dbClient.AddTransmission(t)
	if err != nil {
		LoggingClient.Error(err.Error())
		return id, err
	}
	return id, nil
}

func getTransmissionsByNotificationSlug(slug string, limit int) (t []models.Transmission, err error) {
	t, err = dbClient.GetTransmissionsByNotificationSlug(slug, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return t, err
	}
	return t, nil
}

func getTransmissionsByNotificationSlugAndStartEnd(slug string, start int64, end int64, limit int) (t []models.Transmission, err error) {
	t, err = dbClient.GetTransmissionsByNotificationSlugAndStartEnd(slug, start, end, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return t, err
	}
	return t, nil
}

func getTransmissionsByStartEnd(start int64, end int64, limit int) (t []models.Transmission, err error) {
	t, err = dbClient.GetTransmissionsByStartEnd(start, end, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return t, err
	}
	return t, nil
}

func getTransmissionsByStart(start int64, limit int) (t []models.Transmission, err error) {
	t, err = dbClient.GetTransmissionsByStart(start, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return t, err
	}
	return t, nil
}

func getTransmissionsByEnd(end int64, limit int) (t []models.Transmission, err error) {
	t, err = getTransmissionsByEnd(end, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return t, err
	}
	return t, nil
}

func getTransmissionsByStatus(limit int, status models.TransmissionStatus) (t []models.Transmission, err error) {
	t, err = dbClient.GetTransmissionsByStatus(limit, status)
	if err != nil {
		LoggingClient.Error(err.Error())
		return t, err
	}
	return t, nil
}

func deleteTransmission(age int64, status models.TransmissionStatus) error {
	err := dbClient.DeleteTransmission(age, status)
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	return nil
}

func transmissionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	var t models.Transmission
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&t)

	// Problem Decoding Transmission
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding transmission: " + err.Error())
		return
	}

	id, err := addTransmission(t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))

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

	t, err := getTransmissionsByNotificationSlug(vars["slug"], limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(t, w, LoggingClient)
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

	t, err := getTransmissionsByNotificationSlugAndStartEnd(slug, start, end, limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(t, w, LoggingClient)

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

	t, err := getTransmissionsByStartEnd(start, end, limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(t, w, LoggingClient)

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

	t, err := getTransmissionsByStart(start, limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(t, w, LoggingClient)

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

	t, err := getTransmissionsByEnd(end, limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(t, w, LoggingClient)

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

	t, err := getTransmissionsByStatus(limitNum, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(t, w, LoggingClient)

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

	err = deleteTransmission(age, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))

}
