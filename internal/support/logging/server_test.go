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
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
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

func TestPing(t *testing.T) {
	// create test server with handler
	ts := httptest.NewServer(HttpServer())
	defer ts.Close()

	response, err := http.Get(ts.URL + clients.ApiPingRoute)
	if err != nil {
		t.Errorf("Error getting ping: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Returned status %d, should be %d", response.StatusCode, http.StatusOK)
	}

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
	// create test server with handler
	ts := httptest.NewServer(HttpServer())
	defer ts.Close()

	dbClient = &dummyPersist{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := http.Post(ts.URL+clients.ApiLoggingRoute, "application/json", strings.NewReader(tt.data))
			if err != nil {
				t.Errorf("Error sending log %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
		})
	}
}

func TestGetLogs(t *testing.T) {
	maxLimit := 100
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{}}
	Configuration.Service.ReadMaxLimit = maxLimit

	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{logger.TraceLog, logger.DebugLog, logger.WarnLog,
		logger.InfoLog, logger.ErrorLog}
	var tests = []struct {
		name       string
		url        string
		status     int
		criteria   matchCriteria
		limitCheck int
	}{
		{"withoutParams",
			"",
			http.StatusOK,
			matchCriteria{},
			maxLimit},
		{"limit",
			"1000",
			http.StatusOK,
			matchCriteria{Limit: 1000},
			maxLimit},
		{"invalidlimit",
			"-1",
			http.StatusBadRequest,
			matchCriteria{Limit: 1000},
			maxLimit},
		{"wronglimit",
			"ten",
			http.StatusBadRequest,
			matchCriteria{Limit: 1000},
			maxLimit},
		{"start/end/limit",
			"1/2/3",
			http.StatusOK,
			matchCriteria{Start: 1, End: 2, Limit: 3},
			3},
		{"invalidstart/end/limit",
			"-1/2/3",
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"start/invalidend/limit",
			"1/-2/3",
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"wrongstart/end/limit",
			"one/2/3",
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"start/wrongend/limit",
			"1/two/3",
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"start/end/limit",
			"1/2/3",
			http.StatusOK,
			matchCriteria{Start: 1, End: 2, Limit: 3},
			3},
		{"services/start/end/limit",
			"originServices/service1,service2/1/2/3",
			http.StatusOK,
			matchCriteria{OriginServices: services, Start: 1, End: 2, Limit: 3},
			3},
		{"keywords/start/end/limit",
			"keywords/keyword1,keyword2/1/2/3",
			http.StatusOK,
			matchCriteria{Keywords: keywords, Start: 1, End: 2, Limit: 3},
			3},
		{"levels/start/end/limit",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/1/2/3",
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, Start: 1, End: 2, Limit: 3},
			3},
		{"wronglevels/start/end/limit",
			"logLevels/INF,ERROR/1/2/3",
			http.StatusBadRequest,
			matchCriteria{},
			3},
		{"levels/services/start/end/limit",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/1/2/3",
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2, Limit: 3},
			3},
	}
	// create test server with handler
	ts := httptest.NewServer(HttpServer())
	defer ts.Close()

	dummy := &dummyPersist{}
	dbClient = dummy

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := http.Get(ts.URL + clients.ApiLoggingRoute + "/" + tt.url)
			if err != nil {
				t.Errorf("Error gettings logs %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
			// Only test that criteria is correctly parsed if request is valid
			if tt.status == http.StatusOK {
				//Apply rules for limit validation to original criteria
				tt.criteria.Limit = checkMaxLimit(tt.criteria.Limit)
				//Then compare against what was persisted during the test run
				if !reflect.DeepEqual(dummy.criteria, tt.criteria) {
					t.Errorf("Invalid criteria %v, should be %v", dummy.criteria, tt.criteria)
				}
			}
		})
	}
	Configuration = nil
}

func TestRemoveLogs(t *testing.T) {
	maxLimit := 100
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{}}
	Configuration.Service.ReadMaxLimit = maxLimit

	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{logger.TraceLog, logger.DebugLog, logger.WarnLog,
		logger.InfoLog, logger.ErrorLog}
	var tests = []struct {
		name     string
		url      string
		status   int
		criteria matchCriteria
	}{
		{"start/end",
			"1/2",
			http.StatusOK,
			matchCriteria{Start: 1, End: 2}},
		{"invalidstart/end",
			"-1/2",
			http.StatusBadRequest,
			matchCriteria{}},
		{"start/invalidend",
			"1/-2",
			http.StatusBadRequest,
			matchCriteria{}},
		{"wrongstart/end",
			"one/2",
			http.StatusBadRequest,
			matchCriteria{}},
		{"start/wrongend",
			"1/two",
			http.StatusBadRequest,
			matchCriteria{}},
		{"start/end",
			"1/2",
			http.StatusOK,
			matchCriteria{Start: 1, End: 2}},
		{"services/start/end",
			"originServices/service1,service2/1/2",
			http.StatusOK,
			matchCriteria{OriginServices: services, Start: 1, End: 2}},
		{"keywords/start/end",
			"keywords/keyword1,keyword2/1/2",
			http.StatusOK,
			matchCriteria{Keywords: keywords, Start: 1, End: 2}},
		{"levels/start/end",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/1/2",
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, Start: 1, End: 2}},
		{"wronglevels/start/end",
			"logLevels/INF,ERROR/1/2",
			http.StatusBadRequest,
			matchCriteria{}},
		{"levels/services/start/end",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/1/2",
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2}},
	}
	// create test server with handler
	ts := httptest.NewServer(HttpServer())
	defer ts.Close()

	dummy := &dummyPersist{}
	dbClient = dummy

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, ts.URL+clients.ApiLoggingRoute+"/"+tt.url, nil)
			if err != nil {
				t.Errorf("Error creating request logs %v", err)
			}

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("Error requesting DELETE %v", err)
			}
			defer response.Body.Close()
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
	Configuration = nil
}
