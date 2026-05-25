//
// Copyright (C) 2020-2025 IOTech Ltd
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

func (profile DeviceProfile) Clone() DeviceProfile {
	cloned := DeviceProfile{
		DBTimestamp:  profile.DBTimestamp,
		ApiVersion:   profile.ApiVersion,
		Description:  profile.Description,
		Id:           profile.Id,
		Name:         profile.Name,
		Manufacturer: profile.Manufacturer,
		Model:        profile.Model,
	}
	if len(profile.Labels) > 0 {
		cloned.Labels = make([]string, len(profile.Labels))
		copy(cloned.Labels, profile.Labels)
	}
	if len(profile.DeviceResources) > 0 {
		cloned.DeviceResources = make([]DeviceResource, len(profile.DeviceResources))
		for i := range profile.DeviceResources {
			cloned.DeviceResources[i] = profile.DeviceResources[i].Clone()
		}
	}
	if len(profile.DeviceCommands) > 0 {
		cloned.DeviceCommands = make([]DeviceCommand, len(profile.DeviceCommands))
		for i := range profile.DeviceCommands {
			cloned.DeviceCommands[i] = profile.DeviceCommands[i].Clone()
		}
	}
	return cloned
}
