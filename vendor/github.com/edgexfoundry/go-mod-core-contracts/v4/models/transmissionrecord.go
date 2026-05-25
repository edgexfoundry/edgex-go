//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type TransmissionRecord struct {
	Status   TransmissionStatus
	Response string
	Sent     int64
}
