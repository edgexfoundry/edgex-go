//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// Device and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/Device
type Device struct {
	DBTimestamp    `json:",inline"`
	Id             string                        `json:"id,omitempty" validate:"omitempty,uuid"`
	Name           string                        `json:"name" validate:"required,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	Description    string                        `json:"description,omitempty"`
	AdminState     string                        `json:"adminState" validate:"oneof='LOCKED' 'UNLOCKED'"`
	OperatingState string                        `json:"operatingState" validate:"oneof='UP' 'DOWN' 'UNKNOWN'"`
	LastConnected  int64                         `json:"lastConnected,omitempty"`
	LastReported   int64                         `json:"lastReported,omitempty"`
	Labels         []string                      `json:"labels,omitempty"`
	Location       interface{}                   `json:"location,omitempty"`
	ServiceName    string                        `json:"serviceName" validate:"required,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	ProfileName    string                        `json:"profileName" validate:"required,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	AutoEvents     []AutoEvent                   `json:"autoEvents,omitempty" validate:"dive"`
	Protocols      map[string]ProtocolProperties `json:"protocols" validate:"required,gt=0"`
}

// UpdateDevice and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/UpdateDevice
type UpdateDevice struct {
	Id             *string                       `json:"id" validate:"required_without=Name,edgex-dto-uuid"`
	Name           *string                       `json:"name" validate:"required_without=Id,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	Description    *string                       `json:"description" validate:"omitempty"`
	AdminState     *string                       `json:"adminState" validate:"omitempty,oneof='LOCKED' 'UNLOCKED'"`
	OperatingState *string                       `json:"operatingState" validate:"omitempty,oneof='UP' 'DOWN' 'UNKNOWN'"`
	LastConnected  *int64                        `json:"lastConnected"`
	LastReported   *int64                        `json:"lastReported"`
	ServiceName    *string                       `json:"serviceName" validate:"omitempty,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	ProfileName    *string                       `json:"profileName" validate:"omitempty,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	Labels         []string                      `json:"labels"`
	Location       interface{}                   `json:"location"`
	AutoEvents     []AutoEvent                   `json:"autoEvents" validate:"dive"`
	Protocols      map[string]ProtocolProperties `json:"protocols" validate:"omitempty,gt=0"`
	Notify         *bool                         `json:"notify"`
}

// ToDeviceModel transforms the Device DTO to the Device Model
func ToDeviceModel(dto Device) models.Device {
	var d models.Device
	d.Id = dto.Id
	d.Name = dto.Name
	d.Description = dto.Description
	d.ServiceName = dto.ServiceName
	d.ProfileName = dto.ProfileName
	d.AdminState = models.AdminState(dto.AdminState)
	d.OperatingState = models.OperatingState(dto.OperatingState)
	d.LastReported = dto.LastReported
	d.LastConnected = dto.LastConnected
	d.Labels = dto.Labels
	d.Location = dto.Location
	d.AutoEvents = ToAutoEventModels(dto.AutoEvents)
	d.Protocols = ToProtocolModels(dto.Protocols)
	return d
}

// FromDeviceModelToDTO transforms the Device Model to the Device DTO
func FromDeviceModelToDTO(d models.Device) Device {
	var dto Device
	dto.DBTimestamp = DBTimestamp(d.DBTimestamp)
	dto.Id = d.Id
	dto.Name = d.Name
	dto.Description = d.Description
	dto.ServiceName = d.ServiceName
	dto.ProfileName = d.ProfileName
	dto.AdminState = string(d.AdminState)
	dto.OperatingState = string(d.OperatingState)
	dto.LastReported = d.LastReported
	dto.LastConnected = d.LastConnected
	dto.Labels = d.Labels
	dto.Location = d.Location
	dto.AutoEvents = FromAutoEventModelsToDTOs(d.AutoEvents)
	dto.Protocols = FromProtocolModelsToDTOs(d.Protocols)
	return dto
}

// FromDeviceModelToUpdateDTO transforms the Device Model to the UpdateDevice DTO
func FromDeviceModelToUpdateDTO(d models.Device) UpdateDevice {
	adminState := string(d.AdminState)
	operatingState := string(d.OperatingState)
	dto := UpdateDevice{
		Id:             &d.Id,
		Name:           &d.Name,
		Description:    &d.Description,
		AdminState:     &adminState,
		OperatingState: &operatingState,
		LastConnected:  &d.LastConnected,
		LastReported:   &d.LastReported,
		ServiceName:    &d.ServiceName,
		ProfileName:    &d.ProfileName,
		Location:       d.Location,
		AutoEvents:     FromAutoEventModelsToDTOs(d.AutoEvents),
		Protocols:      FromProtocolModelsToDTOs(d.Protocols),
		Labels:         d.Labels,
		Notify:         &d.Notify,
	}
	return dto
}
