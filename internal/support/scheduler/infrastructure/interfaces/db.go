//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type DBClient interface {
	CloseSession()

	AddInterval(interval model.Interval) (model.Interval, errors.EdgeX)
	IntervalById(id string) (model.Interval, errors.EdgeX)
	IntervalByName(name string) (model.Interval, errors.EdgeX)
	AllIntervals(offset int, limit int) ([]model.Interval, errors.EdgeX)
	DeleteIntervalByName(name string) errors.EdgeX
	UpdateInterval(interval model.Interval) errors.EdgeX
	IntervalTotalCount() (uint32, errors.EdgeX)

	AddIntervalAction(e model.IntervalAction) (model.IntervalAction, errors.EdgeX)
	AllIntervalActions(offset int, limit int) ([]model.IntervalAction, errors.EdgeX)
	IntervalActionByName(name string) (model.IntervalAction, errors.EdgeX)
	IntervalActionsByIntervalName(offset int, limit int, IntervalName string) ([]model.IntervalAction, errors.EdgeX)
	DeleteIntervalActionByName(name string) errors.EdgeX
	IntervalActionById(id string) (model.IntervalAction, errors.EdgeX)
	UpdateIntervalAction(action model.IntervalAction) errors.EdgeX
	IntervalActionTotalCount() (uint32, errors.EdgeX)
}
