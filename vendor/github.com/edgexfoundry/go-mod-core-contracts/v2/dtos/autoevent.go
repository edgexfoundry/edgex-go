//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// AutoEvent and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/AutoEvent
type AutoEvent struct {
	Interval   string `json:"interval" validate:"required,edgex-dto-duration"`
	OnChange   bool   `json:"onChange"`
	SourceName string `json:"sourceName" validate:"required"`
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
