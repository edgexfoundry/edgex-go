//
// Copyright (C) 2022 IOTech Ltd
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

// AddDeviceCommandRequest defines the Request Content for POST DeviceCommand DTO.
type AddDeviceCommandRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	ProfileName           string             `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	DeviceCommand         dtos.DeviceCommand `json:"deviceCommand"`
}

// Validate satisfies the Validator interface
func (request AddDeviceCommandRequest) Validate() error {
	err := common.Validate(request)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the AddDeviceCommandRequest type
func (dc *AddDeviceCommandRequest) UnmarshalJSON(b []byte) error {
	alias := struct {
		dtoCommon.BaseRequest
		ProfileName   string
		DeviceCommand dtos.DeviceCommand
	}{}

	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}
	*dc = AddDeviceCommandRequest(alias)

	if err := dc.Validate(); err != nil {
		return err
	}

	return nil
}

// UpdateDeviceCommandRequest defines the Request Content for PATCH DeviceCommand DTO.
type UpdateDeviceCommandRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	ProfileName           string                   `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	DeviceCommand         dtos.UpdateDeviceCommand `json:"deviceCommand"`
}

// Validate satisfies the Validator interface
func (request UpdateDeviceCommandRequest) Validate() error {
	err := common.Validate(request)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the UpdateDeviceCommandRequest type
func (dc *UpdateDeviceCommandRequest) UnmarshalJSON(b []byte) error {
	alias := struct {
		dtoCommon.BaseRequest
		ProfileName   string
		DeviceCommand dtos.UpdateDeviceCommand
	}{}

	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}
	*dc = UpdateDeviceCommandRequest(alias)

	if err := dc.Validate(); err != nil {
		return err
	}

	return nil
}

// ReplaceDeviceCommandModelFieldsWithDTO replace existing DeviceCommand's fields with DTO patch
func ReplaceDeviceCommandModelFieldsWithDTO(dc *models.DeviceCommand, patch dtos.UpdateDeviceCommand) {
	if patch.IsHidden != nil {
		dc.IsHidden = *patch.IsHidden
	}
}
