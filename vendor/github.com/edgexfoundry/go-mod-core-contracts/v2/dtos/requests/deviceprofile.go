//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// DeviceProfileRequest defines the Request Content for POST DeviceProfile DTO.
// This object and its properties correspond to the DeviceProfileRequest object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/AddDeviceProfileRequest
type DeviceProfileRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Profile               dtos.DeviceProfile `json:"profile"`
}

// Validate satisfies the Validator interface
func (dp DeviceProfileRequest) Validate() error {
	err := common.Validate(dp)
	if err != nil {
		return err
	}
	return dp.Profile.Validate()
}

// UnmarshalJSON implements the Unmarshaler interface for the DeviceProfileRequest type
func (dp *DeviceProfileRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Profile dtos.DeviceProfile
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*dp = DeviceProfileRequest(alias)

	// validate DeviceProfileRequest DTO
	if err := dp.Validate(); err != nil {
		return err
	}

	// Normalize resource's value type
	for i, resource := range dp.Profile.DeviceResources {
		valueType, err := common.NormalizeValueType(resource.Properties.ValueType)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		dp.Profile.DeviceResources[i].Properties.ValueType = valueType
	}
	return nil
}

// DeviceProfileReqToDeviceProfileModel transforms the DeviceProfileRequest DTO to the DeviceProfile model
func DeviceProfileReqToDeviceProfileModel(addReq DeviceProfileRequest) (DeviceProfiles models.DeviceProfile) {
	return dtos.ToDeviceProfileModel(addReq.Profile)
}

// DeviceProfileReqToDeviceProfileModels transforms the DeviceProfileRequest DTO array to the DeviceProfile model array
func DeviceProfileReqToDeviceProfileModels(addRequests []DeviceProfileRequest) (DeviceProfiles []models.DeviceProfile) {
	for _, req := range addRequests {
		dp := DeviceProfileReqToDeviceProfileModel(req)
		DeviceProfiles = append(DeviceProfiles, dp)
	}
	return DeviceProfiles
}

func NewDeviceProfileRequest(dto dtos.DeviceProfile) DeviceProfileRequest {
	return DeviceProfileRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Profile:     dto,
	}
}
