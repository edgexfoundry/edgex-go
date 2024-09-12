//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// Add a new device profile
func (c *Client) AddDeviceProfile(dp model.DeviceProfile) (deviceProfile model.DeviceProfile, err errors.EdgeX) {
	return deviceProfile, nil
}

// UpdateDeviceProfile updates a new device profile
func (c *Client) UpdateDeviceProfile(dp model.DeviceProfile) errors.EdgeX {
	return nil
}

// DeviceProfileById gets a device profile by id
func (c *Client) DeviceProfileById(id string) (deviceProfile model.DeviceProfile, err errors.EdgeX) {
	return deviceProfile, nil
}

// DeviceProfileByName gets a device profile by name
func (c *Client) DeviceProfileByName(name string) (deviceProfile model.DeviceProfile, err errors.EdgeX) {
	return deviceProfile, nil
}

// DeleteDeviceProfileById deletes a device profile by id
func (c *Client) DeleteDeviceProfileById(id string) errors.EdgeX {
	return nil
}

// DeleteDeviceProfileByName deletes a device profile by name
func (c *Client) DeleteDeviceProfileByName(name string) errors.EdgeX {
	return nil
}

// DeviceProfileNameExists checks the device profile exists by name
func (c *Client) DeviceProfileNameExists(name string) (exist bool, err errors.EdgeX) {
	return exist, nil
}

// AllDeviceProfiles query device profiles with offset, limit and labels
func (c *Client) AllDeviceProfiles(offset int, limit int, labels []string) (profiles []model.DeviceProfile, err errors.EdgeX) {
	return profiles, nil
}

// DeviceProfilesByModel query device profiles with offset, limit and model
func (c *Client) DeviceProfilesByModel(offset int, limit int, model string) (profiles []model.DeviceProfile, err errors.EdgeX) {
	return profiles, nil
}

// DeviceProfilesByManufacturer query device profiles with offset, limit and manufacturer
func (c *Client) DeviceProfilesByManufacturer(offset int, limit int, manufacturer string) (profiles []model.DeviceProfile, err errors.EdgeX) {
	return profiles, nil
}

// DeviceProfilesByManufacturerAndModel query device profiles with offset, limit, manufacturer and model
func (c *Client) DeviceProfilesByManufacturerAndModel(offset int, limit int, manufacturer string, model string) (profiles []model.DeviceProfile, count uint32, err errors.EdgeX) {
	return profiles, count, nil
}

// DeviceProfileCountByLabels returns the total count of Device Profiles with labels specified.  If no label is specified, the total count of all device profiles will be returned.
func (c *Client) DeviceProfileCountByLabels(labels []string) (count uint32, err errors.EdgeX) {
	return count, nil
}

// DeviceProfileCountByManufacturer returns the count of Device Profiles associated with specified manufacturer
func (c *Client) DeviceProfileCountByManufacturer(manufacturer string) (count uint32, err errors.EdgeX) {
	return count, nil
}

// DeviceProfileCountByModel returns the count of Device Profiles associated with specified model
func (c *Client) DeviceProfileCountByModel(model string) (count uint32, err errors.EdgeX) {
	return count, nil
}
