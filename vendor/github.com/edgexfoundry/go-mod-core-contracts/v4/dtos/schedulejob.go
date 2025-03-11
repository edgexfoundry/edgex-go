//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type ScheduleJob struct {
	DBTimestamp              `json:",inline"`
	Id                       string           `json:"id,omitempty" validate:"omitempty,uuid"`
	Name                     string           `json:"name" validate:"edgex-dto-none-empty-string"`
	Definition               ScheduleDef      `json:"definition" validate:"required"`
	AutoTriggerMissedRecords bool             `json:"autoTriggerMissedRecords,omitempty"`
	Actions                  []ScheduleAction `json:"actions" validate:"required,gt=0,dive"`
	AdminState               string           `json:"adminState" validate:"omitempty,oneof='LOCKED' 'UNLOCKED'"`
	Labels                   []string         `json:"labels,omitempty"`
	Properties               map[string]any   `json:"properties"`
}

type UpdateScheduleJob struct {
	Id                       *string          `json:"id" validate:"required_without=Name,edgex-dto-uuid"`
	Name                     *string          `json:"name" validate:"required_without=Id,edgex-dto-none-empty-string"`
	Definition               *ScheduleDef     `json:"definition" validate:"omitempty"`
	AutoTriggerMissedRecords *bool            `json:"autoTriggerMissedRecords,omitempty"`
	Actions                  []ScheduleAction `json:"actions,omitempty"`
	AdminState               *string          `json:"adminState" validate:"omitempty,oneof='LOCKED' 'UNLOCKED'"`
	Labels                   []string         `json:"labels,omitempty"`
	Properties               map[string]any   `json:"properties,omitempty"`
}

// Validate satisfies the Validator interface
func (s *ScheduleJob) Validate() error {
	err := common.Validate(s)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleJob.", err)
	}

	err = s.Definition.Validate()
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleDef.", err)
	}

	for _, action := range s.Actions {
		err = action.Validate()
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleAction.", err)
		}
	}

	return nil
}

type ScheduleDef struct {
	Type           string `json:"type" validate:"oneof='INTERVAL' 'CRON'"`
	StartTimestamp int64  `json:"startTimestamp,omitempty"`
	EndTimestamp   int64  `json:"endTimestamp,omitempty"`

	IntervalScheduleDef `json:",inline" validate:"-"`
	CronScheduleDef     `json:",inline" validate:"-"`
}

// Validate satisfies the Validator interface
func (s *ScheduleDef) Validate() error {
	err := common.Validate(s)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleDef.", err)
	}

	switch s.Type {
	case common.DefInterval:
		err = common.Validate(s.IntervalScheduleDef)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid IntervalScheduleDef.", err)
		}
	case common.DefCron:
		err = common.Validate(s.CronScheduleDef)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid CronScheduleDef.", err)
		}
	}

	if s.EndTimestamp != 0 {
		if s.EndTimestamp < s.StartTimestamp {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "endTimestamp must be greater than startTimestamp", nil)
		}
	}

	return nil
}

type IntervalScheduleDef struct {
	Interval string `json:"interval,omitempty" validate:"required,edgex-dto-duration"`
}

type CronScheduleDef struct {
	Crontab string `json:"crontab,omitempty" validate:"required"`
}

type ScheduleAction struct {
	Type        string `json:"type" validate:"oneof='EDGEXMESSAGEBUS' 'REST' 'DEVICECONTROL'"`
	ContentType string `json:"contentType,omitempty"`
	Payload     []byte `json:"payload,omitempty"`

	EdgeXMessageBusAction `json:",inline" validate:"-"`
	RESTAction            `json:",inline" validate:"-"`
	DeviceControlAction   `json:",inline" validate:"-"`
}

func (s *ScheduleAction) UnmarshalJSON(b []byte) error {
	type Alias ScheduleAction
	alias := &struct {
		Payload any `json:"payload,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal ScheduleAction as JSON.", err)
	}

	if alias.Payload == nil {
		return nil
	}

	switch v := alias.Payload.(type) {
	case string:
		// Check if payload is a base64 encoded string
		if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
			s.Payload = decoded
		} else {
			// Or just a plain string
			s.Payload = []byte(v)
		}
	case map[string]any:
		// If payload is a JSON object then marshal it
		if encoded, err := json.Marshal(v); err == nil {
			s.Payload = encoded
		} else {
			return err
		}
	default:
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("Failed to unmarshal ScheduleAction, unsupported payload type: %s.", v), nil)
	}

	return nil
}

func (s *ScheduleAction) Validate() error {
	err := common.Validate(s)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ScheduleAction.", err)
	}

	switch s.Type {
	case common.ActionEdgeXMessageBus:
		err = common.Validate(s.EdgeXMessageBusAction)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid EdgeXMessageBusAction.", err)
		}
	case common.ActionREST:
		err = common.Validate(s.RESTAction)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid RESTAction.", err)
		}
	case common.ActionDeviceControl:
		err = common.Validate(s.DeviceControlAction)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid DeviceControlAction.", err)
		}
	}

	return nil
}

type EdgeXMessageBusAction struct {
	Topic string `json:"topic,omitempty" validate:"required"`
}

type RESTAction struct {
	Address         string `json:"address,omitempty" validate:"required"`
	Method          string `json:"method,omitempty" validate:"required"`
	InjectEdgeXAuth bool   `json:"injectEdgeXAuth,omitempty"`
}

type DeviceControlAction struct {
	DeviceName string `json:"deviceName,omitempty" validate:"required"`
	SourceName string `json:"sourceName,omitempty" validate:"required"`
}

func ToScheduleJobModel(dto ScheduleJob) models.ScheduleJob {
	var model models.ScheduleJob
	model.Id = dto.Id
	model.Name = dto.Name
	model.Definition = ToScheduleDefModel(dto.Definition)
	model.AutoTriggerMissedRecords = dto.AutoTriggerMissedRecords
	model.Actions = ToScheduleActionModels(dto.Actions)
	model.AdminState = models.AssignAdminState(dto.AdminState)
	model.Labels = dto.Labels
	model.Properties = dto.Properties

	if model.Properties == nil {
		model.Properties = make(map[string]any)
	}

	return model
}

func FromScheduleJobModelToDTO(model models.ScheduleJob) ScheduleJob {
	var dto ScheduleJob
	dto.DBTimestamp = DBTimestamp(model.DBTimestamp)
	dto.Id = model.Id
	dto.Name = model.Name
	dto.Definition = FromScheduleDefModelToDTO(model.Definition)
	dto.AutoTriggerMissedRecords = model.AutoTriggerMissedRecords
	dto.Actions = FromScheduleActionModelsToDTOs(model.Actions)
	dto.AdminState = string(model.AdminState)
	dto.Labels = model.Labels
	dto.Properties = model.Properties

	if dto.Properties == nil {
		dto.Properties = make(map[string]any)
	}

	return dto
}

func ToScheduleDefModel(dto ScheduleDef) models.ScheduleDef {
	var model models.ScheduleDef

	switch dto.Type {
	case common.DefInterval:
		model = models.IntervalScheduleDef{
			BaseScheduleDef: models.BaseScheduleDef{
				Type:           common.DefInterval,
				StartTimestamp: dto.StartTimestamp,
				EndTimestamp:   dto.EndTimestamp,
			},
			Interval: dto.Interval,
		}
	case common.DefCron:
		model = models.CronScheduleDef{
			BaseScheduleDef: models.BaseScheduleDef{
				Type:           common.DefCron,
				StartTimestamp: dto.StartTimestamp,
				EndTimestamp:   dto.EndTimestamp,
			},
			Crontab: dto.Crontab,
		}
	}

	return model
}

func FromScheduleDefModelToDTO(model models.ScheduleDef) ScheduleDef {
	var dto ScheduleDef

	switch model.GetBaseScheduleDef().Type {
	case common.DefInterval:
		durationModel := model.(models.IntervalScheduleDef)
		dto = ScheduleDef{
			Type:                common.DefInterval,
			StartTimestamp:      durationModel.StartTimestamp,
			EndTimestamp:        durationModel.EndTimestamp,
			IntervalScheduleDef: IntervalScheduleDef{Interval: durationModel.Interval},
		}
	case common.DefCron:
		cronModel := model.(models.CronScheduleDef)
		dto = ScheduleDef{
			Type:            common.DefCron,
			StartTimestamp:  cronModel.StartTimestamp,
			EndTimestamp:    cronModel.EndTimestamp,
			CronScheduleDef: CronScheduleDef{Crontab: cronModel.Crontab},
		}
	}

	return dto
}

func ToScheduleActionModel(dto ScheduleAction) models.ScheduleAction {
	var model models.ScheduleAction

	switch dto.Type {
	case common.ActionEdgeXMessageBus:
		model = models.EdgeXMessageBusAction{
			BaseScheduleAction: models.BaseScheduleAction{
				Type:        common.ActionEdgeXMessageBus,
				ContentType: dto.ContentType,
				Payload:     dto.Payload,
			},
			Topic: dto.Topic,
		}
	case common.ActionREST:
		model = models.RESTAction{
			BaseScheduleAction: models.BaseScheduleAction{
				Type:        common.ActionREST,
				ContentType: dto.ContentType,
				Payload:     dto.Payload,
			},
			Method:          dto.Method,
			Address:         dto.Address,
			InjectEdgeXAuth: dto.InjectEdgeXAuth,
		}
	case common.ActionDeviceControl:
		model = models.DeviceControlAction{
			BaseScheduleAction: models.BaseScheduleAction{
				Type:        common.ActionDeviceControl,
				ContentType: dto.ContentType,
				Payload:     dto.Payload,
			},
			DeviceName: dto.DeviceName,
			SourceName: dto.SourceName,
		}
	}

	return model
}

func FromScheduleActionModelToDTO(model models.ScheduleAction) ScheduleAction {
	var dto ScheduleAction

	switch model.GetBaseScheduleAction().Type {
	case common.ActionEdgeXMessageBus:
		messageBusModel := model.(models.EdgeXMessageBusAction)
		dto = ScheduleAction{
			Type:        common.ActionEdgeXMessageBus,
			ContentType: messageBusModel.ContentType,
			Payload:     messageBusModel.Payload,
			EdgeXMessageBusAction: EdgeXMessageBusAction{
				Topic: messageBusModel.Topic,
			},
		}
	case common.ActionREST:
		restModel := model.(models.RESTAction)
		dto = ScheduleAction{
			Type:        common.ActionREST,
			ContentType: restModel.ContentType,
			Payload:     restModel.Payload,
			RESTAction: RESTAction{
				Address:         restModel.Address,
				Method:          restModel.Method,
				InjectEdgeXAuth: restModel.InjectEdgeXAuth,
			},
		}
	case common.ActionDeviceControl:
		deviceControlModel := model.(models.DeviceControlAction)
		dto = ScheduleAction{
			Type:        common.ActionDeviceControl,
			ContentType: deviceControlModel.ContentType,
			Payload:     deviceControlModel.Payload,
			DeviceControlAction: DeviceControlAction{
				DeviceName: deviceControlModel.DeviceName,
				SourceName: deviceControlModel.SourceName,
			},
		}
	}

	return dto
}

func ToScheduleActionModels(dtos []ScheduleAction) []models.ScheduleAction {
	models := make([]models.ScheduleAction, len(dtos))
	for i, dto := range dtos {
		models[i] = ToScheduleActionModel(dto)
	}
	return models
}

func FromScheduleActionModelsToDTOs(models []models.ScheduleAction) []ScheduleAction {
	dtos := make([]ScheduleAction, len(models))
	for i, model := range models {
		dtos[i] = FromScheduleActionModelToDTO(model)
	}
	return dtos
}
