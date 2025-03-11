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

// AddDeviceResourceRequest defines the Request Content for POST DeviceResource DTO.
type AddDeviceResourceRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	ProfileName           string              `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	Resource              dtos.DeviceResource `json:"resource"`
}

func (request AddDeviceResourceRequest) Validate() error {
	err := common.Validate(request)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the AddDeviceResourceReques type
func (dr *AddDeviceResourceRequest) UnmarshalJSON(b []byte) error {
	alias := struct {
		dtoCommon.BaseRequest
		ProfileName string
		Resource    dtos.DeviceResource
	}{}

	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}
	*dr = AddDeviceResourceRequest(alias)

	if err := dr.Validate(); err != nil {
		return err
	}

	return nil
}

// UpdateDeviceResourceRequest defines the Request Content for PATCH DeviceResource DTO.
type UpdateDeviceResourceRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	ProfileName           string                    `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	Resource              dtos.UpdateDeviceResource `json:"resource"`
}

func (request UpdateDeviceResourceRequest) Validate() error {
	err := common.Validate(request)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the UpdateDeviceResourceRequest type
func (dr *UpdateDeviceResourceRequest) UnmarshalJSON(b []byte) error {
	alias := struct {
		dtoCommon.BaseRequest
		ProfileName string
		Resource    dtos.UpdateDeviceResource
	}{}

	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}
	*dr = UpdateDeviceResourceRequest(alias)

	if err := dr.Validate(); err != nil {
		return err
	}

	return nil
}

// ReplaceDeviceResourceModelFieldsWithDTO replace existing DeviceResource's fields with DTO patch
func ReplaceDeviceResourceModelFieldsWithDTO(dr *models.DeviceResource, patch dtos.UpdateDeviceResource) {
	if patch.Description != nil {
		dr.Description = *patch.Description
	}
	if patch.IsHidden != nil {
		dr.IsHidden = *patch.IsHidden
	}
}
