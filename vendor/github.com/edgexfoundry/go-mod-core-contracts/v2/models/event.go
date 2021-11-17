//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

// Event and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.x#/Event
// Model fields are same as the DTOs documented by this swagger. Exceptions, if any, are noted below.
type Event struct {
	Id          string
	DeviceName  string
	ProfileName string
	SourceName  string
	Origin      int64
	Readings    []Reading
	Tags        map[string]interface{}
}
