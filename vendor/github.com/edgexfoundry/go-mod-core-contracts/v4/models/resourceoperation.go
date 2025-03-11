//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type ResourceOperation struct {
	DeviceResource string
	DefaultValue   string
	Mappings       map[string]string
}
