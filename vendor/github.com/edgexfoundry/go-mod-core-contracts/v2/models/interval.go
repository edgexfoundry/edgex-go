//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

// Interval and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-scheduler/2.x#/Interval
// Model fields are same as the DTOs documented by this swagger. Exceptions, if any, are noted below.
type Interval struct {
	DBTimestamp
	Id       string
	Name     string
	Start    string
	End      string
	Interval string
}
