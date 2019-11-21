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
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	dbMock "github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces/mocks"
)

func newGetIntervalActionsWithLimitMockDB(expectedLimit int) *dbMock.DBClient {
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalActionsWithLimit", mock.MatchedBy(func(limit int) bool {
		return limit == expectedLimit
	})).Return(func(limit int) []models.IntervalAction {
		intervalActions := make([]models.IntervalAction, 0)
		for i := 0; i < limit; i++ {
			intervalActions = append(intervalActions, testIntervalAction)
		}
		return intervalActions
	}, nil)

	return myMock
}

func TestGetIntervalActionsWithLimit(t *testing.T) {
	reset()

	limit := 1
	myMock := newGetIntervalActionsWithLimitMockDB(limit)

	intervalActions, err := getIntervalActions(limit, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(intervalActions) != limit {
		t.Fatalf("expected %d event", limit)
	}

	myMock.AssertExpectations(t)
}

func TestGetIntervalActions(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalActions").Return([]models.IntervalAction{testIntervalAction}, nil)

	intervalActions, err := getIntervalActions(0, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(intervalActions) == 0 {
		t.Fatalf("no actions found")
	}

	if len(intervalActions) != 1 {
		t.Fatalf("expected 1 event")
	}
}

func TestGetIntervalActionsByIntervalName(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalActionsByIntervalName",
		mock.MatchedBy(
			func(name string) bool {
				return name == testIntervalAction.Interval
			})).Return([]models.IntervalAction{testIntervalAction}, nil)

	intervalActions, err := getIntervalActionsByInterval(testIntervalActionInterval, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(intervalActions) == 0 {
		t.Fatalf("no interval action(s) found")
	}
	if len(intervalActions) != 1 {
		t.Fatalf("expected 1 event")
	}
}

func TestGetIntervalActionByName(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalActionByName",
		mock.MatchedBy(func(name string) bool { return name == testIntervalAction.Name })).Return(testIntervalAction, nil)

	intervalAction, err := getIntervalActionByName(testIntervalActionName, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(intervalAction.Name) == 0 {
		t.Fatalf("no interval action found")
	}
	if intervalAction.Name != testIntervalActionName {
		t.Fatalf("incorrect interval action name found")
	}
}

func TestGetIntervalActionById(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("IntervalActionById",
		mock.MatchedBy(func(id string) bool { return id == testIntervalAction.ID })).Return(testIntervalAction, nil)

	intervalAction, err := getIntervalActionById(testUUIDString, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(intervalAction.ID) == 0 {
		t.Fatalf("no interval action found")
	}
	if intervalAction.ID != testUUIDString {
		t.Fatalf("incorrect UUID found")
	}
}

func TestUpdateIntervalAction(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}
	mySchedulerMock := &dbMock.SchedulerQueueClient{}

	// Validation call
	myMock.On("IntervalActionById",
		mock.Anything).Return(models.IntervalAction{Name: testIntervalActionName}, nil)

	myMock.On("IntervalByName",
		mock.Anything).Return(models.Interval{}, nil)

	// Update IntervalAction call
	myMock.On("UpdateIntervalAction",
		mock.Anything).Return(nil)

	mySchedulerMock.On("QueryIntervalActionByName",
		mock.Anything).Return(models.IntervalAction{}, errors.New("mock db not found"))

	nIntervalAction := models.IntervalAction{Name: testIntervalActionName, Target: testIntervalActionTarget, Origin: testOrigin, Interval: testIntervalActionInterval}
	scClient = mySchedulerMock

	err := updateIntervalAction(nIntervalAction, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}

	myMock.AssertExpectations(t)
}

func TestDeleteIntervalActionById(t *testing.T) {
	reset()

	myMock := &dbMock.DBClient{}
	mySchedulerMock := &dbMock.SchedulerQueueClient{}

	// Validation call
	myMock.On("IntervalActionById",
		mock.MatchedBy(func(id string) bool { return id == testIntervalAction.ID })).Return(testIntervalAction, nil)

	// remove the IntervalAction from DB
	myMock.On("DeleteIntervalActionById",
		mock.Anything).Return(nil)

	// Queue Validation
	mySchedulerMock.On("QueryIntervalActionByID",
		mock.Anything).Return(models.IntervalAction{}, nil)

	// remove the IntervalAction from memory
	mySchedulerMock.On("RemoveIntervalActionQueue",
		mock.Anything).Return(nil)

	// assign clients to mocks
	scClient = mySchedulerMock

	err := deleteIntervalActionById(testUUIDString, myMock)
	if err != nil {
		t.Fatalf(err.Error())
	}
	myMock.AssertExpectations(t)
	mySchedulerMock.AssertExpectations(t)
}
