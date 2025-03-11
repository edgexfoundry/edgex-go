//
// Copyright (C) 2020-2023 IOTech Ltd
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type AutoEvent struct {
	Interval          string    `json:"interval" yaml:"interval" validate:"required,edgex-dto-duration=1ms"` // min/max can be defined as params, ex. edgex-dto-duration=10ms0x2C24h
	OnChange          bool      `json:"onChange" yaml:"onChange"`
	OnChangeThreshold float64   `json:"onChangeThreshold" yaml:"onChangeThreshold" validate:"gte=0"`
	SourceName        string    `json:"sourceName" yaml:"sourceName" validate:"required"`
	Retention         Retention `json:"retention" yaml:"retention" validate:"omitempty"`
}

type Retention struct {
	MaxCap   int64  `json:"maxCap" yaml:"maxCap"`
	MinCap   int64  `json:"minCap" yaml:"minCap"`
	Duration string `json:"duration" yaml:"duration" validate:"omitempty,edgex-dto-duration=0s"`
}

// ToAutoEventModel transforms the AutoEvent DTO to the AutoEvent model
func ToAutoEventModel(a AutoEvent) models.AutoEvent {
	return models.AutoEvent{
		Interval:          a.Interval,
		OnChange:          a.OnChange,
		OnChangeThreshold: a.OnChangeThreshold,
		SourceName:        a.SourceName,
		Retention: models.Retention{
			MaxCap:   a.Retention.MaxCap,
			MinCap:   a.Retention.MinCap,
			Duration: a.Retention.Duration,
		},
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
		Interval:          a.Interval,
		OnChange:          a.OnChange,
		OnChangeThreshold: a.OnChangeThreshold,
		SourceName:        a.SourceName,
		Retention: Retention{
			MaxCap:   a.Retention.MaxCap,
			MinCap:   a.Retention.MinCap,
			Duration: a.Retention.Duration,
		},
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
