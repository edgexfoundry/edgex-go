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
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/container"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"
)

const (
	numberOfLogs = 2
)

type dummyPersist struct {
	criteria MatchCriteria
	deleted  int
	added    int
}

func (dp *dummyPersist) Add(le models.LogEntry) error {
	dp.added += 1
	return nil
}

func (dp *dummyPersist) Remove(criteria interfaces.Criteria) (int, error) {
	matchCriteria, ok := criteria.(MatchCriteria)
	if !ok {
		return 0, errors.New("unknown criteria type")
	}

	dp.criteria = matchCriteria
	dp.deleted = 42
	return dp.deleted, nil
}

func (dp *dummyPersist) Find(criteria interfaces.Criteria) ([]models.LogEntry, error) {
	matchCriteria, ok := criteria.(MatchCriteria)
	if !ok {
		return nil, errors.New("unknown criteria type")
	}

	dp.criteria = matchCriteria

	var retValue []models.LogEntry
	max := numberOfLogs
	if dp.criteria.Limit < max {
		max = dp.criteria.Limit
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
	dummy := &dummyPersist{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.PersistenceName: func(get di.Get) interface{} {
			return dummy
		},
	})

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
	ts := httptest.NewServer(LoadRestRoutes(dic))
	defer ts.Close()

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
	dummy := &dummyPersist{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.PersistenceName: func(get di.Get) interface{} {
			return dummy
		},
	})
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{}}
	maxLimit := 100
	Configuration.Service.MaxResultCount = maxLimit

	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{models.TraceLog, models.DebugLog, models.WarnLog,
		models.InfoLog, models.ErrorLog}
	var tests = []struct {
		name       string
		url        string
		status     int
		criteria   MatchCriteria
		limitCheck int
	}{
		{"withoutParams",
			"",
			http.StatusOK,
			MatchCriteria{},
			maxLimit},
		{"limit",
			"/1000",
			http.StatusOK,
			MatchCriteria{Limit: 1000},
			maxLimit},
		{"invalidlimit",
			"/-1",
			http.StatusBadRequest,
			MatchCriteria{Limit: 1000},
			maxLimit},
		{"wronglimit",
			"/ten",
			http.StatusBadRequest,
			MatchCriteria{Limit: 1000},
			maxLimit},
		{"start/end/limit",
			"/1/2/3",
			http.StatusOK,
			MatchCriteria{Start: 1, End: 2, Limit: 3},
			3},
		{"invalidstart/end/limit",
			"/-1/2/3",
			http.StatusBadRequest,
			MatchCriteria{},
			3},
		{"start/invalidend/limit",
			"/1/-2/3",
			http.StatusBadRequest,
			MatchCriteria{},
			3},
		{"wrongstart/end/limit",
			"/one/2/3",
			http.StatusBadRequest,
			MatchCriteria{},
			3},
		{"start/wrongend/limit",
			"/1/two/3",
			http.StatusBadRequest,
			MatchCriteria{},
			3},
		{"start/end/limit",
			"/1/2/3",
			http.StatusOK,
			MatchCriteria{Start: 1, End: 2, Limit: 3},
			3},
		{"services/start/end/limit",
			"/originServices/service1,service2/1/2/3",
			http.StatusOK,
			MatchCriteria{OriginServices: services, Start: 1, End: 2, Limit: 3},
			3},
		{"keywords/start/end/limit",
			"/keywords/keyword1,keyword2/1/2/3",
			http.StatusOK,
			MatchCriteria{Keywords: keywords, Start: 1, End: 2, Limit: 3},
			3},
		{"levels/start/end/limit",
			"/logLevels/TRACE,DEBUG,WARN,INFO,ERROR/1/2/3",
			http.StatusOK,
			MatchCriteria{LogLevels: logLevels, Start: 1, End: 2, Limit: 3},
			3},
		{"wronglevels/start/end/limit",
			"/logLevels/INF,ERROR/1/2/3",
			http.StatusBadRequest,
			MatchCriteria{},
			3},
		{"levels/services/start/end/limit",
			"/logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/1/2/3",
			http.StatusOK,
			MatchCriteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2, Limit: 3},
			3},
	}
	// create test server with handler
	ts := httptest.NewServer(LoadRestRoutes(dic))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := http.Get(ts.URL + clients.ApiLoggingRoute + tt.url)
			if err != nil {
				t.Errorf("Error gettings logs %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != tt.status {
				t.Errorf("Returned status %d, should be %d", response.StatusCode, tt.status)
			}
			// Only test that criteria is correctly parsed if request is valid
			if tt.status == http.StatusOK {
				// Apply rules for limit validation to original criteria
				tt.criteria.Limit = checkMaxLimitCount(tt.criteria.Limit)
				// Then compare against what was persisted during the test run
				if !reflect.DeepEqual(dummy.criteria, tt.criteria) {
					t.Errorf("Invalid criteria %v, should be %v", dummy.criteria, tt.criteria)
				}
			}
		})
	}
	Configuration = nil
}

func TestRemoveLogs(t *testing.T) {
	dummy := &dummyPersist{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.PersistenceName: func(get di.Get) interface{} {
			return dummy
		},
	})
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{}}
	maxLimit := 100
	Configuration.Service.MaxResultCount = maxLimit

	var services = []string{"service1", "service2"}
	var keywords = []string{"keyword1", "keyword2"}
	var logLevels = []string{models.TraceLog, models.DebugLog, models.WarnLog,
		models.InfoLog, models.ErrorLog}
	var tests = []struct {
		name     string
		url      string
		status   int
		criteria MatchCriteria
	}{
		{"start/end",
			"1/2",
			http.StatusOK,
			MatchCriteria{Start: 1, End: 2}},
		{"invalidstart/end",
			"-1/2",
			http.StatusBadRequest,
			MatchCriteria{}},
		{"start/invalidend",
			"1/-2",
			http.StatusBadRequest,
			MatchCriteria{}},
		{"wrongstart/end",
			"one/2",
			http.StatusBadRequest,
			MatchCriteria{}},
		{"start/wrongend",
			"1/two",
			http.StatusBadRequest,
			MatchCriteria{}},
		{"start/end",
			"1/2",
			http.StatusOK,
			MatchCriteria{Start: 1, End: 2}},
		{"services/start/end",
			"originServices/service1,service2/1/2",
			http.StatusOK,
			MatchCriteria{OriginServices: services, Start: 1, End: 2}},
		{"keywords/start/end",
			"keywords/keyword1,keyword2/1/2",
			http.StatusOK,
			MatchCriteria{Keywords: keywords, Start: 1, End: 2}},
		{"levels/start/end",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/1/2",
			http.StatusOK,
			MatchCriteria{LogLevels: logLevels, Start: 1, End: 2}},
		{"wronglevels/start/end",
			"logLevels/INF,ERROR/1/2",
			http.StatusBadRequest,
			MatchCriteria{}},
		{"levels/services/start/end",
			"logLevels/TRACE,DEBUG,WARN,INFO,ERROR/originServices/service1,service2/1/2",
			http.StatusOK,
			MatchCriteria{LogLevels: logLevels, OriginServices: services, Start: 1, End: 2}},
	}
	// create test server with handler
	ts := httptest.NewServer(LoadRestRoutes(dic))
	defer ts.Close()

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
