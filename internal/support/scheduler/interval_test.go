// *******************************************************************************
// * Copyright 2018 Dell Inc.
// *
// * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// * in compliance with the License. You may obtain a copy of the License at
// *
// * http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing, software distributed under the License
// * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// * or implied. See the License for the specific language governing permissions and limitations under
// * the License.
// *******************************************************************************/

package scheduler

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stretchr/testify/mock"

	errorsSched "github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces/mocks"
)

var testInterval models.Interval
var testIntervalAction models.IntervalAction
var testTimestamps models.Timestamps

const (
	testIntervalName     string = "midnight"
	testInterNewName     string = "noon"
	testOrigin           int64  = 123456789
	testUUIDString       string = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"
	testIntervalActionId string = "ca93c8fa-9919-4ec5-85d3-f81b2b6a7bc1"

	testIntervalActionName     string = "scrub-aged-events"
	testIntervalActionNewName  string = "scub-bub"
	testIntervalActionTarget   string = "core-data"
	testIntervalActionInterval string = "midnight"
)

// Supporting methods
// Reset() re-initializes dependencies for each test
func reset() {
	testTimestamps = models.Timestamps{Origin: testOrigin}
	testInterval.ID = testUUIDString
	testInterval.Timestamps = testTimestamps
	testInterval.Name = testIntervalName

	testIntervalAction.ID = testUUIDString
	testIntervalAction.Name = testIntervalActionName
	testIntervalAction.Target = testIntervalActionTarget
	testIntervalAction.Interval = testIntervalActionInterval
}

func newGetIntervalsWithLimitMockDB(expectedLimit int) *dbMock.DBClient {
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalsWithLimit", mock.MatchedBy(func(limit int) bool {
		return limit == expectedLimit
	})).Return(func(limit int) []models.Interval {
		intervals := make([]models.Interval, 0)
		for i := 0; i < limit; i++ {
			intervals = append(intervals, testInterval)
		}
		return intervals
	}, nil)

	return myMock
}

func TestGetIntervalsWithLimit(t *testing.T) {
	reset()

	limit := 1
	myMock := newGetIntervalsWithLimitMockDB(limit)

	intervals, err := getIntervals(limit, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(intervals) != limit {
		t.Fatalf("expected %d interval", limit)
	}

	myMock.AssertExpectations(t)
}

func TestIntervalBylName(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalByName",
		mock.MatchedBy(func(name string) bool { return name == testInterval.Name })).Return(testInterval, nil)

	interval, err := getIntervalByName(testInterval.Name, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if interval.Name != testInterval.Name {
		t.Fatalf("expected interval name to be the same")
	}
}

func TestAddIntervalFailOnExistingName(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}
	mySchedulerMock := &dbMock.SchedulerQueueClient{}

	// Validation Call
	myMock.On("IntervalByName",
		mock.Anything).Return(testInterval, nil)

	// Add Interval Call
	myMock.On("AddInterval",
		mock.Anything).Return(testUUIDString, nil)

	// Scheduler call
	mySchedulerMock.On("AddIntervalToQueue",
		mock.Anything).Return(nil)

	// Scheduler call
	mySchedulerMock.On("QueryIntervalByName",
		mock.Anything).Return(models.Interval{}, nil)

	nInterval := models.Interval{Name: testInterval.Name, Timestamps: testTimestamps}

	_, err := addNewInterval(nInterval, myMock, mySchedulerMock)
	if err != nil {
		switch err.(type) {
		case errorsSched.ErrIntervalNameInUse:
		// expected
		default:
			t.Fatalf("Expected errors.ErrIntervalNameInUse")
		}
	}
}

func TestAddIntervalFailOnInvalidTimeFormat(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}
	mySchedulerMock := &dbMock.SchedulerQueueClient{}

	// Validation Call
	myMock.On("IntervalByName",
		mock.Anything).Return(models.Interval{}, nil)

	// Add Interval Call
	myMock.On("AddInterval",
		mock.Anything).Return(testUUIDString, nil)

	// Scheduler call
	mySchedulerMock.On("AddIntervalToQueue",
		mock.Anything).Return(nil)

	mySchedulerMock.On("QueryIntervalByName",
		mock.Anything).Return(models.Interval{}, nil)

	nInterval := models.Interval{Name: testInterval.Name, Start: "34343", Timestamps: testTimestamps}

	_, err := addNewInterval(nInterval, myMock, mySchedulerMock)
	if err != nil {
		switch err.(type) {
		case errorsSched.ErrInvalidTimeFormat:
		// expected
		default:
			t.Fatalf("Expected errors.ErrInvalidTimeFormat")
		}
	}
}
