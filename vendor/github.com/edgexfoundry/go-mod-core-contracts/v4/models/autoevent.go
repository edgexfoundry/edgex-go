//
// Copyright (C) 2020-2021 IOTech Ltd
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type AutoEvent struct {
	Interval          string
	OnChange          bool
	OnChangeThreshold float64
	SourceName        string
	Retention         Retention
}

type Retention struct {
	MaxCap   int64
	MinCap   int64
	Duration string
}
