//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type Event struct {
	Id          string
	DeviceName  string
	ProfileName string
	SourceName  string
	Origin      int64
	Readings    []Reading
	Tags        map[string]interface{}
}
