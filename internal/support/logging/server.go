//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	encode(Configuration, w)
}

func addLog(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	l := models.LogEntry{}
	if err := json.Unmarshal(data, &l); err != nil {
		fmt.Println("Failed to parse LogEntry: ", err)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	if !logger.IsValidLogLevel(l.Level) {
		s := fmt.Sprintf("Invalid level in LogEntry: %s", l.Level)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, s)
		return
	}

	l.Created = db.MakeTimestamp()

	w.WriteHeader(http.StatusAccepted)

	dbClient.add(l)
}

func checkMaxLimit(limit int) int {
	if limit > Configuration.Service.ReadMaxLimit || limit == 0 {
		return Configuration.Service.ReadMaxLimit
	}
	return limit
}

func getCriteria(w http.ResponseWriter, r *http.Request) *matchCriteria {
	var criteria matchCriteria
	vars := mux.Vars(r)

	limit := vars["limit"]
	if len(limit) > 0 {
		var err error
		criteria.Limit, err = strconv.Atoi(limit)
		var s string
		if err != nil {
			s = fmt.Sprintf("Could not parse limit %s", limit)
		} else if criteria.Limit < 0 {
			s = fmt.Sprintf("Limit cannot be negative %d", criteria.Limit)
		}
		if len(s) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, s)
			return nil
		}
	}
	//In all cases, cap the # of entries returned at ReadMaxLimit
	criteria.Limit = checkMaxLimit(criteria.Limit)

	start := vars["start"]
	if len(start) > 0 {
		var err error
		criteria.Start, err = strconv.ParseInt(start, 10, 64)
		var s string
		if err != nil {
			s = fmt.Sprintf("Could not parse start %s", start)
		} else if criteria.Start < 0 {
			s = fmt.Sprintf("Start cannot be negative %d", criteria.Start)
		}
		if len(s) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, s)
			return nil
		}
	}

	end := vars["end"]
	if len(end) > 0 {
		var err error
		criteria.End, err = strconv.ParseInt(end, 10, 64)
		var s string
		if err != nil {
			s = fmt.Sprintf("Could not parse end %s", end)
		} else if criteria.End < 0 {
			s = fmt.Sprintf("End cannot be negative %d", criteria.End)
		}
		if len(s) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, s)
			return nil
		}
	}

	age := vars["age"]
	if len(age) > 0 {
		criteria.Start = 0
		now := db.MakeTimestamp()
		var err error
		criteria.End, err = strconv.ParseInt(age, 10, 64)
		var s string
		if err != nil {
			s = fmt.Sprintf("Could not parse age %s", age)
		} else if criteria.End < 0 {
			s = fmt.Sprintf("Age cannot be negative %d", criteria.End)
		} else if criteria.End > now {
			s = fmt.Sprintf("Age value too large %d", criteria.End)
		}
		if len(s) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, s)
			return nil
		}
		criteria.End = now - criteria.End
	}

	services := vars["services"]
	if len(services) > 0 {
		criteria.OriginServices = append(criteria.OriginServices,
			strings.Split(services, ",")...)
	}

	keywords := vars["keywords"]
	if len(keywords) > 0 {
		criteria.Keywords = append(criteria.Keywords,
			strings.Split(keywords, ",")...)
	}

	logLevels := vars["levels"]
	if len(logLevels) > 0 {
		criteria.LogLevels = append(criteria.LogLevels,
			strings.Split(logLevels, ",")...)
		for _, l := range criteria.LogLevels {
			if !logger.IsValidLogLevel(l) {
				s := fmt.Sprintf("Invalid log level '%s'", l)
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, s)
				return nil
			}
		}
	}
	return &criteria
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	criteria := getCriteria(w, r)
	if criteria == nil {
		return
	}

	logs, err := dbClient.find(*criteria)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(logs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(res))
}

func delLogs(w http.ResponseWriter, r *http.Request) {
	criteria := getCriteria(w, r)
	if criteria == nil {
		return
	}

	removed, err := dbClient.remove(*criteria)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, strconv.Itoa(removed))
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	var t internal.Telemetry

	// The micro-service is to be considered the System Of Record (SOR) in terms of accurate information.
	// Fetch metrics for the scheduler service.
	var rtm runtime.MemStats

	// Read full memory stats
	runtime.ReadMemStats(&rtm)

	// Miscellaneous memory stats
	t.Alloc = rtm.Alloc
	t.TotalAlloc = rtm.TotalAlloc
	t.Sys = rtm.Sys
	t.Mallocs = rtm.Mallocs
	t.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	t.LiveObjects = t.Mallocs - t.Frees

	encode(t, w)

	return
}

// Helper function for encoding things for returning from REST calls
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HTTPServer function
func HttpServer() http.Handler {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	// Logs
	r.HandleFunc(clients.ApiLoggingRoute, addLog).Methods(http.MethodPost)

	r.HandleFunc(clients.ApiLoggingRoute, getLogs).Methods(http.MethodGet)
	l := r.PathPrefix(clients.ApiLoggingRoute).Subrouter()
	l.HandleFunc("/{limit}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/{start}/{end}/{limit}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/originServices/{services}/{start}/{end}/{limit}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/keywords/{keywords}/{start}/{end}/{limit}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/logLevels/{levels}/{start}/{end}/{limit}", getLogs).Methods(http.MethodGet)
	l.HandleFunc("/logLevels/{levels}/originServices/{services}/{start}/{end}/{limit}", getLogs).Methods(http.MethodGet)

	l.HandleFunc("/{start}/{end}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/keywords/{keywords}/{start}/{end}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/originServices/{services}/{start}/{end}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/logLevels/{levels}/{start}/{end}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/logLevels/{levels}/originServices/{services}/{start}/{end}", delLogs).Methods(http.MethodDelete)
	l.HandleFunc("/removeold/age/{age}", delLogs).Methods(http.MethodDelete)

	return r
}
