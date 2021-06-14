//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type SchedulerManager interface {
	StartTicker()
	StopTicker()

	AddInterval(interval models.Interval) errors.EdgeX
	UpdateInterval(interval models.Interval) errors.EdgeX
	DeleteIntervalByName(name string) errors.EdgeX

	AddIntervalAction(intervalAction models.IntervalAction) errors.EdgeX
	UpdateIntervalAction(intervalAction models.IntervalAction) errors.EdgeX
	DeleteIntervalActionByName(name string) errors.EdgeX
}
