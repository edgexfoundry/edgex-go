//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type dummyPersist struct {
	criteria db.LogMatcher
	deleted  int
	added    int
}

const (
	numberOfLogs = 2
)

func (dp *dummyPersist) AddLog(le models.LogEntry) error {
	dp.added += 1
	return nil
}

func (dp *dummyPersist) DeleteLog(criteria db.LogMatcher) (int, error) {
	dp.criteria = criteria
	dp.deleted = 42
	return dp.deleted, nil
}

func (dp *dummyPersist) FindLog(criteria db.LogMatcher, limit int) ([]models.LogEntry, error) {
	dp.criteria = criteria

	var retValue []models.LogEntry
	for i := 0; i < numberOfLogs; i++ {
		if i <= limit {
			retValue = append(retValue, models.LogEntry{})
		} else {
			break
		}
	}
	return retValue, nil
}

func (dp dummyPersist) ResetLogs() {
}

func (dp *dummyPersist) CloseSession() {
}

func (dp *dummyPersist) Connect() error {
	return nil
}

func TestPing(t *testing.T) {
	// create test server with handler
	ts := httptest.NewServer(HttpServer())
	defer ts.Close()

	response, err := http.Get(ts.URL + "/api/v1" + "/ping")
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
		{"ok", `{"logLevel":"INFO","labels":null,"originService":"tests","message":"test1"}`,
			http.StatusAccepted},
		{"invalidLevel", `{"logLevel":"NONE","labels":null,"originService":"tests","message":"test1"}`,
			http.StatusBadRequest},
	}
	// create test server with handler
	ts := httptest.NewServer(HttpServer())
	defer ts.Close()

	dbClient = &dummyPersist{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := http.Post(ts.URL+"/api/v1"+"/logs", "application/json", strings.NewReader(tt.data))
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
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{}}

	var labels = []string{"label1", "label2"}
	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{models.TRACE, models.DEBUG, models.WARN,
		models.INFO, models.ERROR}
	var tests = []struct {
		name     string
		url      string
		status   int
		maxLimit int
		criteria matchCriteria
	}{
		{"withoutParams",
			"",
			http.StatusOK,
			100,
			matchCriteria{}},
		{"limit",
			"1000",
			http.StatusOK,
			100,
			matchCriteria{Limit: 1000}},
		{"limitToLow",
			"1",
			http.StatusRequestEntityTooLarge,
			1,
			matchCriteria{Limit: 1}},
		{"invalidlimit",
			"-1",
			http.StatusBadRequest,
			100,
			matchCriteria{Limit: 1000}},
		{"wronglimit",
			"ten",
			http.StatusBadRequest,
			100,
			matchCriteria{Limit: 1000}},
		{"start/end/limit",
			"1/2/3",
			http.StatusOK,
			100,
			matchCriteria{Start: 1, End: 2, Limit: 3}},
		{"invalidstart/end/limit",
			"-1/2/3",
			http.StatusBadRequest,
			100,
			matchCriteria{}},
		{"start/invalidend/limit",
			"1/-2/3",
			http.StatusBadRequest,
			100,
			matchCriteria{}},
		{"wrongstart/end/limit",
			"one/2/3",
			http.StatusBadRequest,
			100,
			matchCriteria{}},
		{"start/wrongend/limit",
			"1/two/3",
			http.StatusBadRequest,
			100,
			matchCriteria{}},
		{"labels/start/end/limit",
			"labels/label1,label2/1/2/3",
			http.StatusOK,
			100,
			matchCriteria{Labels: labels, Start: 1, End: 2, Limit: 3}},
		{"labelsempty/start/end/limit",
			"labels//1/2/3",
			http.StatusOK,
			100,
			matchCriteria{Start: 1, End: 2, Limit: 3}},
		{"services/start/end/limit",
			"originServices/service1,service2/1/2/3",
			http.StatusOK,
			100,
			matchCriteria{OriginServices: services, Start: 1, End: 2, Limit: 3}},
		{"keywords/start/end/limit",
			"keywords/keyword1,keyword2/1/2/3",
			http.StatusOK,
			100,
			matchCriteria{Keywords: keywords, Start: 1, End: 2, Limit: 3}},
		{"levels/start/end/limit",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/1/2/3",
			http.StatusOK,
			100,
			matchCriteria{LogLevels: logLevels, Start: 1, End: 2, Limit: 3}},
		{"wronglevels/start/end/limit",
			"logLevels/INF,ERROR/1/2/3",
			http.StatusBadRequest,
			100,
			matchCriteria{}},
		{"levels/services/start/end/limit",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/1/2/3",
			http.StatusOK,
			100,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2, Limit: 3}},
		{"levels/services/labels/start/end/limit",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/labels/label1,label2/1/2/3",
			http.StatusOK,
			100,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Labels: labels, Start: 1, End: 2, Limit: 3}},
		{"levels/services/labels/keywords/start/end/limit",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/labels/label1,label2/keywords/keyword1,keyword2/1/2/3",
			http.StatusOK,
			100,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Labels: labels, Keywords: keywords, Start: 1, End: 2, Limit: 3}},
	}
	// create test server with handler
	ts := httptest.NewServer(HttpServer())
	defer ts.Close()

	dummy := &dummyPersist{}
	dbClient = dummy

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configuration.Service.ReadMaxLimit = tt.maxLimit
			response, err := http.Get(ts.URL + "/api/v1/logs/" + tt.url)
			if err != nil {
				t.Errorf("Error gettings logs %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
			// Only test that criteria is correctly parsed if request is valid
			if tt.status == http.StatusOK &&
				!reflect.DeepEqual(dummy.criteria, tt.criteria) {
				t.Errorf("Invalid criteria %v, should be %v", dummy.criteria, tt.criteria)
			}
		})
	}

}

func TestRemoveLogs(t *testing.T) {
	var labels = []string{"label1", "label2"}
	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{models.TRACE, models.DEBUG, models.WARN,
		models.INFO, models.ERROR}
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
		{"labels/start/end",
			"labels/label1,label2/1/2",
			http.StatusOK,
			matchCriteria{Labels: labels, Start: 1, End: 2}},
		{"labelsempty/start/end",
			"labels//1/2",
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
		{"levels/services/labels/start/end",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/labels/label1,label2/1/2",
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Labels: labels, Start: 1, End: 2}},
		{"levels/services/labels/keywords/start/end",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/labels/label1,label2/keywords/keyword1,keyword2/1/2",
			http.StatusOK,
			matchCriteria{LogLevels: logLevels, OriginServices: services, Labels: labels, Keywords: keywords, Start: 1, End: 2}},
	}
	// create test server with handler
	ts := httptest.NewServer(HttpServer())
	defer ts.Close()

	dummy := &dummyPersist{}
	dbClient = dummy

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1"+"/logs/"+tt.url, nil)
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
			// Only test that criteria is correctly parsed if request has status ok
			if tt.status == http.StatusOK {
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
