//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func (c *Client) ReadingTotalCount() (uint32, errors.EdgeX) {
	return 0, nil
}

func (c *Client) AllReadings(offset int, limit int) ([]model.Reading, errors.EdgeX) {
	return nil, nil
}

func (c *Client) ReadingsByTimeRange(start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	return nil, nil
}

func (c *Client) ReadingsByResourceName(offset int, limit int, resourceName string) ([]model.Reading, errors.EdgeX) {
	return nil, nil
}

func (c *Client) ReadingsByDeviceName(offset int, limit int, name string) ([]model.Reading, errors.EdgeX) {
	return nil, nil
}

func (c *Client) ReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	return nil, nil
}

func (c *Client) ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	return nil, nil
}

func (c *Client) ReadingCountByDeviceName(deviceName string) (uint32, errors.EdgeX) {
	return 0, nil
}

func (c *Client) ReadingCountByResourceName(resourceName string) (uint32, errors.EdgeX) {
	return 0, nil
}

func (c *Client) ReadingCountByResourceNameAndTimeRange(resourceName string, start int, end int) (uint32, errors.EdgeX) {
	return 0, nil
}

func (c *Client) ReadingCountByDeviceNameAndResourceName(deviceName string, resourceName string) (uint32, errors.EdgeX) {
	return 0, nil
}

func (c *Client) ReadingCountByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int, end int) (uint32, errors.EdgeX) {
	return 0, nil
}

func (c *Client) ReadingCountByTimeRange(start int, end int) (uint32, errors.EdgeX) {
	return 0, nil
}

func (c *Client) ReadingsByResourceNameAndTimeRange(resourceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	return nil, nil
}

func (c *Client) ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start, end, offset, limit int) ([]model.Reading, uint32, errors.EdgeX) {
	return nil, 0, nil
}

func (c *Client) ReadingsByDeviceNameAndTimeRange(deviceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	return nil, nil
}

func (c *Client) ReadingCountByDeviceNameAndTimeRange(deviceName string, start int, end int) (uint32, errors.EdgeX) {
	return 0, nil
}

func (c *Client) LatestReadingByOffset(offset uint32) (model.Reading, errors.EdgeX) {
	return nil, nil
}
