//
// Copyright (C) 2021-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
)

func CheckPayloadSize(payload []byte, sizeLimit int64) errors.EdgeX {
	// Treat 0 as unlimit size
	if sizeLimit < 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("sizeLimit cannot be lower than 0, current sizeLimit: %d", sizeLimit), nil)
	} else if sizeLimit != 0 && int64(len(payload)) > sizeLimit {
		return errors.NewCommonEdgeX(errors.KindLimitExceeded, fmt.Sprintf("request size exceed %d KB", sizeLimit), nil)
	}
	return nil
}
