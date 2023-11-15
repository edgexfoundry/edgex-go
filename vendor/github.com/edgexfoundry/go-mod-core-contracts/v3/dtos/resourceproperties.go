//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type ResourceProperties struct {
	ValueType    string         `json:"valueType" yaml:"valueType" validate:"required,edgex-dto-value-type"`
	ReadWrite    string         `json:"readWrite" yaml:"readWrite" validate:"required,oneof='R' 'W' 'RW' 'WR'"`
	Units        string         `json:"units,omitempty" yaml:"units"`
	Minimum      *float64       `json:"minimum,omitempty" yaml:"minimum"`
	Maximum      *float64       `json:"maximum,omitempty" yaml:"maximum"`
	DefaultValue string         `json:"defaultValue,omitempty" yaml:"defaultValue"`
	Mask         *uint64        `json:"mask,omitempty" yaml:"mask"`
	Shift        *int64         `json:"shift,omitempty" yaml:"shift"`
	Scale        *float64       `json:"scale,omitempty" yaml:"scale"`
	Offset       *float64       `json:"offset,omitempty" yaml:"offset"`
	Base         *float64       `json:"base,omitempty" yaml:"base"`
	Assertion    string         `json:"assertion,omitempty" yaml:"assertion"`
	MediaType    string         `json:"mediaType,omitempty" yaml:"mediaType"`
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
