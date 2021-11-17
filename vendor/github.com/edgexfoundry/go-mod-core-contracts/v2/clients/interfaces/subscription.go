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

// SubscriptionClient defines the interface for interactions with the Subscription endpoint on the EdgeX Foundry support-notifications service.
type SubscriptionClient interface {
	// Add adds new subscriptions.
	Add(ctx context.Context, reqs []requests.AddSubscriptionRequest) ([]common.BaseWithIdResponse, errors.EdgeX)
	// Update updates subscriptions.
	Update(ctx context.Context, reqs []requests.UpdateSubscriptionRequest) ([]common.BaseResponse, errors.EdgeX)
	// AllSubscriptions queries subscriptions with offset and limit
	AllSubscriptions(ctx context.Context, offset int, limit int) (responses.MultiSubscriptionsResponse, errors.EdgeX)
	// SubscriptionsByCategory queries subscriptions with category, offset and limit
	SubscriptionsByCategory(ctx context.Context, category string, offset int, limit int) (responses.MultiSubscriptionsResponse, errors.EdgeX)
	// SubscriptionsByLabel queries subscriptions with label, offset and limit
	SubscriptionsByLabel(ctx context.Context, label string, offset int, limit int) (responses.MultiSubscriptionsResponse, errors.EdgeX)
	// SubscriptionsByReceiver queries subscriptions with receiver, offset and limit
	SubscriptionsByReceiver(ctx context.Context, receiver string, offset int, limit int) (responses.MultiSubscriptionsResponse, errors.EdgeX)
	// SubscriptionByName query subscription by name.
	SubscriptionByName(ctx context.Context, name string) (responses.SubscriptionResponse, errors.EdgeX)
	// DeleteSubscriptionByName deletes a subscription by name.
	DeleteSubscriptionByName(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
}
