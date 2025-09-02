//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/jackc/pgx/v5"
)

// AllReadingsAggregation queries aggregated reading values using the specified SQL aggregation function.
func (c *Client) AllReadingsAggregation(aggregateFun string) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithConds(aggregateFun, false)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, pgx.NamedArgs{})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to get the aggregated readings", err)
	}
	return readings, nil
}

// AllReadingsAggregationByTimeRange queries aggregated reading values within the given time range using the specified SQL aggregation function.
func (c *Client) AllReadingsAggregationByTimeRange(aggregateFun string, start int64, end int64) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithConds(aggregateFun, true)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, pgx.NamedArgs{startTimeCondition: start, endTimeCondition: end})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to get the aggregated readings within the time range - start: %d, end: %d", start, end), err)
	}
	return readings, nil
}

// ReadingsAggregationByResourceName queries aggregated reading values by resource name using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByResourceName(resourceName string, aggregateFun string) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithConds(aggregateFun, false, resourceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, pgx.NamedArgs{resourceNameCol: resourceName})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by resource name '%s'", resourceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByResourceNameAndTimeRange queries aggregated reading values by resource name within the given time range using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByResourceNameAndTimeRange(resourceName string, aggregateFun string, start int64, end int64) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithConds(aggregateFun, true, resourceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, pgx.NamedArgs{resourceNameCol: resourceName, startTimeCondition: start, endTimeCondition: end})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by resource name '%s'", resourceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByDeviceName queries aggregated reading values by device name using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByDeviceName(deviceName string, aggregateFun string) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithConds(aggregateFun, false, deviceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, pgx.NamedArgs{deviceNameCol: deviceName})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device name '%s'", deviceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByDeviceNameAndTimeRange queries aggregated reading values by device name within the given time range using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByDeviceNameAndTimeRange(deviceName string, aggregateFun string, start int64, end int64) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithConds(aggregateFun, true, deviceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, pgx.NamedArgs{deviceNameCol: deviceName, startTimeCondition: start, endTimeCondition: end})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device name '%s'", deviceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByDeviceNameAndResourceName queries aggregated reading values by device name & resource name using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByDeviceNameAndResourceName(deviceName string, resourceName string, aggregateFun string) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithConds(aggregateFun, false, deviceNameCol, resourceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement,
		pgx.NamedArgs{deviceNameCol: deviceName, resourceNameCol: resourceName})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device '%s' and resource '%s'", deviceName, resourceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange queries aggregated reading values by device name & resource name within the given time range using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, aggregateFun string, start int64, end int64) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithConds(aggregateFun, true, deviceNameCol, resourceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement,
		pgx.NamedArgs{deviceNameCol: deviceName, resourceNameCol: resourceName, startTimeCondition: start, endTimeCondition: end})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device '%s' and resource '%s'", deviceName, resourceName), err)
	}
	return readings, nil
}
