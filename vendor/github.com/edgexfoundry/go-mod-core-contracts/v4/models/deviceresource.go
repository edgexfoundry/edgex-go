//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type DeviceResource struct {
	Description string
	Name        string
	IsHidden    bool
	Properties  ResourceProperties
	Attributes  map[string]interface{}
	Tags        map[string]any
}
