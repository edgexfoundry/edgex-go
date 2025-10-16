//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/edgexfoundry/edgex-go/internal/core/keeper/constants"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// ValidateKeys validates if the key contains invalid characters
func ValidateKeys(key string) errors.EdgeX {
	if !constants.KeyAllowedCharsRegex.MatchString(key) {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "key only allows characters between a to z, A to Z, 0 to 9, or one of -_ ~;=./%", nil)
	}
	return nil
}
