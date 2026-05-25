//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type SecretPoster interface {
	// AddSecret adds EdgeX Service exclusive secret to the Secret Store with specified baseUrl and secret data
	AddSecret(ctx context.Context, baseUrl string, request common.SecretRequest) (common.BaseResponse, errors.EdgeX)
}
