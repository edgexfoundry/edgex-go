//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

// KeyData contains the signing or verification key for the JWT token
type KeyData struct {
	Issuer string
	Type   string
	Key    string
}
