//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type DeviceService struct {
	DBTimestamp `json:",inline"`
	Id          string         `json:"id,omitempty" validate:"omitempty,uuid"`
	Name        string         `json:"name" validate:"required,edgex-dto-none-empty-string"`
	Description string         `json:"description,omitempty"`
	Labels      []string       `json:"labels,omitempty"`
	BaseAddress string         `json:"baseAddress" validate:"required,uri"`
	AdminState  string         `json:"adminState" validate:"oneof='LOCKED' 'UNLOCKED'"`
	Properties  map[string]any `json:"properties" yaml:"properties"`
}

type UpdateDeviceService struct {
	Id          *string        `json:"id" validate:"required_without=Name,edgex-dto-uuid"`
	Name        *string        `json:"name" validate:"required_without=Id,edgex-dto-none-empty-string"`
	Description *string        `json:"description"`
	BaseAddress *string        `json:"baseAddress" validate:"omitempty,uri"`
	Labels      []string       `json:"labels"`
	AdminState  *string        `json:"adminState" validate:"omitempty,oneof='LOCKED' 'UNLOCKED'"`
	Properties  map[string]any `json:"properties"`
}

// ToDeviceServiceModel transforms the DeviceService DTO to the DeviceService Model
func ToDeviceServiceModel(dto DeviceService) models.DeviceService {
	var ds models.DeviceService
	ds.Id = dto.Id
	ds.Name = dto.Name
	ds.Description = dto.Description
	ds.BaseAddress = dto.BaseAddress
	ds.Labels = dto.Labels
	ds.AdminState = models.AdminState(dto.AdminState)
	ds.Properties = dto.Properties
	if ds.Properties == nil {
		ds.Properties = make(map[string]any)
	}
	return ds
}

// FromDeviceServiceModelToDTO transforms the DeviceService Model to the DeviceService DTO
func FromDeviceServiceModelToDTO(ds models.DeviceService) DeviceService {
	var dto DeviceService
	dto.DBTimestamp = DBTimestamp(ds.DBTimestamp)
	dto.Id = ds.Id
	dto.Name = ds.Name
	dto.Description = ds.Description
	dto.BaseAddress = ds.BaseAddress
	dto.Labels = ds.Labels
	dto.AdminState = string(ds.AdminState)
	dto.Properties = ds.Properties
	if dto.Properties == nil {
		dto.Properties = make(map[string]any)
	}
	return dto
}

// FromDeviceServiceModelToUpdateDTO transforms the DeviceService Model to the UpdateDeviceService DTO
func FromDeviceServiceModelToUpdateDTO(ds models.DeviceService) UpdateDeviceService {
	adminState := string(ds.AdminState)
	dto := UpdateDeviceService{
		Id:          &ds.Id,
		Name:        &ds.Name,
		Description: &ds.Description,
		Labels:      ds.Labels,
		BaseAddress: &ds.BaseAddress,
		AdminState:  &adminState,
		Properties:  ds.Properties,
	}
	return dto
}
