//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type ResourceProperties struct {
	ValueType    string
	ReadWrite    string
	Units        string
	Minimum      *float64
	Maximum      *float64
	DefaultValue string
	Mask         *uint64
	Shift        *int64
	Scale        *float64
	Offset       *float64
	Base         *float64
	Assertion    string
	MediaType    string
	Optional     map[string]any
}
