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
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	criteria2 "github.com/edgexfoundry/edgex-go/internal/support/logging/criteria"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

const (
	start          = "start"
	end            = "end"
	limit          = "limit"
	age            = "age"
	keywords       = "keywords"
	removeOld      = "removeOld"
	logLevels      = "logLevels"
	originServices = "originServices"
	levels         = "levels"
	services       = "services"
)

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(clients.ContentType, clients.ContentTypeText)
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, loggingClient logger.LoggingClient) {
	pkg.Encode(Configuration, w, loggingClient)
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

	dbClient.Add(l)
}

func getCriteria(w http.ResponseWriter, vars map[string]string) *criteria2.Criteria {
	var criteria criteria2.Criteria

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
	//In all cases, cap the # of entries returned at MaxResultCount
	criteria.Limit = checkMaxLimitCount(criteria.Limit)

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

func getLogs(w http.ResponseWriter, vars map[string]string) {
	criteria := getCriteria(w, vars)
	if criteria == nil {
		return
	}

	logs, err := dbClient.Find(*criteria)
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

func delLogs(w http.ResponseWriter, vars map[string]string) {
	criteria := getCriteria(w, vars)
	if criteria == nil {
		return
	}

	removed, err := dbClient.Remove(*criteria)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, strconv.Itoa(removed))
}

func metricsHandler(w http.ResponseWriter, loggingClient logger.LoggingClient) {
	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, loggingClient)

	return
}

func LoadRestRoutes(dic *di.Container) *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, func(w http.ResponseWriter, r *http.Request) {
		configHandler(w, bootstrapContainer.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, bootstrapContainer.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	// Logs
	r.HandleFunc(clients.ApiLoggingRoute, addLog).Methods(http.MethodPost)

	r.HandleFunc(clients.ApiLoggingRoute, func(w http.ResponseWriter, r *http.Request) {
		getLogs(w, mux.Vars(r))
	}).Methods(http.MethodGet)

	l := r.PathPrefix(clients.ApiLoggingRoute).Subrouter()
	l.HandleFunc("/{"+limit+"}", func(w http.ResponseWriter, r *http.Request) {
		getLogs(w, mux.Vars(r))
	}).Methods(http.MethodGet)
	l.HandleFunc("/{"+start+"}/{"+end+"}/{"+limit+"}", func(w http.ResponseWriter, r *http.Request) {
		getLogs(w, mux.Vars(r))
	}).Methods(http.MethodGet)
	l.HandleFunc("/"+originServices+"/{"+services+"}/{"+start+"}/{"+end+"}/{"+limit+"}", func(w http.ResponseWriter, r *http.Request) {
		getLogs(w, mux.Vars(r))
	}).Methods(http.MethodGet)
	l.HandleFunc("/"+keywords+"/{"+keywords+"}/{"+start+"}/{"+end+"}/{"+limit+"}", func(w http.ResponseWriter, r *http.Request) {
		getLogs(w, mux.Vars(r))
	}).Methods(http.MethodGet)
	l.HandleFunc("/"+logLevels+"/{"+levels+"}/{"+start+"}/{"+end+"}/{"+limit+"}", func(w http.ResponseWriter, r *http.Request) {
		getLogs(w, mux.Vars(r))
	}).Methods(http.MethodGet)
	l.HandleFunc("/"+logLevels+"/{"+levels+"}/"+originServices+"/{"+services+"}/{"+start+"}/{"+end+"}/{"+limit+"}", func(w http.ResponseWriter, r *http.Request) {
		getLogs(w, mux.Vars(r))
	}).Methods(http.MethodGet)

	l.HandleFunc("/{"+start+"}/{"+end+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(w, mux.Vars(r))
	}).Methods(http.MethodDelete)
	l.HandleFunc("/"+keywords+"/{"+keywords+"}/{"+start+"}/{"+end+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(w, mux.Vars(r))
	}).Methods(http.MethodDelete)
	l.HandleFunc("/"+originServices+"/{"+services+"}/{"+start+"}/{"+end+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(w, mux.Vars(r))
	}).Methods(http.MethodDelete)
	l.HandleFunc("/"+logLevels+"/{"+levels+"}/{"+start+"}/{"+end+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(w, mux.Vars(r))
	}).Methods(http.MethodDelete)
	l.HandleFunc("/"+logLevels+"/{"+levels+"}/"+originServices+"/{"+services+"}/{"+start+"}/{"+end+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(w, mux.Vars(r))
	}).Methods(http.MethodDelete)
	l.HandleFunc("/"+removeOld+"/"+age+"/{"+age+"}", func(w http.ResponseWriter, r *http.Request) {
		delLogs(w, mux.Vars(r))
	}).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}
