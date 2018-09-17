//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	mux "github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
)

func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()

	//ping
	b.HandleFunc("/ping", ping).Methods(http.MethodGet)

	// callback
	b.HandleFunc("/callbacks", addCallbackAlert).Methods(http.MethodPost)
	b.HandleFunc("/callbacks", updateCallbackAlert).Methods(http.MethodPut)
	b.HandleFunc("/callbacks", removeCallbackAlert).Methods(http.MethodDelete)

	return r
}

const (
	ContentTypeKey       = "Content-Type"
	ContentTypeJsonValue = "application/json; charset=utf-8"
	ContentLengthKey     = "Content-Length"
)

func ping(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set(ContentTypeKey, ContentTypeJsonValue)
	rw.WriteHeader(http.StatusOK)
	str := `{"value" : "pong"}`
	io.WriteString(rw, str)
}

func addCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		loggingClient.Error("read request body error : " + err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		loggingClient.Error("failed to parse callback alert : " + err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {

	case models.SCHEDULE:
		schedule, err := schedulerClient.QuerySchedule(callbackAlert.Id)

		if err != nil {
			loggingClient.Error("query schedule error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = addSchedule(schedule)
		if err != nil {
			loggingClient.Error("add schedule error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	case models.SCHEDULEEVENT:
		scheduleEvent, err := schedulerClient.QueryScheduleEvent(callbackAlert.Id)
		if err != nil {
			loggingClient.Error("query schedule event error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := addScheduleEvent(scheduleEvent); err != nil {
			loggingClient.Error("add schedule event error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	default:
		loggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}

func updateCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		loggingClient.Error("reading the http request body error : " + err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		loggingClient.Error("failed to parse callback alert : " + err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {

	case models.SCHEDULE:
		schedule, err := schedulerClient.QuerySchedule(callbackAlert.Id)

		if err != nil {
			loggingClient.Error("query schedule error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = updateSchedule(schedule)
		if err != nil {
			loggingClient.Error("update schedule error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	case models.SCHEDULEEVENT:
		scheduleEvent, err := schedulerClient.QueryScheduleEvent(callbackAlert.Id)
		if err != nil {
			loggingClient.Error("query schedule event error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := updateScheduleEvent(scheduleEvent); err != nil {
			loggingClient.Error("query schedule event error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	default:
		loggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}

func removeCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	//here we need the action type, so request the callback alert json
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		loggingClient.Error("reading the http request body error : " + err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		loggingClient.Error("failed to parse callback alert : " + err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {
	case models.SCHEDULE:
		if err := removeSchedule(callbackAlert.Id); err != nil {
			loggingClient.Error("remove schedule error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		break

	case models.SCHEDULEEVENT:
		if err := removeScheduleEvent(callbackAlert.Id); err != nil {
			loggingClient.Error("remove schedule event error : " + err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		break

	default:
		loggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}
