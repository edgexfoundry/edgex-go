//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
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

// SubscriptionByName queries subscription by name
func SubscriptionByName(name string, dic *di.Container) (subscription dtos.Subscription, err errors.EdgeX) {
	if name == "" {
		return subscription, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	subscriptionModel, err := dbClient.SubscriptionByName(name)
	if err != nil {
		return subscription, errors.NewCommonEdgeXWrapper(err)
	}
	subscription = dtos.FromSubscriptionModelToDTO(subscriptionModel)
	return subscription, nil
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

// DeleteSubscriptionByName deletes the subscription by name
func DeleteSubscriptionByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	err := dbClient.DeleteSubscriptionByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// PatchSubscription executes the PATCH operation with the subscription DTO to replace the old data
func PatchSubscription(ctx context.Context, dto dtos.UpdateSubscription, dic *di.Container) errors.EdgeX {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	var subscription models.Subscription
	var edgexErr errors.EdgeX
	if dto.Id != nil {
		subscription, edgexErr = dbClient.SubscriptionById(*dto.Id)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
	} else {
		subscription, edgexErr = dbClient.SubscriptionByName(*dto.Name)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
	}
	if dto.Name != nil && *dto.Name != subscription.Name {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("subscription name '%s' not match the existing '%s' ", *dto.Name, subscription.Name), nil)
	}

	requests.ReplaceSubscriptionModelFieldsWithDTO(&subscription, dto)
	_, edgeXerr := dbClient.SubscriptionByName(subscription.Name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("subscription '%s' existence check failed", subscription.Name), edgeXerr)
	}

	edgexErr = dbClient.DeleteSubscriptionByName(subscription.Name)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	_, edgexErr = dbClient.AddSubscription(subscription)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	lc.Debugf("Subscription patched on DB successfully. Correlation-ID: %s ", correlation.FromContext(ctx))
	return nil
}
