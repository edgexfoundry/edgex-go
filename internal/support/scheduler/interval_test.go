/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package scheduler

import (
	errorsSched "github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	dbMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"testing"
)

var testInterval models.Interval
var testIntervalAction models.IntervalAction

var testRoutes *mux.Router

const (
	testIntervalName     string = "midnight"
	testInterNewName     string = "noon"
	testOrigin           int64  = 123456789
	testBsonString       string = "57e59a71e4b0ca8e6d6d4cc2"
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
	Configuration = &ConfigurationStruct{}
	testInterval.ID = testUUIDString
	testInterval.Origin = testOrigin
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
	dbClient = myMock

	intervals, err := getIntervals(limit)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(intervals) != limit {
		t.Errorf("expected %d interval", limit)
	}

	myMock.AssertExpectations(t)
}

func TestGetIntervals(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("Intervals").Return([]models.Interval{testInterval}, nil)
	dbClient = myMock

	intervals, err := getIntervals(0)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(intervals) == 0 {
		t.Errorf("no interval found")
	}

	if len(intervals) != 1 {
		t.Errorf("expected 1 interval")
	}
}

func TestIntervalBylName(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalByName",
		mock.MatchedBy(func(name string) bool { return name == testInterval.Name })).Return(testInterval, nil)
	dbClient = myMock

	interval, err := getIntervalByName(testInterval.Name)
	if err != nil {
		t.Errorf(err.Error())
	}

	if interval.Name != testInterval.Name {
		t.Errorf("expected interval name to be the same")
	}
}

func TestIntervalById(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalById",
		mock.MatchedBy(func(name string) bool { return name == testInterval.ID })).Return(testInterval, nil)
	dbClient = myMock

	interval, err := getIntervalById(testInterval.ID)
	if err != nil {
		t.Errorf(err.Error())
	}

	if interval.ID != testInterval.ID {
		t.Errorf("expected interval ID to be the same")
	}
}

func TestAddInterval(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}
	mySchedulerMock := &dbMock.SchedulerQueueClient{}

	// Validation call
	myMock.On("IntervalByName",
		mock.Anything).Return(models.Interval{}, nil)

	// Add Interval call
	myMock.On("AddInterval",
		mock.Anything).Return(testUUIDString, nil)

	// Scheduler call
	mySchedulerMock.On("AddIntervalToQueue",
		mock.Anything).Return(nil)

	mySchedulerMock.On("QueryIntervalByName",
		mock.Anything).Return(models.Interval{}, nil)

	nInterval := models.Interval{Name: testInterNewName, Origin: testOrigin}
	dbClient = myMock
	scClient = mySchedulerMock

	id, err := addNewInterval(nInterval)
	if err != nil {
		t.Errorf(err.Error())
	}

	if id != testUUIDString {
		t.Errorf("expected return interval ID to match inserted ID")
	}

	myMock.AssertExpectations(t)
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

	nInterval := models.Interval{Name: testInterval.Name, Origin: testOrigin}
	dbClient = myMock
	scClient = mySchedulerMock

	_, err := addNewInterval(nInterval)
	if err != nil {
		switch err.(type) {
		case *errorsSched.ErrIntervalNameInUse:
		// expected
		default:
			t.Errorf("Expected errors.ErrIntervalNameInUse")
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

	nInterval := models.Interval{Name: testInterval.Name, Start: "34343", Origin: testOrigin}
	dbClient = myMock
	scClient = mySchedulerMock

	_, err := addNewInterval(nInterval)
	if err != nil {
		switch err.(type) {
		case *errorsSched.ErrInvalidTimeFormat:
		// expected
		default:
			t.Errorf("Expected errors.ErrInvalidTimeFormat")
		}
	}
}

//TODO:  TestNoIDPassedByName
//TODO:  TestUpdatingExistingIntervalName
func TestUpdateInterval(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}
	mySchedulerMock := &dbMock.SchedulerQueueClient{}

	// Validation call
	myMock.On("IntervalById",
		mock.Anything).Return(models.Interval{Name: testIntervalName}, nil)

	// Update IntervalAction call
	myMock.On("UpdateInterval",
		mock.Anything).Return(nil)

	mySchedulerMock.On("UpdateIntervalInQueue",
		mock.Anything).Return(nil)

	nInterval := models.Interval{Name: testIntervalName, Origin: testOrigin}
	dbClient = myMock
	scClient = mySchedulerMock

	err := updateInterval(nInterval)

	if err != nil {
		t.Errorf(err.Error())
	}

	myMock.AssertExpectations(t)
}

//TODO: TestFailDeleteOnExistingIntervalActions
func TestDeleteIntervalById(t *testing.T) {
	reset()

	myMock := &dbMock.DBClient{}
	mySchedulerMock := &dbMock.SchedulerQueueClient{}

	// Validation call
	myMock.On("IntervalById",
		mock.MatchedBy(func(id string) bool { return id == testInterval.ID })).Return(testInterval, nil)

	// remove the IntervalAction from DB
	myMock.On("DeleteIntervalById",
		mock.Anything).Return(nil)

	// no associated interval actions
	myMock.On("IntervalActionsByIntervalName",
		mock.Anything).Return([]models.IntervalAction{}, nil)

	// Queue Validation
	mySchedulerMock.On("QueryIntervalByID",
		mock.Anything).Return(models.Interval{ID: testUUIDString}, nil)

	// remove the IntervalAction from memory
	mySchedulerMock.On("RemoveIntervalInQueue",
		mock.Anything).Return(nil)

	// assign clients to mocks
	dbClient = myMock
	scClient = mySchedulerMock

	err := deleteIntervalById(testUUIDString)
	if err != nil {
		t.Errorf(err.Error())
	}
	myMock.AssertExpectations(t)
	mySchedulerMock.AssertExpectations(t)
}

//TODO: TestFailDeleteOnExistingIntervalActions
func TestDeleteIntervalByName(t *testing.T) {
	reset()

	myMock := &dbMock.DBClient{}
	mySchedulerMock := &dbMock.SchedulerQueueClient{}

	// Validation call
	myMock.On("IntervalByName",
		mock.MatchedBy(func(name string) bool { return name == testInterval.Name })).Return(testInterval, nil)

	// remove the IntervalAction from DB
	myMock.On("DeleteIntervalById",
		mock.Anything).Return(nil)

	// no associated interval actions
	myMock.On("IntervalActionsByIntervalName",
		mock.Anything).Return([]models.IntervalAction{}, nil)

	// Queue Validation
	mySchedulerMock.On("QueryIntervalByName",
		mock.Anything).Return(models.Interval{Name: testIntervalName}, nil)

	// remove the IntervalAction from memory
	mySchedulerMock.On("RemoveIntervalInQueue",
		mock.Anything).Return(nil)

	// assign clients to mocks
	dbClient = myMock
	scClient = mySchedulerMock

	err := deleteIntervalByName(testIntervalName)
	if err != nil {
		t.Errorf(err.Error())
	}
	myMock.AssertExpectations(t)
	mySchedulerMock.AssertExpectations(t)
}
