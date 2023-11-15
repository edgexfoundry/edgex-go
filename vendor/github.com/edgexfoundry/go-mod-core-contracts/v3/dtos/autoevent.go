//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type AutoEvent struct {
	Interval   string `json:"interval" yaml:"interval" validate:"required,edgex-dto-duration"`
	OnChange   bool   `json:"onChange" yaml:"onChange"`
	SourceName string `json:"sourceName" yaml:"sourceName" validate:"required"`
}

// ToAutoEventModel transforms the AutoEvent DTO to the AutoEvent model
func ToAutoEventModel(a AutoEvent) models.AutoEvent {
	return models.AutoEvent{
		Interval:   a.Interval,
		OnChange:   a.OnChange,
		SourceName: a.SourceName,
	}
}

// ToAutoEventModels transforms the AutoEvent DTO array to the AutoEvent model array
func ToAutoEventModels(autoEventDTOs []AutoEvent) []models.AutoEvent {
	autoEventModels := make([]models.AutoEvent, len(autoEventDTOs))
	for i, a := range autoEventDTOs {
		autoEventModels[i] = ToAutoEventModel(a)
	}
	return autoEventModels
}

// FromAutoEventModelToDTO transforms the AutoEvent model to the AutoEvent DTO
func FromAutoEventModelToDTO(a models.AutoEvent) AutoEvent {
	return AutoEvent{
		Interval:   a.Interval,
		OnChange:   a.OnChange,
		SourceName: a.SourceName,
	}
}

// FromAutoEventModelsToDTOs transforms the AutoEvent model array to the AutoEvent DTO array
func FromAutoEventModelsToDTOs(autoEvents []models.AutoEvent) []AutoEvent {
	dtos := make([]AutoEvent, len(autoEvents))
	for i, a := range autoEvents {
		dtos[i] = FromAutoEventModelToDTO(a)
	}
	return dtos
}
