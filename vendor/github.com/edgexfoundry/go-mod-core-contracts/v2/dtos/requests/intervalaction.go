//
// Copyright (C) 2021 IOTech Ltd
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

// AddIntervalRequest defines the Request Content for POST Interval DTO.
// This object and its properties correspond to the AddIntervalRequest object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-scheduler/2.1.0#/AddIntervalActionRequest
type AddIntervalActionRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Action                dtos.IntervalAction `json:"action"`
}

// Validate satisfies the Validator interface
func (request AddIntervalActionRequest) Validate() error {
	err := common.Validate(request)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = request.Action.Address.Validate()
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// UnmarshalJSON implements the Unmarshaler interface for the AddIntervalActionRequest type
func (request *AddIntervalActionRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Action dtos.IntervalAction
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*request = AddIntervalActionRequest(alias)

	// validate AddIntervalActionRequest DTO
	if err := request.Validate(); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// AddIntervalActionReqToIntervalActionModels transforms the AddIntervalActionRequest DTO array to the IntervalAction model array
func AddIntervalActionReqToIntervalActionModels(addRequests []AddIntervalActionRequest) (actions []models.IntervalAction) {
	for _, req := range addRequests {
		d := dtos.ToIntervalActionModel(req.Action)
		actions = append(actions, d)
	}
	return actions
}

// UpdateIntervalActionRequest defines the Request Content for PUT event as pushed DTO.
// This object and its properties correspond to the UpdateIntervalActionRequest object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-scheduler/2.1.0#/UpdateIntervalActionRequest
type UpdateIntervalActionRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Action                dtos.UpdateIntervalAction `json:"action"`
}

// Validate satisfies the Validator interface
func (request UpdateIntervalActionRequest) Validate() error {
	err := common.Validate(request)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if request.Action.Address != nil {
		err = request.Action.Address.Validate()
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	return nil
}

// UnmarshalJSON implements the Unmarshaler interface for the UpdateIntervalActionRequest type
func (request *UpdateIntervalActionRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Action dtos.UpdateIntervalAction
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*request = UpdateIntervalActionRequest(alias)

	// validate UpdateIntervalActionRequest DTO
	if err := request.Validate(); err != nil {
		return err
	}
	return nil
}

// ReplaceIntervalActionModelFieldsWithDTO replace existing IntervalAction's fields with DTO patch
func ReplaceIntervalActionModelFieldsWithDTO(action *models.IntervalAction, patch dtos.UpdateIntervalAction) {
	if patch.IntervalName != nil {
		action.IntervalName = *patch.IntervalName
	}
	if patch.Address != nil {
		action.Address = dtos.ToAddressModel(*patch.Address)
	}
	if patch.Content != nil {
		action.Content = *patch.Content
	}
	if patch.ContentType != nil {
		action.ContentType = *patch.ContentType
	}
	if patch.AdminState != nil {
		action.AdminState = models.AdminState(*patch.AdminState)
	}
}

func NewAddIntervalActionRequest(dto dtos.IntervalAction) AddIntervalActionRequest {
	return AddIntervalActionRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Action:      dto,
	}
}

func NewUpdateIntervalActionRequest(dto dtos.UpdateIntervalAction) UpdateIntervalActionRequest {
	return UpdateIntervalActionRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Action:      dto,
	}
}
