//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type Interval struct {
	DBTimestamp
	Id       string
	Name     string
	Start    string
	End      string
	Interval string
}
