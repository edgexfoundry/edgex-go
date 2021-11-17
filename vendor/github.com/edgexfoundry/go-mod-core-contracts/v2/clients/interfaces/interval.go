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

// IntervalClient defines the interface for interactions with the Interval endpoint on the EdgeX Foundry support-scheduler service.
type IntervalClient interface {
	// Add adds new intervals.
	Add(ctx context.Context, reqs []requests.AddIntervalRequest) ([]common.BaseWithIdResponse, errors.EdgeX)
	// Update updates intervals.
	Update(ctx context.Context, reqs []requests.UpdateIntervalRequest) ([]common.BaseResponse, errors.EdgeX)
	// AllIntervals returns all intervals.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	AllIntervals(ctx context.Context, offset int, limit int) (responses.MultiIntervalsResponse, errors.EdgeX)
	// IntervalByName returns a interval by name.
	IntervalByName(ctx context.Context, name string) (responses.IntervalResponse, errors.EdgeX)
	// DeleteIntervalByName deletes a interval by name.
	DeleteIntervalByName(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
}
