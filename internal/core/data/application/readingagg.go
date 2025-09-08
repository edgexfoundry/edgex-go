//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/data/query"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/spf13/cast"
)

func AllAggregateReadings(aggFunc string, dic *di.Container, params query.Parameters) (readings []dtos.BaseReading, err errors.EdgeX) {
	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClient.AllReadingsAggregation(aggFunc, params.Offset, params.Limit)
	}
	return getReadingAggregation(dic, aggDBFunc)
}

func AllAggregateReadingsByTimeRange(aggFunc string, params query.Parameters, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClient.AllReadingsAggregationByTimeRange(aggFunc, params.Start, params.End, params.Offset, params.Limit)
	}
	return getReadingAggregation(dic, aggDBFunc)
}

func AggregateReadingsByResourceName(resourceName string, aggFunc string, dic *di.Container, params query.Parameters) (readings []dtos.BaseReading, err errors.EdgeX) {
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClient.ReadingsAggregationByResourceName(resourceName, aggFunc, params.Offset, params.Limit)
	}
	return getReadingAggregation(dic, aggDBFunc)
}

func AggregateReadingsByResourceNameAndTimeRange(resourceName string, aggFunc string, params query.Parameters, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClient.ReadingsAggregationByResourceNameAndTimeRange(resourceName, aggFunc, params.Start, params.End, params.Offset, params.Limit)
	}
	return getReadingAggregation(dic, aggDBFunc)
}

func AggregateReadingsByDeviceName(deviceName string, aggFunc string, dic *di.Container, params query.Parameters) (readings []dtos.BaseReading, err errors.EdgeX) {
	if deviceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClient.ReadingsAggregationByDeviceName(deviceName, aggFunc, params.Offset, params.Limit)
	}
	return getReadingAggregation(dic, aggDBFunc)
}

func AggregateReadingsByDeviceNameAndTimeRange(deviceName string, aggFunc string, params query.Parameters, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if deviceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClient.ReadingsAggregationByDeviceNameAndTimeRange(deviceName, aggFunc, params.Start, params.End, params.Offset, params.Limit)
	}
	return getReadingAggregation(dic, aggDBFunc)
}

func AggregateReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, aggFunc string, dic *di.Container, params query.Parameters) (readings []dtos.BaseReading, err errors.EdgeX) {
	if deviceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClient.ReadingsAggregationByDeviceNameAndResourceName(deviceName, resourceName, aggFunc, params.Offset, params.Limit)
	}
	return getReadingAggregation(dic, aggDBFunc)
}

func AggregateReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, aggFunc string, params query.Parameters, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if deviceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	aggDBFunc := func(dbClient interfaces.DBClient) ([]models.Reading, errors.EdgeX) {
		return dbClient.ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, aggFunc, params.Start, params.End, params.Offset, params.Limit)
	}
	return getReadingAggregation(dic, aggDBFunc)
}

// getReadingAggregation calls the provided infrastructure layer function to obtain the aggregated readings, and converts Reading models to DTOs
func getReadingAggregation(dic *di.Container, aggReadingDBFunc func(interfaces.DBClient) ([]models.Reading, errors.EdgeX)) (readings []dtos.BaseReading, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := aggReadingDBFunc(dbClient)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}

	convertAggregateReadingsToNumeric(readings)
	return readings, nil
}

// convertAggregateReadingsToNumeric iterates the BaseReading DTOs and normalizes their Value field into a numeric representation
func convertAggregateReadingsToNumeric(readings []dtos.BaseReading) {
	for i := range readings {
		// Skip the reading with Value field is empty
		if len(readings[i].Value) == 0 {
			continue
		}

		readingPtr := &readings[i]
		readingPtr.NumericReading.NumericValue = cast.ToFloat64(readingPtr.Value)
	}
}
