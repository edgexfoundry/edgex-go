//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import "github.com/edgexfoundry/go-mod-core-contracts/v2/models"

// ResourceOperation and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/ResourceOperation
type ResourceOperation struct {
	DeviceResource string            `json:"deviceResource" yaml:"deviceResource" validate:"required"` // The replacement of Object field
	DefaultValue   string            `json:"defaultValue" yaml:"defaultValue"`
	Mappings       map[string]string `json:"mappings" yaml:"mappings"`
}

// ToResourceOperationModel transforms the ResourceOperation DTO to the ResourceOperation model
func ToResourceOperationModel(ro ResourceOperation) models.ResourceOperation {
	return models.ResourceOperation{
		DeviceResource: ro.DeviceResource,
		DefaultValue:   ro.DefaultValue,
		Mappings:       ro.Mappings,
	}
}

// FromResourceOperationModelToDTO transforms the ResourceOperation model to the ResourceOperation DTO
func FromResourceOperationModelToDTO(ro models.ResourceOperation) ResourceOperation {
	return ResourceOperation{
		DeviceResource: ro.DeviceResource,
		DefaultValue:   ro.DefaultValue,
		Mappings:       ro.Mappings,
	}
}
