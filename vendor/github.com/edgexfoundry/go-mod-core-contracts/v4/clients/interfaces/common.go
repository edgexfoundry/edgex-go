//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type CommonClient interface {
	// Configuration obtains configuration information from the target service.
	Configuration(ctx context.Context) (common.ConfigResponse, errors.EdgeX)
	// Ping tests whether the service is working
	Ping(ctx context.Context) (common.PingResponse, errors.EdgeX)
	// Version obtains version information from the target service.
	Version(ctx context.Context) (common.VersionResponse, errors.EdgeX)
	// AddSecret adds EdgeX Service exclusive secret to the Secret Store
	AddSecret(ctx context.Context, request common.SecretRequest) (common.BaseResponse, errors.EdgeX)
}
