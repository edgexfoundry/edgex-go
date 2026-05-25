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

// AddDeviceServiceRequest defines the Request Content for POST DeviceService DTO.
type AddDeviceServiceRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Service               dtos.DeviceService `json:"service"`
}

// Validate satisfies the Validator interface
func (ds AddDeviceServiceRequest) Validate() error {
	err := common.Validate(ds)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the AddDeviceServiceRequest type
func (ds *AddDeviceServiceRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Service dtos.DeviceService
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	if alias.Service.Properties == nil {
		alias.Service.Properties = make(map[string]any)
	}

	*ds = AddDeviceServiceRequest(alias)

	// validate AddDeviceServiceRequest DTO
	if err := ds.Validate(); err != nil {
		return err
	}
	return nil
}

// AddDeviceServiceReqToDeviceServiceModels transforms the AddDeviceServiceRequest DTO array to the DeviceService model array
func AddDeviceServiceReqToDeviceServiceModels(addRequests []AddDeviceServiceRequest) (DeviceServices []models.DeviceService) {
	for _, req := range addRequests {
		ds := dtos.ToDeviceServiceModel(req.Service)
		DeviceServices = append(DeviceServices, ds)
	}
	return DeviceServices
}

// UpdateDeviceServiceRequest defines the Request Content for PUT event as pushed DTO.
type UpdateDeviceServiceRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Service               dtos.UpdateDeviceService `json:"service"`
}

// Validate satisfies the Validator interface
func (ds UpdateDeviceServiceRequest) Validate() error {
	err := common.Validate(ds)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the UpdateDeviceServiceRequest type
func (ds *UpdateDeviceServiceRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Service dtos.UpdateDeviceService
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*ds = UpdateDeviceServiceRequest(alias)

	// validate UpdateDeviceServiceRequest DTO
	if err := ds.Validate(); err != nil {
		return err
	}
	return nil
}

// ReplaceDeviceServiceModelFieldsWithDTO replace existing DeviceService's fields with DTO patch
func ReplaceDeviceServiceModelFieldsWithDTO(ds *models.DeviceService, patch dtos.UpdateDeviceService) {
	if patch.Description != nil {
		ds.Description = *patch.Description
	}
	if patch.AdminState != nil {
		ds.AdminState = models.AdminState(*patch.AdminState)
	}
	if patch.Labels != nil {
		ds.Labels = patch.Labels
	}
	if patch.BaseAddress != nil {
		ds.BaseAddress = *patch.BaseAddress
	}
	if patch.Properties != nil {
		ds.Properties = patch.Properties
	}
}

func NewAddDeviceServiceRequest(dto dtos.DeviceService) AddDeviceServiceRequest {
	return AddDeviceServiceRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Service:     dto,
	}
}

func NewUpdateDeviceServiceRequest(dto dtos.UpdateDeviceService) UpdateDeviceServiceRequest {
	return UpdateDeviceServiceRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Service:     dto,
	}
}
