//
// Copyright (c) 2018 Tencent
//
// Copyright (c) 2018 Dell Inc.
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/go-zoo/bone"

	"io"
	"io/ioutil"
	"net/http"
)

func LoadRestRoutes() http.Handler {

	mux := bone.New()

	// config
	mux.Get(internal.ApiConfigRoute, http.HandlerFunc(replyConfig))

	// default api route
	mv1 := mux.Prefix("/api/v1")

	// ping and info
	mv1.Get("/info/:name", http.HandlerFunc(replyInfo))
	mv1.Get("/ping", http.HandlerFunc(replyPing))

	// callbacks
	mv1.Post("/callbacks", http.HandlerFunc(addCallbackAlert))
	mv1.Put("/callbacks", http.HandlerFunc(updateCallbackAlert))
	mv1.Delete("/callbacks", http.HandlerFunc(removeCallbackAlert))

	return mux
}

func replyPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(ContentTypeKey, ContentTypeJsonValue)
	w.WriteHeader(http.StatusOK)
	str := `{"value" : "pong"}`
	io.WriteString(w, str)
}

func replyConfig(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Add(ContentTypeKey, ContentTypeJsonValue)

	enc := json.NewEncoder(w)
	err := enc.Encode(Configuration)
	// Problems encoding
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error encoding the data: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func replyInfo(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Add(ContentTypeKey, ContentTypeJsonValue)

	vars := bone.GetValue(r, "name")
	schedule, err := queryScheduleByName(vars)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("read info request error %s", err.Error()))
		http.Error(w,"Schedule/Event not found",http.StatusNotFound)
		return
	}

	enc := json.NewEncoder(w)
	errEnc := enc.Encode(schedule)

	// Problems encoding
	if errEnc != nil {
		LoggingClient.Error(fmt.Sprintf("Error encoding the data: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("read request body error : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to parse callback alert : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {

	case models.SCHEDULE:

		schedule, err := querySchedule(callbackAlert.Id)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = addSchedule(schedule)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("add schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	case models.SCHEDULEEVENT:

		scheduleEvent, err := queryScheduleEvent(callbackAlert.Id)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule event error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := addScheduleEvent(scheduleEvent); err != nil {
			LoggingClient.Error(fmt.Sprintf("add schedule event error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	default:
		LoggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}

func updateCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("reading the http request body error : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to parse callback alert : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {

	case models.SCHEDULE:

		schedule, err := querySchedule(callbackAlert.Id)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = updateSchedule(schedule)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("update schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	case models.SCHEDULEEVENT:

		scheduleEvent, err := queryScheduleEvent(callbackAlert.Id)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule event error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := updateScheduleEvent(scheduleEvent); err != nil {
			LoggingClient.Error(fmt.Sprintf("query schedule event error :%s ", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusCreated)
		}

		break

	default:
		LoggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}

func removeCallbackAlert(rw http.ResponseWriter, r *http.Request) {
	//here we need the action type, so request the callback alert json
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("reading the http request body error : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	callbackAlert := models.CallbackAlert{}
	if err := json.Unmarshal(data, &callbackAlert); err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to parse callback alert : %s", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	switch callbackAlert.ActionType {
	case models.SCHEDULE:
		if err := removeSchedule(callbackAlert.Id); err != nil {
			LoggingClient.Error(fmt.Sprintf("remove schedule error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		break

	case models.SCHEDULEEVENT:
		if err := removeScheduleEvent(callbackAlert.Id); err != nil {
			LoggingClient.Error(fmt.Sprintf("remove schedule event error : %s", err.Error()))
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			rw.WriteHeader(http.StatusOK)
		}
		break

	default:
		LoggingClient.Error(fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType))
		http.Error(rw, fmt.Sprintf("unsupported action type : %s", callbackAlert.ActionType), http.StatusBadRequest)
		break
	}
}

