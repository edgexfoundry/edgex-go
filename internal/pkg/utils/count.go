//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// CheckCountRange evaluates if the offset and limit parameters are within the valid range.
func CheckCountRange(totalCount int64, offset, limit int) (continueExec bool, err errors.EdgeX) {
	if limit == 0 || totalCount == 0 {
		return false, nil
	}
	if int64(offset) > totalCount {
		return false, errors.NewCommonEdgeX(errors.KindRangeNotSatisfiable, fmt.Sprintf("query objects bounds out of range. length:%v offset:%v", totalCount, offset), nil)
	}

	return true, nil
}
