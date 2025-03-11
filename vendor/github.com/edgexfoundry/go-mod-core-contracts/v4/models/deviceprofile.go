//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type DeviceProfile struct {
	DBTimestamp
	ApiVersion      string
	Description     string
	Id              string
	Name            string
	Manufacturer    string
	Model           string
	Labels          []string
	DeviceResources []DeviceResource
	DeviceCommands  []DeviceCommand
}
