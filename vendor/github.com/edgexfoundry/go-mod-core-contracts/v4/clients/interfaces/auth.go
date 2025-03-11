//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// AuthClient defines the interface for interactions with the auth API endpoint on the security-proxy-auth service.
type AuthClient interface {
	// AddKey adds the JWT signing or verification key
	AddKey(ctx context.Context, req requests.AddKeyDataRequest) (common.BaseResponse, errors.EdgeX)
	// VerificationKeyByIssuer returns the JWT verification key by the specified issuer
	VerificationKeyByIssuer(ctx context.Context, issuer string) (res responses.KeyDataResponse, err errors.EdgeX)
}
