//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type secretPoster struct {
	authInjector interfaces.AuthenticationInjector
}

// NewSecretPoster creates an instance of SecretPoster
func NewSecretPoster(authInjector interfaces.AuthenticationInjector) interfaces.SecretPoster {
	return &secretPoster{
		authInjector: authInjector,
	}
}

func (sp *secretPoster) AddSecret(ctx context.Context, baseUrl string, request dtoCommon.SecretRequest) (res dtoCommon.BaseResponse, err errors.EdgeX) {
	err = utils.PostRequestWithRawData(ctx, &res, baseUrl, common.ApiSecretRoute, nil, request, sp.authInjector)
	if err != nil {
		return res, errors.NewCommonEdgeXWrapper(err)
	}
	return res, nil
}
