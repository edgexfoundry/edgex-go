//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

type UpdateDeviceProfileTags struct {
	DeviceResources []UpdateTags `json:"deviceResources,omitempty" validate:"dive"`
	DeviceCommands  []UpdateTags `json:"deviceCommands,omitempty" validate:"dive"`
}

type UpdateTags struct {
	Name string         `json:"name" validate:"required,edgex-dto-none-empty-string"`
	Tags map[string]any `json:"tags" validate:"required,gt=0,dive"`
}
