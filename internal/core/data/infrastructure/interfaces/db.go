//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type DBClient interface {
	CloseSession()

	AddEvent(e model.Event) (model.Event, errors.EdgeX)
	EventById(id string) (model.Event, errors.EdgeX)
	DeleteEventById(id string) errors.EdgeX
	EventTotalCount() (uint32, errors.EdgeX)
	EventCountByDeviceName(deviceName string) (uint32, errors.EdgeX)
	EventCountByTimeRange(start int, end int) (uint32, errors.EdgeX)
	AllEvents(offset int, limit int) ([]model.Event, errors.EdgeX)
	EventsByDeviceName(offset int, limit int, name string) ([]model.Event, errors.EdgeX)
	DeleteEventsByDeviceName(deviceName string) errors.EdgeX
	EventsByTimeRange(start int, end int, offset int, limit int) ([]model.Event, errors.EdgeX)
	DeleteEventsByAge(age int64) errors.EdgeX
	ReadingTotalCount() (uint32, errors.EdgeX)
	AllReadings(offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsByTimeRange(start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsByResourceName(offset int, limit int, resourceName string) ([]model.Reading, errors.EdgeX)
	ReadingsByDeviceName(offset int, limit int, name string) ([]model.Reading, errors.EdgeX)
	ReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingCountByDeviceName(deviceName string) (uint32, errors.EdgeX)
	ReadingCountByResourceName(resourceName string) (uint32, errors.EdgeX)
	ReadingCountByResourceNameAndTimeRange(resourceName string, start int, end int) (uint32, errors.EdgeX)
	ReadingCountByDeviceNameAndResourceName(deviceName string, resourceName string) (uint32, errors.EdgeX)
	ReadingCountByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int, end int) (uint32, errors.EdgeX)
	ReadingCountByTimeRange(start int, end int) (uint32, errors.EdgeX)
	ReadingsByResourceNameAndTimeRange(resourceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start, end, offset, limit int) ([]model.Reading, uint32, errors.EdgeX)
	ReadingsByDeviceNameAndTimeRange(deviceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingCountByDeviceNameAndTimeRange(deviceName string, start int, end int) (uint32, errors.EdgeX)
}
