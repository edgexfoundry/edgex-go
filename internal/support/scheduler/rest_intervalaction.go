/*******************************************************************************
 * Copyright 2019 VMware Inc.
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
 *******************************************************************************/

package scheduler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/intervalaction"
)

func restGetIntervalAction(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}
	op := intervalaction.NewAllExecutor(dbClient, Configuration.Service)
	intervalActions, err := op.Execute()

	if err != nil {
		loggingClient.Error(err.Error())
		switch err.(type) {
		case errors.ErrLimitExceeded:
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	pkg.Encode(intervalActions, w, loggingClient)
}

func restAddIntervalAction(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	scClient interfaces.SchedulerQueueClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}
	var intervalAction contract.IntervalAction
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&intervalAction)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	loggingClient.Info("posting new intervalAction: " + intervalAction.String())

	op := intervalaction.NewAddExecutor(dbClient, scClient, intervalAction)
	newId, err := op.Execute()
	if err != nil {
		switch t := err.(type) {
		case errors.ErrIntervalActionNameInUse:
			http.Error(w, t.Error(), http.StatusBadRequest)
		case errors.ErrIntervalNotFound:
			http.Error(w, t.Error(), http.StatusBadRequest)
		default:
			http.Error(w, t.Error(), http.StatusInternalServerError)
		}
		loggingClient.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(newId))
}

/*
Handler for the IntervalAction API
Status code 400 - bad request, malformed or missing data
Status code 404 - interval not found
Status code 413 - number of interval actions exceeds limit
Status code 500 - unanticipated issues
api/v1/interval
*/
func intervalActionHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	scClient interfaces.SchedulerQueueClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	case http.MethodGet:
		intervalActions, err := getIntervalActions(Configuration.Service.MaxResultCount, dbClient)
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pkg.Encode(intervalActions, w, loggingClient)
		break
		// Post a new IntervalAction
	case http.MethodPost:
		var intervalAction models.IntervalAction
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&intervalAction)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			loggingClient.Error("error decoding intervalAction" + err.Error())
			return
		}
		loggingClient.Info("posting new intervalAction: " + intervalAction.String())

		newId, err := addNewIntervalAction(intervalAction, dbClient, scClient)
		if err != nil {
			switch t := err.(type) {
			case errors.ErrIntervalActionNameInUse:
				http.Error(w, t.Error(), http.StatusBadRequest)
			case errors.ErrInvalidTimeFormat:
				http.Error(w, t.Error(), http.StatusBadRequest)
			case errors.ErrInvalidFrequencyFormat:
				http.Error(w, t.Error(), http.StatusBadRequest)
			default:
				http.Error(w, t.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(newId))
		break
	case http.MethodPut:
		var from models.IntervalAction
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&from)

		// Problem decoding
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			loggingClient.Error("Error decoding the intervalAction: " + err.Error())
			return
		}

		loggingClient.Info("Updating IntervalAction: " + from.ID)
		err = updateIntervalAction(from, dbClient, scClient)
		if err != nil {
			switch t := err.(type) {
			case errors.ErrIntervalNotFound:
				http.Error(w, t.Error(), http.StatusNotFound)
			case errors.ErrInvalidCronFormat:
				http.Error(w, t.Error(), http.StatusBadRequest)
			case errors.ErrInvalidFrequencyFormat:
				http.Error(w, t.Error(), http.StatusBadRequest)
			case errors.ErrInvalidTimeFormat:
				http.Error(w, t.Error(), http.StatusBadRequest)
			case errors.ErrIntervalStillUsedByIntervalActions:
				http.Error(w, t.Error(), http.StatusBadRequest)
			case errors.ErrIntervalNameInUse:
				http.Error(w, t.Error(), http.StatusBadRequest)
			default:
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

/*
Handler for the IntervalAction By-ID API
Status code 404 - interval not found
Status code 500 - unanticipated issues
api/v1/interval
*/
func intervalActionByIdHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	scClient interfaces.SchedulerQueueClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	id, err := url.QueryUnescape(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		loggingClient.Error("Error un-escaping the value interval id: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		intervalAction, err := getIntervalActionById(id, dbClient)
		if err != nil {
			switch x := err.(type) {
			case errors.ErrIntervalActionNotFound:
				http.Error(w, x.Error(), http.StatusNotFound)
			default:
				http.Error(w, x.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}
		pkg.Encode(intervalAction, w, loggingClient)
		// Post a new Interval Action
	case http.MethodDelete:
		if err = deleteIntervalActionById(id, dbClient, scClient); err != nil {
			switch err.(type) {
			case errors.ErrIntervalActionNotFound:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

/*
Handler for the IntervalAction By-Name API
Status code 404 - interval action not found
Status code 500 - unanticipated issues
api/v1/interval
*/
func intervalActionByNameHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	scClient interfaces.SchedulerQueueClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	name, err := url.QueryUnescape(vars["name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		loggingClient.Error("Error un-escaping the value name: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		intervalAction, err := getIntervalActionByName(name, dbClient)
		if err != nil {
			switch x := err.(type) {
			case errors.ErrIntervalActionNotFound:
				http.Error(w, x.Error(), http.StatusNotFound)
			default:
				http.Error(w, x.Error(), http.StatusInternalServerError)
			}
			loggingClient.Error(err.Error())
			return
		}
		pkg.Encode(intervalAction, w, loggingClient)
		// Post a new Interval Action
	case http.MethodDelete:
		if err = deleteIntervalActionByName(name, dbClient, scClient); err != nil {
			switch err.(type) {
			case errors.ErrIntervalActionNotFound:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

/*
Handler for the IntervalAction By-Target API
Status code 404 - interval action not found
Status code 500 - unanticipated issues
api/v1/interval
*/
func intervalActionByTargetHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	target, err := url.QueryUnescape(vars["target"])

	// Issues un-escaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		loggingClient.Error("Error un-escaping the value descriptor name: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		intervalActions, err := getIntervalActionsByTarget(target, dbClient)
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pkg.Encode(intervalActions, w, loggingClient)
		break
	}
}

/*
Handler for the IntervalAction By-Interval API
Status code 404 - interval action not found
Status code 500 - unanticipated issues
api/v1/interval
*/
func intervalActionByIntervalHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	interval, err := url.QueryUnescape(vars["interval"])

	// Issues un-escaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		loggingClient.Error("Error un-escaping the value interval name: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		intervalActions, err := getIntervalActionsByInterval(interval, dbClient)
		if err != nil {
			loggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pkg.Encode(intervalActions, w, loggingClient)
		break
	}
}

// Scrub only the IntervalAction(s) leaving the Interval(s) behind
func scrubIntervalActionsHandler(
	w http.ResponseWriter,
	r *http.Request,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient) {

	defer r.Body.Close()

	switch r.Method {
	case http.MethodDelete:
		count, err := scrubAllInteralActions(loggingClient, dbClient)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	}
}
