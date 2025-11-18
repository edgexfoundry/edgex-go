//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "maps"

type ResourceOperation struct {
	DeviceResource string
	DefaultValue   string
	Mappings       map[string]string
}

func (r ResourceOperation) Clone() ResourceOperation {
	cloned := ResourceOperation{
		DeviceResource: r.DeviceResource,
		DefaultValue:   r.DefaultValue,
	}
	if len(r.Mappings) > 0 {
		cloned.Mappings = make(map[string]string)
		maps.Copy(cloned.Mappings, r.Mappings)
	}
	return cloned
}
