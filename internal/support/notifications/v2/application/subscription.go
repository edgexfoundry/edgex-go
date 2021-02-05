//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// The AddSubscription function accepts the new Subscription model from the controller function
// and then invokes AddSubscription function of infrastructure layer to add new Subscription
func AddSubscription(d models.Subscription, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	addedSubscription, err := dbClient.AddSubscription(d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Subscription created on DB successfully. Subscription ID: %s, Correlation-ID: %s ",
		addedSubscription.Id,
		correlation.FromContext(ctx))

	return addedSubscription.Id, nil
}

// AllSubscriptions queries subscriptions by offset and limit
func AllSubscriptions(offset, limit int, dic *di.Container) (subscriptions []dtos.Subscription, err errors.EdgeX) {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	subs, err := dbClient.AllSubscriptions(offset, limit)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}
	subscriptions = make([]dtos.Subscription, len(subs))
	for i, sub := range subs {
		subscriptions[i] = dtos.FromSubscriptionModelToDTO(sub)
	}
	return subscriptions, nil
}

// SubscriptionsByCategory queries subscriptions with offset, limit, and category
func SubscriptionsByCategory(offset, limit int, category string, dic *di.Container) (subscriptions []dtos.Subscription, err errors.EdgeX) {
	if category == "" {
		return subscriptions, errors.NewCommonEdgeX(errors.KindContractInvalid, "category is empty", nil)
	}
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	subscriptionModels, err := dbClient.SubscriptionsByCategory(offset, limit, category)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}
	subscriptions = make([]dtos.Subscription, len(subscriptionModels))
	for i, s := range subscriptionModels {
		subscriptions[i] = dtos.FromSubscriptionModelToDTO(s)
	}
	return subscriptions, nil
}

// SubscriptionsByLabel queries subscriptions with offset, limit, and label
func SubscriptionsByLabel(offset, limit int, label string, dic *di.Container) (subscriptions []dtos.Subscription, err errors.EdgeX) {
	if label == "" {
		return subscriptions, errors.NewCommonEdgeX(errors.KindContractInvalid, "label is empty", nil)
	}
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	subscriptionModels, err := dbClient.SubscriptionsByLabel(offset, limit, label)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}
	subscriptions = make([]dtos.Subscription, len(subscriptionModels))
	for i, s := range subscriptionModels {
		subscriptions[i] = dtos.FromSubscriptionModelToDTO(s)
	}
	return subscriptions, nil
}

// SubscriptionsByReceiver queries subscriptions with offset, limit, and receiver
func SubscriptionsByReceiver(offset, limit int, receiver string, dic *di.Container) (subscriptions []dtos.Subscription, err errors.EdgeX) {
	if receiver == "" {
		return subscriptions, errors.NewCommonEdgeX(errors.KindContractInvalid, "receiver is empty", nil)
	}
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	subscriptionModels, err := dbClient.SubscriptionsByReceiver(offset, limit, receiver)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}
	subscriptions = make([]dtos.Subscription, len(subscriptionModels))
	for i, s := range subscriptionModels {
		subscriptions[i] = dtos.FromSubscriptionModelToDTO(s)
	}
	return subscriptions, nil
}
