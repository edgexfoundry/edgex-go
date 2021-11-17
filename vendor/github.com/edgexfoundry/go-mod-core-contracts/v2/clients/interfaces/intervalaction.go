//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// IntervalActionClient defines the interface for interactions with the IntervalAction endpoint on the EdgeX Foundry support-scheduler service.
type IntervalActionClient interface {
	// Add adds new intervalActions.
	Add(ctx context.Context, reqs []requests.AddIntervalActionRequest) ([]common.BaseWithIdResponse, errors.EdgeX)
	// Update updates intervalActions.
	Update(ctx context.Context, reqs []requests.UpdateIntervalActionRequest) ([]common.BaseResponse, errors.EdgeX)
	// AllIntervalActions returns all intervalActions.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	AllIntervalActions(ctx context.Context, offset int, limit int) (responses.MultiIntervalActionsResponse, errors.EdgeX)
	// IntervalActionByName returns a intervalAction by name.
	IntervalActionByName(ctx context.Context, name string) (responses.IntervalActionResponse, errors.EdgeX)
	// DeleteIntervalActionByName deletes a intervalAction by name.
	DeleteIntervalActionByName(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
}
