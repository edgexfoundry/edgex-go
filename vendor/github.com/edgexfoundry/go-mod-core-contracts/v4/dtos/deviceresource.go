//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type DeviceResource struct {
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Name        string                 `json:"name" yaml:"name" validate:"required,edgex-dto-none-empty-string"`
	IsHidden    bool                   `json:"isHidden" yaml:"isHidden"`
	Properties  ResourceProperties     `json:"properties" yaml:"properties"`
	Attributes  map[string]interface{} `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Tags        map[string]any         `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Validate satisfies the Validator interface
func (dr *DeviceResource) Validate() error {
	err := common.Validate(dr)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid DeviceResource.", err)
	}

	return nil
}

type UpdateDeviceResource struct {
	Description *string `json:"description"`
	Name        *string `json:"name" validate:"required,edgex-dto-none-empty-string"`
	IsHidden    *bool   `json:"isHidden"`
}

// ToDeviceResourceModel transforms the DeviceResource DTO to the DeviceResource model
func ToDeviceResourceModel(d DeviceResource) models.DeviceResource {
	return models.DeviceResource{
		Description: d.Description,
		Name:        d.Name,
		IsHidden:    d.IsHidden,
		Properties:  ToResourcePropertiesModel(d.Properties),
		Attributes:  d.Attributes,
		Tags:        d.Tags,
	}
}

// ToDeviceResourceModels transforms the DeviceResource DTOs to the DeviceResource models
func ToDeviceResourceModels(deviceResourceDTOs []DeviceResource) []models.DeviceResource {
	deviceResourceModels := make([]models.DeviceResource, len(deviceResourceDTOs))
	for i, d := range deviceResourceDTOs {
		deviceResourceModels[i] = ToDeviceResourceModel(d)
	}
	return deviceResourceModels
}

// FromDeviceResourceModelToDTO transforms the DeviceResource model to the DeviceResource DTO
func FromDeviceResourceModelToDTO(d models.DeviceResource) DeviceResource {
	return DeviceResource{
		Description: d.Description,
		Name:        d.Name,
		IsHidden:    d.IsHidden,
		Properties:  FromResourcePropertiesModelToDTO(d.Properties),
		Attributes:  d.Attributes,
		Tags:        d.Tags,
	}
}

// FromDeviceResourceModelsToDTOs transforms the DeviceResource models to the DeviceResource DTOs
func FromDeviceResourceModelsToDTOs(deviceResourceModels []models.DeviceResource) []DeviceResource {
	deviceResourceDTOs := make([]DeviceResource, len(deviceResourceModels))
	for i, d := range deviceResourceModels {
		deviceResourceDTOs[i] = FromDeviceResourceModelToDTO(d)
	}
	return deviceResourceDTOs
}
