//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type DeviceInfo struct {
	Id           int
	DeviceName   string
	ProfileName  string
	SourceName   string
	Tags         map[string]any
	ResourceName string
	ValueType    string
	Units        string
	MediaType    string
	MarkDeleted  bool
}
