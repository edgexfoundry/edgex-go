//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type IntervalAction struct {
	DBTimestamp  `json:",inline"`
	Id           string  `json:"id,omitempty" validate:"omitempty,uuid"`
	Name         string  `json:"name" validate:"edgex-dto-none-empty-string"`
	IntervalName string  `json:"intervalName" validate:"edgex-dto-none-empty-string"`
	Address      Address `json:"address" validate:"required"`
	Content      string  `json:"content,omitempty"`
	ContentType  string  `json:"contentType,omitempty"`
	AdminState   string  `json:"adminState" validate:"oneof='LOCKED' 'UNLOCKED'"`
	AuthMethod   string  `json:"authMethod" validate:"oneof='' 'NONE' 'JWT'"`
}

// NewIntervalAction creates intervalAction DTO with required fields
func NewIntervalAction(name string, intervalName string, address Address) IntervalAction {
	return IntervalAction{
		Name:         name,
		IntervalName: intervalName,
		Address:      address,
		AdminState:   models.Unlocked,
	}
}

type UpdateIntervalAction struct {
	Id           *string  `json:"id" validate:"required_without=Name,edgex-dto-uuid"`
	Name         *string  `json:"name" validate:"required_without=Id,edgex-dto-none-empty-string"`
	IntervalName *string  `json:"intervalName" validate:"omitempty,edgex-dto-none-empty-string"`
	Content      *string  `json:"content"`
	ContentType  *string  `json:"contentType"`
	Address      *Address `json:"address"`
	AdminState   *string  `json:"adminState" validate:"omitempty,oneof='LOCKED' 'UNLOCKED'"`
	AuthMethod   *string  `json:"authMethod" validate:"omitempty,oneof='' 'NONE' 'JWT'"`
}

// NewUpdateIntervalAction creates updateIntervalAction DTO with required field
func NewUpdateIntervalAction(name string) UpdateIntervalAction {
	return UpdateIntervalAction{Name: &name}
}

// ToIntervalActionModel transforms the IntervalAction DTO to the IntervalAction Model
func ToIntervalActionModel(dto IntervalAction) models.IntervalAction {
	var model models.IntervalAction
	model.Id = dto.Id
	model.Name = dto.Name
	model.IntervalName = dto.IntervalName
	model.Content = dto.Content
	model.ContentType = dto.ContentType
	model.Address = ToAddressModel(dto.Address)
	model.AdminState = models.AdminState(dto.AdminState)
	model.AuthMethod = models.AuthMethod(dto.AuthMethod)
	return model
}

// FromIntervalActionModelToDTO transforms the IntervalAction Model to the IntervalAction DTO
func FromIntervalActionModelToDTO(model models.IntervalAction) IntervalAction {
	var dto IntervalAction
	dto.DBTimestamp = DBTimestamp(model.DBTimestamp)
	dto.Id = model.Id
	dto.Name = model.Name
	dto.IntervalName = model.IntervalName
	dto.Content = model.Content
	dto.ContentType = model.ContentType
	dto.Address = FromAddressModelToDTO(model.Address)
	dto.AdminState = string(model.AdminState)
	dto.AuthMethod = string(model.AuthMethod)
	return dto
}
