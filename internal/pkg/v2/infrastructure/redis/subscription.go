//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"

	"github.com/gomodule/redigo/redis"
)

const (
	SubscriptionCollection         = "sn|sub"
	SubscriptionCollectionName     = SubscriptionCollection + DBKeySeparator + v2.Name
	SubscriptionCollectionCategory = SubscriptionCollection + DBKeySeparator + v2.Category
	SubscriptionCollectionLabel    = SubscriptionCollection + DBKeySeparator + v2.Label
	SubscriptionCollectionReceiver = SubscriptionCollection + DBKeySeparator + v2.Receiver
)

// subscriptionStoredKey return the subscription's stored key which combines the collection name and object id
func subscriptionStoredKey(id string) string {
	return CreateKey(SubscriptionCollection, id)
}

// addSubscription adds a new subscription into DB
func addSubscription(conn redis.Conn, subscription models.Subscription) (models.Subscription, errors.EdgeX) {
	exists, edgeXerr := objectIdExists(conn, subscriptionStoredKey(subscription.Id))
	if edgeXerr != nil {
		return subscription, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return subscription, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("subscription id %s already exists", subscription.Id), edgeXerr)
	}

	exists, edgeXerr = objectNameExists(conn, SubscriptionCollectionName, subscription.Name)
	if edgeXerr != nil {
		return subscription, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return subscription, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("subscription name %s already exists", subscription.Name), edgeXerr)
	}

	ts := common.MakeTimestamp()
	if subscription.Created == 0 {
		subscription.Created = ts
	}
	subscription.Modified = ts

	dsJSONBytes, err := json.Marshal(subscription)
	if err != nil {
		return subscription, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal subscription for Redis persistence", err)
	}

	redisKey := subscriptionStoredKey(subscription.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, redisKey, dsJSONBytes)
	_ = conn.Send(ZADD, SubscriptionCollection, 0, redisKey)
	_ = conn.Send(HSET, SubscriptionCollectionName, subscription.Name, redisKey)
	for _, category := range subscription.Categories {
		_ = conn.Send(ZADD, CreateKey(SubscriptionCollectionCategory, string(category)), subscription.Modified, redisKey)
	}
	for _, label := range subscription.Labels {
		_ = conn.Send(ZADD, CreateKey(SubscriptionCollectionLabel, label), subscription.Modified, redisKey)
	}
	_ = conn.Send(ZADD, CreateKey(SubscriptionCollectionReceiver, subscription.Receiver), subscription.Modified, redisKey)
	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "subscription creation failed", err)
	}

	return subscription, edgeXerr
}
