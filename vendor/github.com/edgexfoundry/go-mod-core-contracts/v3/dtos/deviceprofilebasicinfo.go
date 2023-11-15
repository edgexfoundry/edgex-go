//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

type DeviceProfileBasicInfo struct {
	Id           string   `json:"id" validate:"omitempty,uuid"`
	Name         string   `json:"name" yaml:"name" validate:"required,edgex-dto-none-empty-string"`
	Manufacturer string   `json:"manufacturer" yaml:"manufacturer"`
	Description  string   `json:"description" yaml:"description"`
	Model        string   `json:"model" yaml:"model"`
	Labels       []string `json:"labels" yaml:"labels,flow"`
}

type UpdateDeviceProfileBasicInfo struct {
	Id           *string  `json:"id" validate:"required_without=Name,edgex-dto-uuid"`
	Name         *string  `json:"name" validate:"required_without=Id,edgex-dto-none-empty-string"`
	Manufacturer *string  `json:"manufacturer"`
	Description  *string  `json:"description"`
	Model        *string  `json:"model"`
	Labels       []string `json:"labels"`
}
