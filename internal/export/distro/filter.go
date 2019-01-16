//
// Copyright (c) 2017
// Cavium
//
// SPDX-License-Identifier: Apache-2.0

package distro

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type devIdFilterDetails struct {
	deviceIDs []string
}

func newDevIdFilter(filter models.Filter) filterer {

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
			LoggingClient.Debug(fmt.Sprintf("Event accepted: %s", event.Device))
			return true, event
		}
	}
	return false, event
}

type valueDescFilterDetails struct {
	valueDescIDs []string
}

func newValueDescFilter(filter models.Filter) filterer {
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
				LoggingClient.Debug(fmt.Sprintf("Reading filtered: %s", reading.Name))
				auxEvent.Readings = append(auxEvent.Readings, reading)
			}
		}
	}
	return len(auxEvent.Readings) > 0, auxEvent
}
