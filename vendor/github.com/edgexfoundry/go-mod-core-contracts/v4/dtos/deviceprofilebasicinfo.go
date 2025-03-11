//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

type DeviceProfileBasicInfo struct {
	DBTimestamp  `json:",inline" yaml:"dbTimestamp,omitempty"`
	Id           string   `json:"id,omitempty" validate:"omitempty,uuid" yaml:"id,omitempty"`
	Name         string   `json:"name" yaml:"name" validate:"required,edgex-dto-none-empty-string"`
	Manufacturer string   `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
	Description  string   `json:"description,omitempty" yaml:"description,omitempty"`
	Model        string   `json:"model,omitempty" yaml:"model,omitempty"`
	Labels       []string `json:"labels,omitempty" yaml:"labels,flow,omitempty"`
}

type UpdateDeviceProfileBasicInfo struct {
	Id           *string  `json:"id" validate:"required_without=Name,edgex-dto-uuid"`
	Name         *string  `json:"name" validate:"required_without=Id,edgex-dto-none-empty-string"`
	Manufacturer *string  `json:"manufacturer"`
	Description  *string  `json:"description"`
	Model        *string  `json:"model"`
	Labels       []string `json:"labels"`
}
