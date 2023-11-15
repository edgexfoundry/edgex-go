//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import "github.com/edgexfoundry/go-mod-core-contracts/v3/models"

type DiscoveredDevice struct {
	ProfileName string         `json:"profileName" yaml:"profileName" validate:"len=0|edgex-dto-none-empty-string"`
	AdminState  string         `json:"adminState" yaml:"adminState" validate:"oneof='LOCKED' 'UNLOCKED'"`
	AutoEvents  []AutoEvent    `json:"autoEvents,omitempty" yaml:"autoEvents,omitempty" validate:"dive"`
	Properties  map[string]any `json:"properties,omitempty" yaml:"properties,omitempty"`
}

type UpdateDiscoveredDevice struct {
	ProfileName *string        `json:"profileName" validate:"omitempty,len=0|edgex-dto-none-empty-string"`
	AdminState  *string        `json:"adminState" validate:"omitempty,oneof='LOCKED' 'UNLOCKED'"`
	AutoEvents  []AutoEvent    `json:"autoEvents" validate:"dive"`
	Properties  map[string]any `json:"properties"`
}

func ToDiscoveredDeviceModel(dto DiscoveredDevice) models.DiscoveredDevice {
	return models.DiscoveredDevice{
		ProfileName: dto.ProfileName,
		AdminState:  models.AdminState(dto.AdminState),
		AutoEvents:  ToAutoEventModels(dto.AutoEvents),
		Properties:  dto.Properties,
	}
}

func FromDiscoveredDeviceModelToDTO(d models.DiscoveredDevice) DiscoveredDevice {
	return DiscoveredDevice{
		ProfileName: d.ProfileName,
		AdminState:  string(d.AdminState),
		AutoEvents:  FromAutoEventModelsToDTOs(d.AutoEvents),
		Properties:  d.Properties,
	}
}

func FromDiscoveredDeviceModelToUpdateDTO(d models.DiscoveredDevice) UpdateDiscoveredDevice {
	adminState := string(d.AdminState)
	return UpdateDiscoveredDevice{
		ProfileName: &d.ProfileName,
		AdminState:  &adminState,
		AutoEvents:  FromAutoEventModelsToDTOs(d.AutoEvents),
		Properties:  d.Properties,
	}
}
