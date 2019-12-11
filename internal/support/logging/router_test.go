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

	types "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/config"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/filter"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type dummyPersist struct {
	criteria filter.Criteria
	deleted  int
	added    int
}

const (
	numberOfLogs = 2
)

func (dp *dummyPersist) Add(_ models.LogEntry) error {
	dp.added += 1
	return nil
}

func (dp *dummyPersist) Remove(criteria filter.Criteria) (int, error) {
	dp.criteria = criteria
	dp.deleted = 42
	return dp.deleted, nil
}

func (dp *dummyPersist) Find(criteria filter.Criteria) ([]models.LogEntry, error) {
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

func (dp dummyPersist) Reset() {
}

func (dp *dummyPersist) CloseSession() {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.data))
			addLog(rr, req, &dummyPersist{})
			response := rr.Result()
			defer func() { _ = response.Body.Close() }()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
		})
	}
}

func TestGetLogs(t *testing.T) {
	const maxLimit = 100
	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{models.TraceLog, models.DebugLog, models.WarnLog, models.InfoLog, models.ErrorLog}
	var tests = []struct {
		name       string
		vars       map[string]string
		status     int
		criteria   filter.Criteria
		limitCheck int
	}{
		{"withoutParams",
			map[string]string{},
			http.StatusOK,
			filter.Criteria{},
			maxLimit},
		{"limit",
			map[string]string{"limit": "1000"},
			http.StatusOK,
			filter.Criteria{Limit: 1000},
			maxLimit},
		{"invalidlimit",
			map[string]string{"limit": "-1"},
			http.StatusBadRequest,
			filter.Criteria{Limit: 1000},
			maxLimit},
		{"wronglimit",
			map[string]string{"limit": "ten"},
			http.StatusBadRequest,
			filter.Criteria{Limit: 1000},
			maxLimit},
		{"start/end/limit",
			map[string]string{"start": "1", "end": "2", "limit": "3"},
			http.StatusOK,
			filter.Criteria{Start: 1, End: 2, Limit: 3},
			3},
		{"invalidstart/end/limit",
			map[string]string{"start": "-1", "end": "2", "limit": "3"},
			http.StatusBadRequest,
			filter.Criteria{},
			3},
		{"start/invalidend/limit",
			map[string]string{"start": "1", "end": "-2", "limit": "3"},
			http.StatusBadRequest,
			filter.Criteria{},
			3},
		{"wrongstart/end/limit",
			map[string]string{"start": "one", "end": "2", "limit": "3"},
			http.StatusBadRequest,
			filter.Criteria{},
			3},
		{"start/wrongend/limit",
			map[string]string{"start": "1", "end": "two", "limit": "3"},
			http.StatusBadRequest,
			filter.Criteria{},
			3},
		{"services/start/end/limit",
			map[string]string{"services": "service1,service2", "start": "1", "end": "2", "limit": "3"},
			http.StatusOK,
			filter.Criteria{OriginServices: services, Start: 1, End: 2, Limit: 3},
			3},
		{"keywords/start/end/limit",
			map[string]string{"keywords": "keyword1,keyword2", "start": "1", "end": "2", "limit": "3"},
			http.StatusOK,
			filter.Criteria{Keywords: keywords, Start: 1, End: 2, Limit: 3},
			3},
		{"levels/start/end/limit",
			map[string]string{"levels": "TRACE,DEBUG,WARN,INFO,ERROR", "start": "1", "end": "2", "limit": "3"},
			http.StatusOK,
			filter.Criteria{LogLevels: logLevels, Start: 1, End: 2, Limit: 3},
			3},
		{"wronglevels/start/end/limit",
			map[string]string{"levels": "INF,ERROR", "start": "1", "end": "2", "limit": "3"},
			http.StatusBadRequest,
			filter.Criteria{},
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
			filter.Criteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2, Limit: 3},
			3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configuration := &config.ConfigurationStruct{
				Service: types.ServiceInfo{
					MaxResultCount: maxLimit,
				},
			}
			persistence := &dummyPersist{}
			rr := httptest.NewRecorder()
			getLogs(rr, tt.vars, persistence, configuration)
			response := rr.Result()
			defer func() { _ = response.Body.Close() }()

			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
			// Only test that criteria is correctly parsed if request is valid
			if tt.status == http.StatusOK {
				//Apply rules for limit validation to original criteria
				tt.criteria.Limit = checkMaxLimitCount(tt.criteria.Limit, configuration)
				//Then compare against what was persisted during the test run
				if !reflect.DeepEqual(persistence.criteria, tt.criteria) {
					t.Errorf("Invalid criteria %v, should be %v", persistence.criteria, tt.criteria)
				}
			}
		})
	}
}

func TestRemoveLogs(t *testing.T) {
	const maxLimit = 100

	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{models.TraceLog, models.DebugLog, models.WarnLog, models.InfoLog, models.ErrorLog}
	var tests = []struct {
		name     string
		vars     map[string]string
		status   int
		criteria filter.Criteria
	}{
		{"start/end",
			map[string]string{"start": "1", "end": "2"},
			http.StatusOK,
			filter.Criteria{Start: 1, End: 2}},
		{"invalidstart/end",
			map[string]string{"start": "-1", "end": "2"},
			http.StatusBadRequest,
			filter.Criteria{}},
		{"start/invalidend",
			map[string]string{"start": "1", "end": "-2"},
			http.StatusBadRequest,
			filter.Criteria{}},
		{"wrongstart/end",
			map[string]string{"start": "one", "end": "2"},
			http.StatusBadRequest,
			filter.Criteria{}},
		{"start/wrongend",
			map[string]string{"start": "1", "end": "two"},
			http.StatusBadRequest,
			filter.Criteria{}},
		{"services/start/end",
			map[string]string{"services": "service1,service2", "start": "1", "end": "2"},
			http.StatusOK,
			filter.Criteria{OriginServices: services, Start: 1, End: 2}},
		{"keywords/start/end",
			map[string]string{"keywords": "keyword1,keyword2", "start": "1", "end": "2"},
			http.StatusOK,
			filter.Criteria{Keywords: keywords, Start: 1, End: 2}},
		{"levels/start/end",
			map[string]string{"levels": "TRACE,DEBUG,WARN,INFO,ERROR", "start": "1", "end": "2"},
			http.StatusOK,
			filter.Criteria{LogLevels: logLevels, Start: 1, End: 2}},
		{"wronglevels/start/end",
			map[string]string{"levels": "INF,ERROR", "start": "1", "end": "2"},
			http.StatusBadRequest,
			filter.Criteria{}},
		{"levels/services/start/end",
			map[string]string{
				"levels":   "TRACE,DEBUG,WARN,INFO,ERROR",
				"services": "service1,service2",
				"start":    "1",
				"end":      "2",
			},
			http.StatusOK,
			filter.Criteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configuration := &config.ConfigurationStruct{
				Service: types.ServiceInfo{
					MaxResultCount: maxLimit,
				},
			}
			persistence := &dummyPersist{}
			rr := httptest.NewRecorder()
			delLogs(rr, tt.vars, persistence, configuration)
			response := rr.Result()
			defer func() { _ = response.Body.Close() }()

			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
			// Only test that criteria is correctly parsed if request has status ok and limit > 0
			if tt.status == http.StatusOK && tt.criteria.Limit != 0 {
				if !reflect.DeepEqual(persistence.criteria, tt.criteria) {
					t.Errorf("Invalid criteria %v, should be %v", persistence.criteria, tt.criteria)
				}
				bodyBytes, _ := ioutil.ReadAll(response.Body)
				if string(bodyBytes) != strconv.Itoa(persistence.deleted) {
					t.Errorf("Invalid criteria %v, should be %v", persistence.criteria, tt.criteria)
				}
			}
		})
	}
}
