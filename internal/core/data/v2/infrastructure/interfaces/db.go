//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type DBClient interface {
	CloseSession()

	AddEvent(e model.Event) (model.Event, errors.EdgeX)
	EventById(id string) (model.Event, errors.EdgeX)
	DeleteEventById(id string) errors.EdgeX
	EventTotalCount() (uint32, errors.EdgeX)
	EventCountByDevice(deviceName string) (uint32, errors.EdgeX)
	UpdateEventPushedById(id string) errors.EdgeX
	AllEvents(offset int, limit int) ([]model.Event, errors.EdgeX)
	EventsByDeviceName(offset int, limit int, name string) ([]model.Event, errors.EdgeX)
	DeletePushedEvents() errors.EdgeX
}
