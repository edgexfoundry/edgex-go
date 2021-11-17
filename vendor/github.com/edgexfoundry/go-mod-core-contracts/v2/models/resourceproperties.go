//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

// ResourceProperties and its properties care defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.x#/ResourceProperties
// Model fields are same as the DTOs documented by this swagger. Exceptions, if any, are noted below.
type ResourceProperties struct {
	ValueType    string
	ReadWrite    string
	Units        string
	Minimum      string
	Maximum      string
	DefaultValue string
	Mask         string
	Shift        string
	Scale        string
	Offset       string
	Base         string
	Assertion    string
	MediaType    string
}
