//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/stretchr/testify/require"
	"gopkg.in/eapache/queue.v1"
)

const (
	testIntervalName       = "testIntervalName"
	testIntervalActionName = "testIntervalActionName"
)

func testManager() interfaces.SchedulerManager {
	lc := logger.NewMockClient()
	config := &config.ConfigurationStruct{
		Intervals:            nil,
		IntervalActions:      nil,
		ScheduleIntervalTime: 500,
	}
	return &manager{
		ticker:                time.NewTicker(time.Duration(config.ScheduleIntervalTime) * time.Millisecond),
		lc:                    lc,
		config:                config,
		executorQueue:         queue.New(),
		intervalToExecutorMap: make(map[string]*Executor),
		actionToIntervalMap:   make(map[string]string),
	}
}

func intervalData() models.Interval {
	return models.Interval{
		Name:     testIntervalName,
		Start:    "",
		End:      "",
		Interval: "10s",
	}
}

func intervalActionData() models.IntervalAction {
	return models.IntervalAction{
		Name:         testIntervalActionName,
		IntervalName: testIntervalName,
		Address:      models.RESTAddress{},
	}
}

func TestManager_AddInterval(t *testing.T) {
	interval := intervalData()
	intervalExists := testManager()
	err := intervalExists.AddInterval(interval)
	require.NoError(t, err)

	invalidStartTime := intervalData()
	invalidStartTime.Start = "20060102T"
	invalidEndTime := intervalData()
	invalidEndTime.End = "20060102T"
	invalidInterval := intervalData()
	invalidInterval.Interval = "10"

	tests := []struct {
		name              string
		manager           interfaces.SchedulerManager
		interval          models.Interval
		expectedErrorKind errors.ErrKind
	}{
		{"valid", testManager(), interval, ""},
		{"interval exists", intervalExists, interval, errors.KindStatusConflict},
		{"invalid start time format", testManager(), invalidStartTime, errors.KindContractInvalid},
		{"invalid end time format", testManager(), invalidEndTime, errors.KindContractInvalid},
		{"invalid interval format", testManager(), invalidInterval, errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.manager.AddInterval(testCase.interval)
			if testCase.expectedErrorKind != "" {
				require.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestManager_UpdateInterval(t *testing.T) {
	interval := intervalData()
	m := testManager()
	err := m.AddInterval(interval)
	require.NoError(t, err)

	invalidStartTime := intervalData()
	invalidStartTime.Start = "20060102T"
	invalidEndTime := intervalData()
	invalidEndTime.End = "20060102T"
	invalidInterval := intervalData()
	invalidInterval.Interval = "10"

	tests := []struct {
		name              string
		manager           interfaces.SchedulerManager
		interval          models.Interval
		expectedErrorKind errors.ErrKind
	}{
		{"valid", m, interval, ""},
		{"interval not found", testManager(), interval, errors.KindEntityDoesNotExist},
		{"invalid start time format", m, invalidStartTime, errors.KindContractInvalid},
		{"invalid end time format", m, invalidEndTime, errors.KindContractInvalid},
		{"invalid interval format", m, invalidInterval, errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.manager.UpdateInterval(testCase.interval)
			if testCase.expectedErrorKind != "" {
				require.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestManager_DeleteIntervalByName(t *testing.T) {
	interval := intervalData()
	m := testManager()
	err := m.AddInterval(interval)
	require.NoError(t, err)

	tests := []struct {
		name              string
		manager           interfaces.SchedulerManager
		intervalName      string
		expectedErrorKind errors.ErrKind
	}{
		{"valid", m, interval.Name, ""},
		{"interval not exists", testManager(), interval.Name, errors.KindEntityDoesNotExist},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.manager.DeleteIntervalByName(testCase.intervalName)
			if testCase.expectedErrorKind != "" {
				require.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestManager_AddIntervalAction(t *testing.T) {
	interval := intervalData()
	action := intervalActionData()

	m := testManager()
	err := m.AddInterval(interval)
	require.NoError(t, err)

	actionExists := testManager()
	err = actionExists.AddInterval(interval)
	require.NoError(t, err)
	err = actionExists.AddIntervalAction(action)
	require.NoError(t, err)

	tests := []struct {
		name              string
		manager           interfaces.SchedulerManager
		action            models.IntervalAction
		expectedErrorKind errors.ErrKind
	}{
		{"valid", m, action, ""},
		{"action already exists", actionExists, action, errors.KindStatusConflict},
		{"interval not exists", testManager(), action, errors.KindEntityDoesNotExist},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.manager.AddIntervalAction(testCase.action)
			if testCase.expectedErrorKind != "" {
				require.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestManager_UpdateIntervalAction(t *testing.T) {
	interval := intervalData()
	action := intervalActionData()

	m := testManager()
	err := m.AddInterval(interval)
	require.NoError(t, err)
	err = m.AddIntervalAction(action)
	require.NoError(t, err)

	actionNotFound := testManager()
	err = actionNotFound.AddInterval(interval)
	require.NoError(t, err)

	tests := []struct {
		name              string
		manager           interfaces.SchedulerManager
		action            models.IntervalAction
		expectedErrorKind errors.ErrKind
	}{
		{"valid", m, action, ""},
		{"action not found", actionNotFound, action, errors.KindEntityDoesNotExist},
		{"interval not found", testManager(), action, errors.KindEntityDoesNotExist},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.manager.UpdateIntervalAction(testCase.action)
			if testCase.expectedErrorKind != "" {
				require.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestManager_DeleteIntervalActionByName(t *testing.T) {
	interval := intervalData()
	action := intervalActionData()

	m := testManager()
	err := m.AddInterval(interval)
	require.NoError(t, err)
	err = m.AddIntervalAction(action)
	require.NoError(t, err)

	tests := []struct {
		name              string
		manager           interfaces.SchedulerManager
		actionName        string
		expectedErrorKind errors.ErrKind
	}{
		{"valid", m, action.Name, ""},
		{"action not found", m, "notFoundName", errors.KindEntityDoesNotExist},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.manager.DeleteIntervalActionByName(testCase.actionName)
			if testCase.expectedErrorKind != "" {
				require.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
				return
			}
			require.NoError(t, err)
		})
	}
}
