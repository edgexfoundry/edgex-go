//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type ResourceProperties struct {
	ValueType    string         `json:"valueType" yaml:"valueType" validate:"required,edgex-dto-value-type"`
	ReadWrite    string         `json:"readWrite" yaml:"readWrite" validate:"required,oneof='R' 'W' 'RW' 'WR'"`
	Units        string         `json:"units,omitempty" yaml:"units,omitempty"`
	Minimum      *float64       `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum      *float64       `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	DefaultValue string         `json:"defaultValue,omitempty" yaml:"defaultValue,omitempty"`
	Mask         *uint64        `json:"mask,omitempty" yaml:"mask,omitempty"`
	Shift        *int64         `json:"shift,omitempty" yaml:"shift,omitempty"`
	Scale        *float64       `json:"scale,omitempty" yaml:"scale,omitempty"`
	Offset       *float64       `json:"offset,omitempty" yaml:"offset,omitempty"`
	Base         *float64       `json:"base,omitempty" yaml:"base,omitempty"`
	Assertion    string         `json:"assertion,omitempty" yaml:"assertion,omitempty"`
	MediaType    string         `json:"mediaType,omitempty" yaml:"mediaType,omitempty"`
	Optional     map[string]any `json:"optional,omitempty" yaml:"optional"`
}

// ToResourcePropertiesModel transforms the ResourceProperties DTO to the ResourceProperties model
func ToResourcePropertiesModel(p ResourceProperties) models.ResourceProperties {
	return models.ResourceProperties{
		ValueType:    p.ValueType,
		ReadWrite:    p.ReadWrite,
		Units:        p.Units,
		Minimum:      p.Minimum,
		Maximum:      p.Maximum,
		DefaultValue: p.DefaultValue,
		Mask:         p.Mask,
		Shift:        p.Shift,
		Scale:        p.Scale,
		Offset:       p.Offset,
		Base:         p.Base,
		Assertion:    p.Assertion,
		MediaType:    p.MediaType,
		Optional:     p.Optional,
	}
}

// FromResourcePropertiesModelToDTO transforms the ResourceProperties Model to the ResourceProperties DTO
func FromResourcePropertiesModelToDTO(p models.ResourceProperties) ResourceProperties {
	return ResourceProperties{
		ValueType:    p.ValueType,
		ReadWrite:    p.ReadWrite,
		Units:        p.Units,
		Minimum:      p.Minimum,
		Maximum:      p.Maximum,
		DefaultValue: p.DefaultValue,
		Mask:         p.Mask,
		Shift:        p.Shift,
		Scale:        p.Scale,
		Offset:       p.Offset,
		Base:         p.Base,
		Assertion:    p.Assertion,
		MediaType:    p.MediaType,
		Optional:     p.Optional,
	}
}
