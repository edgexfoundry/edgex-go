//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/infrastructure/interfaces"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// The AddSubscription function accepts the new Subscription model from the controller function
// and then invokes AddSubscription function of infrastructure layer to add new Subscription
func AddSubscription(d models.Subscription, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

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
func AllSubscriptions(offset, limit int, dic *di.Container) (subscriptions []dtos.Subscription, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	subs, err := dbClient.AllSubscriptions(offset, limit)
	if err == nil {
		totalCount, err = dbClient.SubscriptionTotalCount()
	}
	if err != nil {
		return subscriptions, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	subscriptions = make([]dtos.Subscription, len(subs))
	for i, sub := range subs {
		subscriptions[i] = dtos.FromSubscriptionModelToDTO(sub)
	}
	return subscriptions, totalCount, nil
}

// SubscriptionByName queries subscription by name
func SubscriptionByName(name string, dic *di.Container) (subscription dtos.Subscription, err errors.EdgeX) {
	if name == "" {
		return subscription, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	subscriptionModel, err := dbClient.SubscriptionByName(name)
	if err != nil {
		return subscription, errors.NewCommonEdgeXWrapper(err)
	}
	subscription = dtos.FromSubscriptionModelToDTO(subscriptionModel)
	return subscription, nil
}

// SubscriptionsByCategory queries subscriptions with offset, limit, and category
func SubscriptionsByCategory(offset, limit int, category string, dic *di.Container) (subscriptions []dtos.Subscription, totalCount uint32, err errors.EdgeX) {
	if category == "" {
		return subscriptions, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "category is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	subscriptionModels, err := dbClient.SubscriptionsByCategory(offset, limit, category)
	if err == nil {
		totalCount, err = dbClient.SubscriptionCountByCategory(category)
	}
	if err != nil {
		return subscriptions, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	subscriptions = make([]dtos.Subscription, len(subscriptionModels))
	for i, s := range subscriptionModels {
		subscriptions[i] = dtos.FromSubscriptionModelToDTO(s)
	}
	return subscriptions, totalCount, nil
}

// SubscriptionsByLabel queries subscriptions with offset, limit, and label
func SubscriptionsByLabel(offset, limit int, label string, dic *di.Container) (subscriptions []dtos.Subscription, totalCount uint32, err errors.EdgeX) {
	if label == "" {
		return subscriptions, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "label is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	subscriptionModels, err := dbClient.SubscriptionsByLabel(offset, limit, label)
	if err == nil {
		totalCount, err = dbClient.SubscriptionCountByLabel(label)
	}
	if err != nil {
		return subscriptions, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	subscriptions = make([]dtos.Subscription, len(subscriptionModels))
	for i, s := range subscriptionModels {
		subscriptions[i] = dtos.FromSubscriptionModelToDTO(s)
	}
	return subscriptions, totalCount, nil
}

// SubscriptionsByReceiver queries subscriptions with offset, limit, and receiver
func SubscriptionsByReceiver(offset, limit int, receiver string, dic *di.Container) (subscriptions []dtos.Subscription, totalCount uint32, err errors.EdgeX) {
	if receiver == "" {
		return subscriptions, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "receiver is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	subscriptionModels, err := dbClient.SubscriptionsByReceiver(offset, limit, receiver)
	if err == nil {
		totalCount, err = dbClient.SubscriptionCountByReceiver(receiver)
	}
	if err != nil {
		return subscriptions, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	subscriptions = make([]dtos.Subscription, len(subscriptionModels))
	for i, s := range subscriptionModels {
		subscriptions[i] = dtos.FromSubscriptionModelToDTO(s)
	}
	return subscriptions, totalCount, nil
}

// DeleteSubscriptionByName deletes the subscription by name
func DeleteSubscriptionByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	err := dbClient.DeleteSubscriptionByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// PatchSubscription executes the PATCH operation with the subscription DTO to replace the old data
func PatchSubscription(ctx context.Context, dto dtos.UpdateSubscription, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	subscription, err := subscriptionByDTO(dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	requests.ReplaceSubscriptionModelFieldsWithDTO(&subscription, dto)

	if len(subscription.Categories) == 0 && len(subscription.Labels) == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "subscription categories and labels can not be both empty", nil)
	}

	err = dbClient.UpdateSubscription(subscription)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Subscription patched on DB successfully. Correlation-ID: %s ", correlation.FromContext(ctx))
	return nil
}

func subscriptionByDTO(dbClient interfaces.DBClient, dto dtos.UpdateSubscription) (subscription models.Subscription, err errors.EdgeX) {
	// The ID or Name is required by DTO and the DTO also accepts empty string ID if the Name is provided
	if dto.Id != nil && *dto.Id != "" {
		subscription, err = dbClient.SubscriptionById(*dto.Id)
		if err != nil {
			return subscription, errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		subscription, err = dbClient.SubscriptionByName(*dto.Name)
		if err != nil {
			return subscription, errors.NewCommonEdgeXWrapper(err)
		}
	}
	if dto.Name != nil && *dto.Name != subscription.Name {
		return subscription, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("subscription name '%s' not match the existing '%s' ", *dto.Name, subscription.Name), nil)
	}
	return subscription, nil
}
