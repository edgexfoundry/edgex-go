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
 *******************************************************************************/

package scheduler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	// Interval
	r.HandleFunc(clients.ApiIntervalRoute, restGetIntervals).Methods(http.MethodGet)
	r.HandleFunc(clients.ApiIntervalRoute, restUpdateInterval).Methods(http.MethodPut)
	r.HandleFunc(clients.ApiIntervalRoute, restAddInterval).Methods(http.MethodPost)
	interval := r.PathPrefix(clients.ApiIntervalRoute).Subrouter()
	interval.HandleFunc("/{"+ID+"}", restGetIntervalByID).Methods(http.MethodGet)
	interval.HandleFunc("/{"+ID+"}", restDeleteIntervalByID).Methods(http.MethodDelete)
	interval.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetIntervalByName).Methods(http.MethodGet)
	interval.HandleFunc("/"+NAME+"/{"+NAME+"}", restDeleteIntervalByName).Methods(http.MethodDelete)
	// Scrub "Intervals and IntervalActions"
	interval.HandleFunc("/"+SCRUB+"/", restScrubAllIntervals).Methods(http.MethodDelete)

	// IntervalAction
	r.HandleFunc(clients.ApiIntervalActionRoute, restGetIntervalAction).Methods(http.MethodGet)
	r.HandleFunc(clients.ApiIntervalActionRoute, intervalActionHandler).Methods(http.MethodPut, http.MethodPost)
	intervalAction := r.PathPrefix(clients.ApiIntervalActionRoute).Subrouter()
	intervalAction.HandleFunc("/{"+ID+"}", intervalActionByIdHandler).Methods(http.MethodGet, http.MethodDelete)
	intervalAction.HandleFunc("/"+NAME+"/{"+NAME+"}", intervalActionByNameHandler).Methods(http.MethodGet, http.MethodDelete)
	intervalAction.HandleFunc("/"+TARGET+"/{"+TARGET+"}", intervalActionByTargetHandler).Methods(http.MethodGet)
	intervalAction.HandleFunc("/"+INTERVAL+"/{"+INTERVAL+"}", intervalActionByIntervalHandler).Methods(http.MethodGet)

	// Scrub "IntervalActions"
	intervalAction.HandleFunc("/"+SCRUB+"/", scrubIntervalActionsHandler).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	pkg.Encode(Configuration, w, LoggingClient)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, LoggingClient)

	return
}

// ************************ INTERVAL ACTION HANDLERS ****************************

/*
Handler for the IntervalAction API
Status code 400 - bad request, malformed or missing data
Status code 404 - interval not found
Status code 413 - number of interval actions exceeds limit
Status code 500 - unanticipated issues
api/v1/interval
*/
func intervalActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	case http.MethodGet:
		intervalActions, err := getIntervalActions(Configuration.Service.MaxResultCount)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pkg.Encode(intervalActions, w, LoggingClient)
		break
		// Post a new IntervalAction
	case http.MethodPost:
		var intervalAction models.IntervalAction
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&intervalAction)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("error decoding intervalAction" + err.Error())
			return
		}
		LoggingClient.Info("posting new intervalAction: " + intervalAction.String())

		newId, err := addNewIntervalAction(intervalAction)
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
			LoggingClient.Error(err.Error())
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
			LoggingClient.Error("Error decoding the intervalAction: " + err.Error())
			return
		}

		LoggingClient.Info("Updating IntervalAction: " + from.ID)
		err = updateIntervalAction(from)
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
			default: //return an error on everything else.
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			LoggingClient.Error(err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
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
func intervalActionByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	id, err := url.QueryUnescape(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error un-escaping the value interval id: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		intervalAction, err := getIntervalActionById(id)
		if err != nil {
			switch x := err.(type) {
			case errors.ErrIntervalActionNotFound:
				http.Error(w, x.Error(), http.StatusNotFound)
			default:
				http.Error(w, x.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}
		pkg.Encode(intervalAction, w, LoggingClient)
		// Post a new Interval Action
	case http.MethodDelete:
		if err = deleteIntervalActionById(id); err != nil {
			switch err.(type) {
			case errors.ErrIntervalActionNotFound:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
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
func intervalActionByNameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	name, err := url.QueryUnescape(vars["name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error un-escaping the value name: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		intervalAction, err := getIntervalActionByName(name)
		if err != nil {
			switch x := err.(type) {
			case errors.ErrIntervalActionNotFound:
				http.Error(w, x.Error(), http.StatusNotFound)
			default:
				http.Error(w, x.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}
		pkg.Encode(intervalAction, w, LoggingClient)
		// Post a new Interval Action
	case http.MethodDelete:
		if err = deleteIntervalActionByName(name); err != nil {
			switch err.(type) {
			case errors.ErrIntervalActionNotFound:
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
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
func intervalActionByTargetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	target, err := url.QueryUnescape(vars["target"])
	//Issues un-escaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error un-escaping the value descriptor name: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		intervalActions, err := getIntervalActionsByTarget(target)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pkg.Encode(intervalActions, w, LoggingClient)
		break
	}
}

/*
Handler for the IntervalAction By-Interval API
Status code 404 - interval action not found
Status code 500 - unanticipated issues
api/v1/interval
*/
func intervalActionByIntervalHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// URL parameters
	vars := mux.Vars(r)
	interval, err := url.QueryUnescape(vars["interval"])
	//Issues un-escaping
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error un-escaping the value interval name: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		intervalActions, err := getIntervalActionsByInterval(interval)
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pkg.Encode(intervalActions, w, LoggingClient)
		break
	}
}

// Scrub only the IntervalAction(s) leaving the Interval(s) behind
func scrubIntervalActionsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case http.MethodDelete:
		count, err := scrubAllInteralActions()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
	}
}
