//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// SecretStoreTokenClient defines the interface for interactions with the API endpoint on the security-secretstore-setup service.
type SecretStoreTokenClient interface {
	// RegenToken regenerates the secret store client token based on the specified entity id
	RegenToken(ctx context.Context, entityId string) (dtoCommon.BaseResponse, errors.EdgeX)
}
