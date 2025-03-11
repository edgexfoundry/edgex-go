//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import "github.com/edgexfoundry/go-mod-core-contracts/v4/models"

type DeviceCommand struct {
	Name               string              `json:"name" yaml:"name" validate:"required,edgex-dto-none-empty-string"`
	IsHidden           bool                `json:"isHidden" yaml:"isHidden"`
	ReadWrite          string              `json:"readWrite" yaml:"readWrite" validate:"required,oneof='R' 'W' 'RW' 'WR'"`
	ResourceOperations []ResourceOperation `json:"resourceOperations" yaml:"resourceOperations" validate:"gt=0,dive"`
	Tags               map[string]any      `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type UpdateDeviceCommand struct {
	Name     *string `json:"name" validate:"required,edgex-dto-none-empty-string"`
	IsHidden *bool   `json:"isHidden"`
}

// ToDeviceCommandModel transforms the DeviceCommand DTO to the DeviceCommand model
func ToDeviceCommandModel(dto DeviceCommand) models.DeviceCommand {
	resourceOperations := ToResourceOperationModels(dto.ResourceOperations)

	return models.DeviceCommand{
		Name:               dto.Name,
		IsHidden:           dto.IsHidden,
		ReadWrite:          dto.ReadWrite,
		ResourceOperations: resourceOperations,
		Tags:               dto.Tags,
	}
}

// ToDeviceCommandModels transforms the DeviceCommand DTOs to the DeviceCommand models
func ToDeviceCommandModels(deviceCommandDTOs []DeviceCommand) []models.DeviceCommand {
	deviceCommandModels := make([]models.DeviceCommand, len(deviceCommandDTOs))
	for i, p := range deviceCommandDTOs {
		deviceCommandModels[i] = ToDeviceCommandModel(p)
	}
	return deviceCommandModels
}

// FromDeviceCommandModelToDTO transforms the DeviceCommand model to the DeviceCommand DTO
func FromDeviceCommandModelToDTO(d models.DeviceCommand) DeviceCommand {
	resourceOperations := make([]ResourceOperation, len(d.ResourceOperations))
	for i, ro := range d.ResourceOperations {
		resourceOperations[i] = FromResourceOperationModelToDTO(ro)
	}
	return DeviceCommand{
		Name:               d.Name,
		IsHidden:           d.IsHidden,
		ReadWrite:          d.ReadWrite,
		ResourceOperations: resourceOperations,
		Tags:               d.Tags,
	}
}

// FromDeviceCommandModelsToDTOs transforms the DeviceCommand models to the DeviceCommand DTOs
func FromDeviceCommandModelsToDTOs(deviceCommandModels []models.DeviceCommand) []DeviceCommand {
	deviceCommandDTOs := make([]DeviceCommand, len(deviceCommandModels))
	for i, p := range deviceCommandModels {
		deviceCommandDTOs[i] = FromDeviceCommandModelToDTO(p)
	}
	return deviceCommandDTOs
}
