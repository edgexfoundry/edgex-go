//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/models"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type DBClient interface {
	CloseSession()

	AddEvent(e model.Event) (model.Event, errors.EdgeX)
	EventById(id string) (model.Event, errors.EdgeX)
	DeleteEventById(id string) errors.EdgeX
	EventTotalCount() (int64, errors.EdgeX)
	EventCountByDeviceName(deviceName string) (int64, errors.EdgeX)
	EventCountByTimeRange(start int64, end int64) (int64, errors.EdgeX)
	EventCountByDeviceNameAndSourceNameAndLimit(deviceName, sourceName string, limit int) (int64, errors.EdgeX)
	AllEvents(offset int, limit int) ([]model.Event, errors.EdgeX)
	EventsByDeviceName(offset int, limit int, name string) ([]model.Event, errors.EdgeX)
	DeleteEventsByDeviceName(deviceName string) errors.EdgeX
	DeleteEventsByDeviceNameAndSourceName(deviceName, sourceName string) errors.EdgeX
	EventsByTimeRange(start int64, end int64, offset int, limit int) ([]model.Event, errors.EdgeX)
	DeleteEventsByAge(age int64) errors.EdgeX
	DeleteEventsByAgeAndDeviceNameAndSourceName(age int64, deviceName, sourceName string) errors.EdgeX
	ReadingTotalCount() (int64, errors.EdgeX)
	AllReadings(offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsByTimeRange(start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsByResourceName(offset int, limit int, resourceName string) ([]model.Reading, errors.EdgeX)
	ReadingsByDeviceName(offset int, limit int, name string) ([]model.Reading, errors.EdgeX)
	ReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingCountByDeviceName(deviceName string) (int64, errors.EdgeX)
	ReadingCountByResourceName(resourceName string) (int64, errors.EdgeX)
	ReadingCountByResourceNameAndTimeRange(resourceName string, start int64, end int64) (int64, errors.EdgeX)
	ReadingCountByDeviceNameAndResourceName(deviceName string, resourceName string) (int64, errors.EdgeX)
	ReadingCountByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int64, end int64) (int64, errors.EdgeX)
	ReadingCountByTimeRange(start int64, end int64) (int64, errors.EdgeX)
	ReadingsByResourceNameAndTimeRange(resourceName string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start int64, end int64, offset, limit int) ([]model.Reading, errors.EdgeX)
	ReadingCountByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceName []string, start int64, end int64) (int64, errors.EdgeX)
	ReadingsByDeviceNameAndTimeRange(deviceName string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingCountByDeviceNameAndTimeRange(deviceName string, start int64, end int64) (int64, errors.EdgeX)
	LatestReadingByOffset(offset uint32) (model.Reading, errors.EdgeX)
	LatestEventByDeviceNameAndSourceNameAndOffset(deviceName string, sourceName string, offset int64) (model.Event, errors.EdgeX)
	LatestEventByDeviceNameAndSourceNameAndAgeAndOffset(deviceName string, sourceName string, age, offset int64) (model.Event, errors.EdgeX)

	AllDeviceInfos(offset int, limit int) ([]models.DeviceInfo, errors.EdgeX)

	AllReadingsAggregation(aggregateFunc string, offset int, limit int) ([]model.Reading, errors.EdgeX)
	AllReadingsAggregationByTimeRange(aggregateFun string, start, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsAggregationByResourceName(resourceName string, aggregateFunc string, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsAggregationByResourceNameAndTimeRange(resourceName string, aggregateFun string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsAggregationByDeviceName(deviceName string, aggregateFunc string, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsAggregationByDeviceNameAndTimeRange(deviceName string, aggregateFun string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsAggregationByDeviceNameAndResourceName(deviceName string, resourceName string, aggregateFunc string, offset int, limit int) ([]model.Reading, errors.EdgeX)
	ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, aggregateFunc string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX)
}
