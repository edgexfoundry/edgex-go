//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type ScheduleActionRecord struct {
	Id          string         `json:"id,omitempty" validate:"omitempty,uuid"`
	JobName     string         `json:"jobName" validate:"edgex-dto-none-empty-string"`
	Action      ScheduleAction `json:"action" validate:"required"`
	Status      string         `json:"status" validate:"required,oneof='SUCCEEDED' 'FAILED' 'MISSED'"`
	ScheduledAt int64          `json:"scheduledAt,omitempty"`
	Created     int64          `json:"created,omitempty"`
}

// Validate satisfies the Validator interface
func (c *ScheduleActionRecord) Validate() error {
	err := common.Validate(c)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleActionRecord.", err)
	}

	err = c.Action.Validate()
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleAction.", err)
	}

	return nil
}

func ToScheduleActionRecordModel(dto ScheduleActionRecord) models.ScheduleActionRecord {
	var model models.ScheduleActionRecord
	model.Id = dto.Id
	model.JobName = dto.JobName
	model.Action = ToScheduleActionModel(dto.Action)
	model.Status = models.ScheduleActionRecordStatus(dto.Status)
	model.ScheduledAt = dto.ScheduledAt
	model.Created = dto.Created

	return model
}

func ToScheduleActionRecordModels(dtos []ScheduleActionRecord) []models.ScheduleActionRecord {
	scheduleActionRecordModels := make([]models.ScheduleActionRecord, len(dtos))
	for i, dto := range dtos {
		scheduleActionRecordModels[i] = ToScheduleActionRecordModel(dto)
	}
	return scheduleActionRecordModels
}

func FromScheduleActionRecordModelToDTO(model models.ScheduleActionRecord) ScheduleActionRecord {
	var dto ScheduleActionRecord
	dto.Id = model.Id
	dto.JobName = model.JobName
	dto.Action = FromScheduleActionModelToDTO(model.Action)
	dto.Status = string(model.Status)
	dto.ScheduledAt = model.ScheduledAt
	dto.Created = model.Created

	return dto
}

func FromScheduleActionRecordModelsToDTOs(records []models.ScheduleActionRecord) []ScheduleActionRecord {
	scheduleActionRecordDTOs := make([]ScheduleActionRecord, len(records))
	for i, record := range records {
		scheduleActionRecordDTOs[i] = FromScheduleActionRecordModelToDTO(record)
	}
	return scheduleActionRecordDTOs
}
