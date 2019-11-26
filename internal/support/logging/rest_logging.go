/*******************************************************************************
* Copyright 2019 Dell Inc.
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

package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	logging "github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"
)

func addLog(w http.ResponseWriter, r *http.Request, dbClient logging.Persistence) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, err.Error())
		return
	}

	l := models.LogEntry{}
	if err := json.Unmarshal(data, &l); err != nil {
		fmt.Println("Failed to parse LogEntry: ", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, err.Error())
		return
	}

	if !logger.IsValidLogLevel(l.Level) {
		s := fmt.Sprintf("Invalid level in LogEntry: %s", l.Level)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, s)
		return
	}

	l.Created = db.MakeTimestamp()

	w.WriteHeader(http.StatusAccepted)

	_ = dbClient.Add(l)
}

func getCriteria(w http.ResponseWriter, r *http.Request) *MatchCriteria {
	var criteria MatchCriteria
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
			_, _ = io.WriteString(w, s)
			return nil
		}
	}
	// In all cases, cap the # of entries returned at MaxResultCount
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
			_, _ = io.WriteString(w, s)
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
			_, _ = io.WriteString(w, s)
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
			_, _ = io.WriteString(w, s)
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
				_, _ = io.WriteString(w, s)
				return nil
			}
		}
	}

	return &criteria
}

func getLogs(w http.ResponseWriter, r *http.Request, dbClient logging.Persistence) {
	criteria := getCriteria(w, r)
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
	_, _ = io.WriteString(w, string(res))
}

func delLogs(w http.ResponseWriter, r *http.Request, dbClient logging.Persistence) {
	criteria := getCriteria(w, r)
	if criteria == nil {
		return
	}

	removed, err := dbClient.Remove(*criteria)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, strconv.Itoa(removed))
}
