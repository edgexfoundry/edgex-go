//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/query"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

func AllAggregateReadings(aggFunc string, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.AllReadingsAggregation(aggFunc)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	return readings, err
}

func AllAggregateReadingsByTimeRange(aggFunc string, params query.Parameters, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.AllReadingsAggregationByTimeRange(aggFunc, params.Start, params.End)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	return readings, err
}

func AggregateReadingsByResourceName(resourceName string, aggFunc string, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsAggregationByResourceName(resourceName, aggFunc)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	return readings, err
}

func AggregateReadingsByResourceNameAndTimeRange(resourceName string, aggFunc string, params query.Parameters, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsAggregationByResourceNameAndTimeRange(resourceName, aggFunc, params.Start, params.End)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	return readings, err
}

func AggregateReadingsByDeviceName(deviceName string, aggFunc string, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if deviceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsAggregationByDeviceName(deviceName, aggFunc)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	return readings, err
}

func AggregateReadingsByDeviceNameAndTimeRange(deviceName string, aggFunc string, params query.Parameters, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if deviceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsAggregationByDeviceNameAndTimeRange(deviceName, aggFunc, params.Start, params.End)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	return readings, err
}

func AggregateReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, aggFunc string, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if deviceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsAggregationByDeviceNameAndResourceName(deviceName, resourceName, aggFunc)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	return readings, err
}

func AggregateReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, aggFunc string, params query.Parameters, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if deviceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, aggFunc, params.Start, params.End)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	return readings, err
}
