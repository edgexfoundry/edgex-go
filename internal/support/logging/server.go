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

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/go-zoo/bone"
)

func replyPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	str := `{"value" : "pong"}`
	io.WriteString(w, str)
}

func replyConfig(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(Configuration)
	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	if !models.IsValidLogLevel(l.Level) {
		s := fmt.Sprintf("Invalid level in LogEntry: %s", l.Level)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, s)
		return
	}

	l.Created = db.MakeTimestamp()

	w.WriteHeader(http.StatusAccepted)

	dbClient.AddLog(l)
}

func getCriteria(w http.ResponseWriter, r *http.Request) *matchCriteria {
	var criteria matchCriteria
	limit := bone.GetValue(r, "limit")
	if len(limit) > 0 {
		var err error
		criteria.Limit, err = strconv.Atoi(limit)
		var s string
		if err != nil {
			s = fmt.Sprintf("Could not parse limit %s", limit)
		} else if criteria.Limit < 0 {
			s = fmt.Sprintf("Limit is not positive %d", criteria.Limit)
		}
		if len(s) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, s)
			return nil
		}
	}

	start := bone.GetValue(r, "start")
	if len(start) > 0 {
		var err error
		criteria.Start, err = strconv.ParseInt(start, 10, 64)
		var s string
		if err != nil {
			s = fmt.Sprintf("Could not parse start %s", start)
		} else if criteria.Start < 0 {
			s = fmt.Sprintf("Start is not positive %d", criteria.Start)
		}
		if len(s) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, s)
			return nil
		}
	}

	end := bone.GetValue(r, "end")
	if len(end) > 0 {
		var err error
		criteria.End, err = strconv.ParseInt(end, 10, 64)
		var s string
		if err != nil {
			s = fmt.Sprintf("Could not parse end %s", end)
		} else if criteria.End < 0 {
			s = fmt.Sprintf("End is not positive %d", criteria.End)
		}
		if len(s) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, s)
			return nil
		}
	}

	labels := bone.GetValue(r, "labels")
	if len(labels) > 0 {
		criteria.Labels = append(criteria.Labels, strings.Split(labels, ",")...)
	}

	services := bone.GetValue(r, "services")
	if len(services) > 0 {
		criteria.OriginServices = append(criteria.OriginServices,
			strings.Split(services, ",")...)
	}

	keywords := bone.GetValue(r, "keywords")
	if len(keywords) > 0 {
		criteria.Keywords = append(criteria.Keywords,
			strings.Split(keywords, ",")...)
	}

	logLevels := bone.GetValue(r, "levels")
	if len(logLevels) > 0 {
		criteria.LogLevels = append(criteria.LogLevels,
			strings.Split(logLevels, ",")...)
		for _, l := range criteria.LogLevels {
			if !models.IsValidLogLevel(l) {
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

	logs, err := dbClient.FindLog(*criteria, criteria.Limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if criteria.Limit > 0 && len(logs) > Configuration.Service.ReadMaxLimit {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		s := fmt.Sprintf("more logs %d than configured limit %d",
			len(logs), Configuration.Service.ReadMaxLimit)
		io.WriteString(w, s)
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

	removed, err := dbClient.DeleteLog(*criteria)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, strconv.Itoa(removed))
}

// HTTPServer function
func HttpServer() http.Handler {
	mux := bone.New()
	mux.Get(internal.ApiConfigRoute, http.HandlerFunc(replyConfig))

	mv1 := mux.Prefix("/api/v1")

	mv1.Get("/ping", http.HandlerFunc(replyPing))

	mv1.Post("/logs", http.HandlerFunc(addLog))
	mv1.Get("/logs/:limit", http.HandlerFunc(getLogs))
	mv1.Get("/logs/:start/:end/:limit", http.HandlerFunc(getLogs))
	mv1.Get("/logs/labels/:labels/:start/:end/:limit", http.HandlerFunc(getLogs))
	mv1.Get("/logs/originServices/:services/:start/:end/:limit", http.HandlerFunc(getLogs))
	mv1.Get("/logs/keywords/:keywords/:start/:end/:limit", http.HandlerFunc(getLogs))
	mv1.Get("/logs/logLevels/:levels/:start/:end/:limit", http.HandlerFunc(getLogs))
	mv1.Get("/logs/logLevels/:levels/originServices/:services/:start/:end/:limit", http.HandlerFunc(getLogs))
	mv1.Get("/logs/logLevels/:levels/originServices/:services/labels/:labels/:start/:end/:limit", http.HandlerFunc(getLogs))
	mv1.Get("/logs/logLevels/:levels/originServices/:services/labels/:labels/keywords/:keywords/:start/:end/:limit", http.HandlerFunc(getLogs))

	mv1.Delete("/logs/:start/:end", http.HandlerFunc(delLogs))
	mv1.Delete("/logs/keywords/:keywords/:start/:end", http.HandlerFunc(delLogs))
	mv1.Delete("/logs/labels/:labels/:start/:end", http.HandlerFunc(delLogs))
	mv1.Delete("/logs/originServices/:services/:start/:end", http.HandlerFunc(delLogs))
	mv1.Delete("/logs/logLevels/:levels/:start/:end", http.HandlerFunc(delLogs))
	mv1.Delete("/logs/logLevels/:levels/originServices/:services/:start/:end", http.HandlerFunc(delLogs))
	mv1.Delete("/logs/logLevels/:levels/originServices/:services/labels/:labels/:start/:end", http.HandlerFunc(delLogs))
	mv1.Delete("/logs/logLevels/:levels/originServices/:services/labels/:labels/keywords/:keywords/:start/:end", http.HandlerFunc(delLogs))
	return mux
}
