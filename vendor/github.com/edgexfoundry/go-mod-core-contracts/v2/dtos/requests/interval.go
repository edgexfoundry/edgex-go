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
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-scheduler/2.1.0#/AddIntervalRequest
type AddIntervalRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Interval              dtos.Interval `json:"interval"`
}

// Validate satisfies the Validator interface
func (request AddIntervalRequest) Validate() error {
	err := common.Validate(request)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the AddIntervalRequest type
func (request *AddIntervalRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Interval dtos.Interval
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*request = AddIntervalRequest(alias)

	// validate AddIntervalRequest DTO
	if err := request.Validate(); err != nil {
		return err
	}
	return nil
}

// AddIntervalReqToIntervalModels transforms the AddIntervalRequest DTO array to the Interval model array
func AddIntervalReqToIntervalModels(addRequests []AddIntervalRequest) (intervals []models.Interval) {
	for _, req := range addRequests {
		d := dtos.ToIntervalModel(req.Interval)
		intervals = append(intervals, d)
	}
	return intervals
}

// UpdateIntervalRequest defines the Request Content for PUT event as pushed DTO.
// This object and its properties correspond to the UpdateIntervalRequest object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-scheduler/2.1.0#/UpdateIntervalRequest
type UpdateIntervalRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Interval              dtos.UpdateInterval `json:"interval"`
}

// Validate satisfies the Validator interface
func (request UpdateIntervalRequest) Validate() error {
	err := common.Validate(request)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the UpdateIntervalRequest type
func (request *UpdateIntervalRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Interval dtos.UpdateInterval
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*request = UpdateIntervalRequest(alias)

	// validate UpdateIntervalRequest DTO
	if err := request.Validate(); err != nil {
		return err
	}
	return nil
}

// ReplaceIntervalModelFieldsWithDTO replace existing Interval's fields with DTO patch
func ReplaceIntervalModelFieldsWithDTO(interval *models.Interval, patch dtos.UpdateInterval) {
	if patch.Start != nil {
		interval.Start = *patch.Start
	}
	if patch.End != nil {
		interval.End = *patch.End
	}
	if patch.Interval != nil {
		interval.Interval = *patch.Interval
	}
}

func NewAddIntervalRequest(dto dtos.Interval) AddIntervalRequest {
	return AddIntervalRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Interval:    dto,
	}
}

func NewUpdateIntervalRequest(dto dtos.UpdateInterval) UpdateIntervalRequest {
	return UpdateIntervalRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		Interval:    dto,
	}
}
