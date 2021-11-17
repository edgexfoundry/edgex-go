//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

type DBTimestamp struct {
	Created  int64 `json:"created,omitempty"`
	Modified int64 `json:"modified,omitempty"`
}
