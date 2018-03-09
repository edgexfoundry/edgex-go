//
// Copyright (c) 2017
// Cavium
//
// SPDX-License-Identifier: Apache-2.0

package distro

import (
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/export"
	"go.uber.org/zap"
)

type devIdFilterDetails struct {
	deviceIDs []string
}

func newDevIdFilter(filter export.Filter) Filterer {

	filterer := devIdFilterDetails{
		deviceIDs: filter.DeviceIDs,
	}
	return filterer
}

func (filter devIdFilterDetails) Filter(event *models.Event) (bool, *models.Event) {

	if event == nil {
		return false, nil
	}

	for _, devId := range filter.deviceIDs {
		if event.Device == devId {
			logger.Debug("Event accepted", zap.Any("Event", event))
			return true, event
		}
	}
	return false, event
}

type valueDescFilterDetails struct {
	valueDescIDs []string
}

func newValueDescFilter(filter export.Filter) Filterer {
	filterer := valueDescFilterDetails{
		valueDescIDs: filter.ValueDescriptorIDs,
	}
	return filterer
}

func (filter valueDescFilterDetails) Filter(event *models.Event) (bool, *models.Event) {

	if event == nil {
		return false, nil
	}

	auxEvent := &models.Event{
		Pushed:   event.Pushed,
		Device:   event.Device,
		Created:  event.Created,
		Modified: event.Modified,
		Origin:   event.Origin,
		Readings: []models.Reading{},
	}

	for _, filterId := range filter.valueDescIDs {
		for _, reading := range event.Readings {
			if reading.Name == filterId {
				logger.Debug("Reading filtered", zap.Any("Reading", reading))
				auxEvent.Readings = append(auxEvent.Readings, reading)
			}
		}
	}
	return len(auxEvent.Readings) > 0, auxEvent
}
