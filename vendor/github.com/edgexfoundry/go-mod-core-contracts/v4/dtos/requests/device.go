//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// AddDeviceRequest defines the Request Content for POST Device DTO.
type AddDeviceRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Device                dtos.Device `json:"device"`
}

// Validate satisfies the Validator interface
func (d AddDeviceRequest) Validate() error {
	err := common.Validate(d)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the AddDeviceRequest type
func (d *AddDeviceRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Device dtos.Device
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	if alias.Device.Properties == nil {
		alias.Device.Properties = make(map[string]any)
	}

	*d = AddDeviceRequest(alias)

	// validate AddDeviceRequest DTO
	if err := d.Validate(); err != nil {
		return err
	}
	return nil
}

// AddDeviceReqToDeviceModels transforms the AddDeviceRequest DTO array to the Device model array
func AddDeviceReqToDeviceModels(addRequests []AddDeviceRequest) (Devices []models.Device) {
	for _, req := range addRequests {
		d := dtos.ToDeviceModel(req.Device)
		Devices = append(Devices, d)
	}
	return Devices
}

// UpdateDeviceRequest defines the Request Content for PUT event as pushed DTO.
type UpdateDeviceRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Device                dtos.UpdateDevice `json:"device"`
}

// Validate satisfies the Validator interface
func (d UpdateDeviceRequest) Validate() error {
	err := common.Validate(d)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the UpdateDeviceRequest type
func (d *UpdateDeviceRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Device dtos.UpdateDevice
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*d = UpdateDeviceRequest(alias)

	// validate UpdateDeviceRequest DTO
	if err := d.Validate(); err != nil {
		return err
	}
	return nil
}

// ReplaceDeviceModelFieldsWithDTO replace existing Device's fields with DTO patch
func ReplaceDeviceModelFieldsWithDTO(device *models.Device, patch dtos.UpdateDevice) {
	if patch.Description != nil {
		device.Description = *patch.Description
	}
	if patch.Parent != nil {
		device.Parent = *patch.Parent
	}
	if patch.AdminState != nil {
		device.AdminState = models.AdminState(*patch.AdminState)
	}
	if patch.OperatingState != nil {
		device.OperatingState = models.OperatingState(*patch.OperatingState)
	}
	if patch.ServiceName != nil {
		device.ServiceName = *patch.ServiceName
	}
	if patch.ProfileName != nil {
		device.ProfileName = *patch.ProfileName
	}
	if patch.Labels != nil {
		device.Labels = patch.Labels
	}
	if patch.Location != nil {
		device.Location = patch.Location
	}
	if patch.AutoEvents != nil {
		device.AutoEvents = dtos.ToAutoEventModels(patch.AutoEvents)
	}
	if patch.Protocols != nil {
		device.Protocols = dtos.ToProtocolModels(patch.Protocols)
	}
	if patch.Tags != nil {
		device.Tags = patch.Tags
	}
	if patch.Properties != nil {
		device.Properties = patch.Properties
	}
}

func NewAddDeviceRequest(dto dtos.Device) AddDeviceRequest {
	return AddDeviceRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Device:      dto,
	}
}

func NewUpdateDeviceRequest(dto dtos.UpdateDevice) UpdateDeviceRequest {
	return UpdateDeviceRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Device:      dto,
	}
}
