//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import "github.com/edgexfoundry/go-mod-core-contracts/v4/common"

func SigningKeyName(issuer string) string {
	return issuer + "/" + common.SigningKeyType
}

func VerificationKeyName(issuer string) string {
	return issuer + "/" + common.VerificationKeyType
}
