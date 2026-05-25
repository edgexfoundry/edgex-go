//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "maps"

type DeviceResource struct {
	Description string
	Name        string
	IsHidden    bool
	Properties  ResourceProperties
	Attributes  map[string]interface{}
	Tags        map[string]any
}

func (dr DeviceResource) Clone() DeviceResource {
	cloned := DeviceResource{
		Description: dr.Description,
		Name:        dr.Name,
		IsHidden:    dr.IsHidden,
		Properties:  dr.Properties.Clone(),
	}
	if len(dr.Attributes) > 0 {
		cloned.Attributes = make(map[string]any)
		maps.Copy(cloned.Attributes, dr.Attributes)
	}
	if len(dr.Tags) > 0 {
		cloned.Tags = make(map[string]any)
		maps.Copy(cloned.Tags, dr.Tags)
	}
	return cloned
}
