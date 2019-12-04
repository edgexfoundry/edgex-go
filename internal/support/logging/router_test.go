//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type dummyPersist struct {
	criteria matchCriteria
	deleted  int
	added    int
}

const (
	numberOfLogs = 2
)

func (dp *dummyPersist) add(le models.LogEntry) error {
	dp.added += 1
	return nil
}

func (dp *dummyPersist) remove(criteria matchCriteria) (int, error) {
	dp.criteria = criteria
	dp.deleted = 42
	return dp.deleted, nil
}

func (dp *dummyPersist) find(criteria matchCriteria) ([]models.LogEntry, error) {
	dp.criteria = criteria

	var retValue []models.LogEntry
	max := numberOfLogs
	if criteria.Limit < max {
		max = criteria.Limit
	}

	for i := 0; i < max; i++ {
		retValue = append(retValue, models.LogEntry{})
	}
	return retValue, nil
}

func (dp dummyPersist) reset() {
}

func (dp *dummyPersist) closeSession() {
}

func TestAddLog(t *testing.T) {
	var tests = []struct {
		name   string
		data   string
		status int
	}{
		{"emptyPost", "", http.StatusBadRequest},
		{"invalidJSON", "aa", http.StatusBadRequest},
		{"ok", `{"logLevel":"INFO","originService":"tests","message":"test1"}`,
			http.StatusAccepted},
		{"invalidLevel", `{"logLevel":"NONE","originService":"tests","message":"test1"}`,
			http.StatusBadRequest},
	}

	dbClient = &dummyPersist{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.data))
			addLog(rr, req)
			response := rr.Result()
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
		})
	}
}

func TestGetLogs(t *testing.T) {
	const maxLimit = 100
	defer func() { Configuration = nil }()
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{}}
	Configuration.Service.MaxResultCount = maxLimit

	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{models.TraceLog, models.DebugLog, models.WarnLog,
		models.InfoLog, models.ErrorLog}
	var tests = []struct {
		name       string
		vars       map[string]string
		status     int
		criteria   matchCriteria
		limitCheck int
	}{
		{"withoutParams",
			map[string]string{},
			http.StatusOK,
			matchCriteria{},
			maxLimit},
		{"limit",
			map[string]string{"limit": "1000"},
			http.StatusOK,
			matchCriteria{Limit: 1000},
			maxLimit},
		{"invalidlimit",
			map[string]string{"limit": "-1"},
			http.StatusBadRequest,
			matchCriteria{Limit: 1000},
			maxLimit},
		{"wronglimit",
			map[string]string{"limit": "ten"},
			http.StatusBadRequest,
			matchCriteria{Limit: 1000},
			maxLimit},
		{"start/end/limit",
			map[string]string{"start": "1", "end": "2", "limit": "3"},
			http.StatusOK,
			matchCriteria{Start: 1, End: 2, Limit: 3},
			3},
		{"invalidstart/end/limit",
			map[string]string{"start": "-1", "end": "2", "limit": "3"},
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"start/invalidend/limit",
			map[string]string{"start": "1", "end": "-2", "limit": "3"},
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"wrongstart/end/limit",
			map[string]string{"start": "one", "end": "2", "limit": "3"},
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"start/wrongend/limit",
			map[string]string{"start": "1", "end": "two", "limit": "3"},
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"services/start/end/limit",
			map[string]string{"services": "service1,service2", "start": "1", "end": "2", "limit": "3"},
			http.StatusOK,
			matchCriteria{OriginServices: services, Start: 1, End: 2, Limit: 3},
			3},
		{"keywords/start/end/limit",
			map[string]string{"keywords": "keyword1,keyword2", "start": "1", "end": "2", "limit": "3"},
			http.StatusOK,
			matchCriteria{Keywords: keywords, Start: 1, End: 2, Limit: 3},
			3},
		{"levels/start/end/limit",
			map[string]string{"levels": "TRACE,DEBUG,WARN,INFO,ERROR", "start": "1", "end": "2", "limit": "3"},
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, Start: 1, End: 2, Limit: 3},
			3},
		{"wronglevels/start/end/limit",
			map[string]string{"levels": "INF,ERROR", "start": "1", "end": "2", "limit": "3"},
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"levels/services/start/end/limit",
			map[string]string{
				"levels":   "TRACE,DEBUG,WARN,INFO,ERROR",
				"services": "service1,service2",
				"start":    "1",
				"end":      "2",
				"limit":    "3",
			},
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2, Limit: 3},
			3},
	}

	dummy := &dummyPersist{}
	dbClient = dummy

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			getLogs(rr, tt.vars)
			response := rr.Result()

			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
			// Only test that criteria is correctly parsed if request is valid
			if tt.status == http.StatusOK {
				//Apply rules for limit validation to original criteria
				tt.criteria.Limit = checkMaxLimitCount(tt.criteria.Limit)
				//Then compare against what was persisted during the test run
				if !reflect.DeepEqual(dummy.criteria, tt.criteria) {
					t.Errorf("Invalid criteria %v, should be %v", dummy.criteria, tt.criteria)
				}
			}
		})
	}
}

func TestRemoveLogs(t *testing.T) {
	const maxLimit = 100
	defer func() { Configuration = nil }()
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{}}
	Configuration.Service.MaxResultCount = maxLimit

	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{models.TraceLog, models.DebugLog, models.WarnLog,
		models.InfoLog, models.ErrorLog}
	var tests = []struct {
		name     string
		vars     map[string]string
		status   int
		criteria matchCriteria
	}{
		{"start/end",
			map[string]string{"start": "1", "end": "2"},
			http.StatusOK,
			matchCriteria{Start: 1, End: 2}},
		{"invalidstart/end",
			map[string]string{"start": "-1", "end": "2"},
			http.StatusBadRequest,
			matchCriteria{}},
		{"start/invalidend",
			map[string]string{"start": "1", "end": "-2"},
			http.StatusBadRequest,
			matchCriteria{}},
		{"wrongstart/end",
			map[string]string{"start": "one", "end": "2"},
			http.StatusBadRequest,
			matchCriteria{}},
		{"start/wrongend",
			map[string]string{"start": "1", "end": "two"},
			http.StatusBadRequest,
			matchCriteria{}},
		{"services/start/end",
			map[string]string{"services": "service1,service2", "start": "1", "end": "2"},
			http.StatusOK,
			matchCriteria{OriginServices: services, Start: 1, End: 2}},
		{"keywords/start/end",
			map[string]string{"keywords": "keyword1,keyword2", "start": "1", "end": "2"},
			http.StatusOK,
			matchCriteria{Keywords: keywords, Start: 1, End: 2}},
		{"levels/start/end",
			map[string]string{"levels": "TRACE,DEBUG,WARN,INFO,ERROR", "start": "1", "end": "2"},
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, Start: 1, End: 2}},
		{"wronglevels/start/end",
			map[string]string{"levels": "INF,ERROR", "start": "1", "end": "2"},
			http.StatusBadRequest,
			matchCriteria{}},
		{"levels/services/start/end",
			map[string]string{
				"levels":   "TRACE,DEBUG,WARN,INFO,ERROR",
				"services": "service1,service2",
				"start":    "1",
				"end":      "2",
			},
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2}},
	}

	dummy := &dummyPersist{}
	dbClient = dummy

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			delLogs(rr, tt.vars)
			response := rr.Result()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
			// Only test that criteria is correctly parsed if request has status ok and limit > 0
			if tt.status == http.StatusOK && tt.criteria.Limit != 0 {
				if !reflect.DeepEqual(dummy.criteria, tt.criteria) {
					t.Errorf("Invalid criteria %v, should be %v", dummy.criteria, tt.criteria)
				}
				bodyBytes, _ := ioutil.ReadAll(response.Body)
				if string(bodyBytes) != strconv.Itoa(dummy.deleted) {
					t.Errorf("Invalid criteria %v, should be %v", dummy.criteria, tt.criteria)
				}
			}
		})
	}
}
