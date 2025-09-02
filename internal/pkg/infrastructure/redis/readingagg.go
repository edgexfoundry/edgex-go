//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func (c *Client) AllReadingsAggregation(aggregateFunc string) ([]models.Reading, errors.EdgeX) {
	c.loggingClient.Warn("AllReadingsAggregation function didn't implement")
	return nil, nil
}

func (c *Client) AllReadingsAggregationByTimeRange(aggregateFun string, start int64, end int64) ([]models.Reading, errors.EdgeX) {
	c.loggingClient.Warn("AllReadingsAggregationByTimeRange function didn't implement")
	return nil, nil
}

func (c *Client) ReadingsAggregationByResourceName(resourceName string, aggregateFunc string) ([]models.Reading, errors.EdgeX) {
	c.loggingClient.Warn("ReadingsAggregationByResourceName function didn't implement")
	return nil, nil
}

func (c *Client) ReadingsAggregationByResourceNameAndTimeRange(resourceName string, aggregateFun string, start int64, end int64) ([]models.Reading, errors.EdgeX) {
	c.loggingClient.Warn("ReadingsAggregationByResourceNameAndTimeRange function didn't implement")
	return nil, nil
}

func (c *Client) ReadingsAggregationByDeviceName(deviceName string, aggregateFunc string) ([]models.Reading, errors.EdgeX) {
	c.loggingClient.Warn("ReadingsAggregationByDeviceName function didn't implement")
	return nil, nil
}

func (c *Client) ReadingsAggregationByDeviceNameAndTimeRange(deviceName string, aggregateFun string, start int64, end int64) ([]models.Reading, errors.EdgeX) {
	c.loggingClient.Warn("ReadingsAggregationByDeviceNameAndTimeRange function didn't implement")
	return nil, nil
}

func (c *Client) ReadingsAggregationByDeviceNameAndResourceName(deviceName string, resourceName string, aggregateFunc string) ([]models.Reading, errors.EdgeX) {
	c.loggingClient.Warn("AllReadingsAggregation function didn't implement")
	return nil, nil
}

func (c *Client) ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, aggregateFunc string, start int64, end int64) ([]models.Reading, errors.EdgeX) {
	c.loggingClient.Warn("ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange function didn't implement")
	return nil, nil
}
