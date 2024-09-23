//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// Add a new device
func (c *Client) AddDevice(d model.Device) (device model.Device, err errors.EdgeX) {
	return device, nil
}

// DeleteDeviceById deletes a device by id
func (c *Client) DeleteDeviceById(id string) errors.EdgeX {
	return nil
}

// DeleteDeviceByName deletes a device by name
func (c *Client) DeleteDeviceByName(name string) errors.EdgeX {
	return nil
}

// DevicesByServiceName query devices by offset, limit and name
func (c *Client) DevicesByServiceName(offset int, limit int, name string) (ds []model.Device, err errors.EdgeX) {
	return ds, nil
}

// DeviceIdExists checks the device existence by id
func (c *Client) DeviceIdExists(id string) (exist bool, err errors.EdgeX) {
	return exist, nil
}

// DeviceNameExists checks the device existence by name
func (c *Client) DeviceNameExists(id string) (exist bool, err errors.EdgeX) {
	return exist, nil
}

// DeviceById gets a device by id
func (c *Client) DeviceById(id string) (device model.Device, err errors.EdgeX) {
	return device, nil
}

// DeviceByName gets a device by name
func (c *Client) DeviceByName(name string) (device model.Device, err errors.EdgeX) {
	return device, nil
}

// AllDevices query the devices with offset, limit, and labels
func (c *Client) AllDevices(offset int, limit int, labels []string) (ds []model.Device, err errors.EdgeX) {
	return ds, nil
}

// DevicesByProfileName query devices by offset, limit and profile name
func (c *Client) DevicesByProfileName(offset int, limit int, profileName string) (ds []model.Device, err errors.EdgeX) {
	return ds, nil
}

// Update a device
func (c *Client) UpdateDevice(d model.Device) errors.EdgeX {
	return nil
}

// DeviceCountByLabels returns the total count of Devices with labels specified.  If no label is specified, the total count of all devices will be returned.
func (c *Client) DeviceCountByLabels(labels []string) (count uint32, err errors.EdgeX) {
	return count, nil
}

// DeviceCountByProfileName returns the count of Devices associated with specified profile
func (c *Client) DeviceCountByProfileName(profileName string) (count uint32, err errors.EdgeX) {
	return count, nil
}

// DeviceCountByServiceName returns the count of Devices associated with specified service
func (c *Client) DeviceCountByServiceName(serviceName string) (count uint32, err errors.EdgeX) {
	return count, nil
}
