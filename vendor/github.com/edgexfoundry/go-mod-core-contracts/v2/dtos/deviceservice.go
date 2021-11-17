//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// DeviceService and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/DeviceService
type DeviceService struct {
	DBTimestamp   `json:",inline"`
	Id            string   `json:"id,omitempty" validate:"omitempty,uuid"`
	Name          string   `json:"name" validate:"required,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	Description   string   `json:"description,omitempty"`
	LastConnected int64    `json:"lastConnected,omitempty"`
	LastReported  int64    `json:"lastReported,omitempty"`
	Labels        []string `json:"labels,omitempty"`
	BaseAddress   string   `json:"baseAddress" validate:"required,uri"`
	AdminState    string   `json:"adminState" validate:"oneof='LOCKED' 'UNLOCKED'"`
}

// UpdateDeviceService and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/UpdateDeviceService
type UpdateDeviceService struct {
	Id            *string  `json:"id" validate:"required_without=Name,edgex-dto-uuid"`
	Name          *string  `json:"name" validate:"required_without=Id,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	Description   *string  `json:"description"`
	LastConnected *int64   `json:"lastConnected"`
	LastReported  *int64   `json:"lastReported"`
	BaseAddress   *string  `json:"baseAddress" validate:"omitempty,uri"`
	Labels        []string `json:"labels"`
	AdminState    *string  `json:"adminState" validate:"omitempty,oneof='LOCKED' 'UNLOCKED'"`
}

// ToDeviceServiceModel transforms the DeviceService DTO to the DeviceService Model
func ToDeviceServiceModel(dto DeviceService) models.DeviceService {
	var ds models.DeviceService
	ds.Id = dto.Id
	ds.Name = dto.Name
	ds.Description = dto.Description
	ds.LastReported = dto.LastReported
	ds.LastConnected = dto.LastConnected
	ds.BaseAddress = dto.BaseAddress
	ds.Labels = dto.Labels
	ds.AdminState = models.AdminState(dto.AdminState)
	return ds
}

// FromDeviceServiceModelToDTO transforms the DeviceService Model to the DeviceService DTO
func FromDeviceServiceModelToDTO(ds models.DeviceService) DeviceService {
	var dto DeviceService
	dto.DBTimestamp = DBTimestamp(ds.DBTimestamp)
	dto.Id = ds.Id
	dto.Name = ds.Name
	dto.Description = ds.Description
	dto.LastReported = ds.LastReported
	dto.LastConnected = ds.LastConnected
	dto.BaseAddress = ds.BaseAddress
	dto.Labels = ds.Labels
	dto.AdminState = string(ds.AdminState)
	return dto
}

// FromDeviceServiceModelToUpdateDTO transforms the DeviceService Model to the UpdateDeviceService DTO
func FromDeviceServiceModelToUpdateDTO(ds models.DeviceService) UpdateDeviceService {
	adminState := string(ds.AdminState)
	dto := UpdateDeviceService{
		Id:            &ds.Id,
		Name:          &ds.Name,
		Description:   &ds.Description,
		Labels:        ds.Labels,
		LastReported:  &ds.LastReported,
		LastConnected: &ds.LastConnected,
		BaseAddress:   &ds.BaseAddress,
		AdminState:    &adminState,
	}
	return dto
}
