//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

type DBClient interface {
	CloseSession()

	AddSubscription(e models.Subscription) (models.Subscription, errors.EdgeX)
	SubscriptionById(id string) (models.Subscription, errors.EdgeX)
	AllSubscriptions(offset int, limit int) ([]models.Subscription, errors.EdgeX)
	SubscriptionByName(name string) (models.Subscription, errors.EdgeX)
	SubscriptionsByCategory(offset, limit int, category string) ([]models.Subscription, errors.EdgeX)
	SubscriptionsByLabel(offset, limit int, label string) ([]models.Subscription, errors.EdgeX)
	SubscriptionsByReceiver(offset, limit int, receiver string) ([]models.Subscription, errors.EdgeX)
	DeleteSubscriptionByName(name string) errors.EdgeX
}
