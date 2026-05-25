//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// KVSClient defines the interface for interactions with the kvs endpoint on the EdgeX core-keeper service.
type KVSClient interface {
	// UpdateValuesByKey updates values of the specified key and the child keys defined in the request payload.
	// If no key exists at the given path, the key(s) will be created.
	// If 'flatten' is true, the request json object would be flattened before storing into database.
	UpdateValuesByKey(ctx context.Context, key string, flatten bool, reqs requests.UpdateKeysRequest) (responses.KeysResponse, errors.EdgeX)
	// ValuesByKey returns the values of the specified key prefix.
	ValuesByKey(ctx context.Context, key string) (responses.MultiKeyValueResponse, errors.EdgeX)
	// ListKeys returns the list of the keys with the specified key prefix.
	ListKeys(ctx context.Context, key string) (responses.KeysResponse, errors.EdgeX)
	// DeleteKey deletes the specified key.
	DeleteKey(ctx context.Context, key string) (responses.KeysResponse, errors.EdgeX)
	// DeleteKeysByPrefix deletes all keys with the specified prefix.
	DeleteKeysByPrefix(ctx context.Context, key string) (responses.KeysResponse, errors.EdgeX)
}
