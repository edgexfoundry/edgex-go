//
// Copyright (C) 2024 IOTech Ltd
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

// AddScheduleJobRequest defines the Request Content for POST ScheduleJob DTO.
type AddScheduleJobRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	ScheduleJob           dtos.ScheduleJob `json:"scheduleJob"`
}

// Validate satisfies the Validator interface
func (a *AddScheduleJobRequest) Validate() error {
	err := common.Validate(a)
	if err != nil {
		return err
	}

	err = a.ScheduleJob.Validate()
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalJSON implements the Unmarshaler interface for the AddScheduleJobRequest type
func (a *AddScheduleJobRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		ScheduleJob dtos.ScheduleJob
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	if alias.ScheduleJob.Properties == nil {
		alias.ScheduleJob.Properties = make(map[string]any)
	}

	*a = AddScheduleJobRequest(alias)

	// validate AddScheduleJobRequest DTO
	if err := a.Validate(); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// UpdateScheduleJobRequest defines the Request Content for PUT event as pushed DTO.
type UpdateScheduleJobRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	ScheduleJob           dtos.UpdateScheduleJob `json:"scheduleJob"`
}

// Validate satisfies the Validator interface
func (u *UpdateScheduleJobRequest) Validate() error {
	err := common.Validate(u)
	if err != nil {
		return err
	}

	if u.ScheduleJob.Definition != nil {
		err = u.ScheduleJob.Definition.Validate()
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleDef.", err)
		}
	}

	if u.ScheduleJob.Actions != nil {
		for _, action := range u.ScheduleJob.Actions {
			err = action.Validate()
			if err != nil {
				return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleAction.", err)
			}
		}
	}

	return nil
}

// UnmarshalJSON implements the Unmarshaler interface for the UpdateScheduleJobRequest type
func (u *UpdateScheduleJobRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		ScheduleJob dtos.UpdateScheduleJob
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*u = UpdateScheduleJobRequest(alias)

	// validate AddScheduleJobRequest DTO
	if err := u.Validate(); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// ReplaceScheduleJobModelFieldsWithDTO replace existing ScheduleJob's fields with DTO patch
func ReplaceScheduleJobModelFieldsWithDTO(ds *models.ScheduleJob, patch dtos.UpdateScheduleJob) {
	if patch.Actions != nil {
		ds.Actions = dtos.ToScheduleActionModels(patch.Actions)
	}
	if patch.AdminState != nil {
		ds.AdminState = models.AdminState(*patch.AdminState)
	}
	if patch.AutoTriggerMissedRecords != nil {
		ds.AutoTriggerMissedRecords = *patch.AutoTriggerMissedRecords
	}
	if patch.Labels != nil {
		ds.Labels = patch.Labels
	}
	if patch.Definition != nil {
		ds.Definition = dtos.ToScheduleDefModel(*patch.Definition)
	}
	if patch.Properties != nil {
		ds.Properties = patch.Properties
	}
}

// NewAddScheduleJobRequest creates, initializes and returns an AddScheduleJobRequest
func NewAddScheduleJobRequest(dto dtos.ScheduleJob) AddScheduleJobRequest {
	return AddScheduleJobRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		ScheduleJob: dto,
	}
}

func NewUpdateScheduleJobRequest(dto dtos.UpdateScheduleJob) UpdateScheduleJobRequest {
	return UpdateScheduleJobRequest{
		BaseRequest: dtoCommon.NewBaseRequest(),
		ScheduleJob: dto,
	}
}
