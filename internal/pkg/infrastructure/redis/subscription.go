//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gomodule/redigo/redis"
)

const (
	SubscriptionCollection         = "sn|sub"
	SubscriptionCollectionName     = SubscriptionCollection + DBKeySeparator + common.Name
	SubscriptionCollectionCategory = SubscriptionCollection + DBKeySeparator + common.Category
	SubscriptionCollectionLabel    = SubscriptionCollection + DBKeySeparator + common.Label
	SubscriptionCollectionReceiver = SubscriptionCollection + DBKeySeparator + common.Receiver
)

// subscriptionStoredKey return the subscription's stored key which combines the collection name and object id
func subscriptionStoredKey(id string) string {
	return CreateKey(SubscriptionCollection, id)
}

// sendAddSubscriptionCmd sends redis command for adding subscription
func sendAddSubscriptionCmd(conn redis.Conn, storedKey string, subscription models.Subscription) errors.EdgeX {
	m, err := json.Marshal(subscription)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal subscription for Redis persistence", err)
	}
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, SubscriptionCollection, 0, storedKey)
	_ = conn.Send(HSET, SubscriptionCollectionName, subscription.Name, storedKey)
	for _, category := range subscription.Categories {
		_ = conn.Send(ZADD, CreateKey(SubscriptionCollectionCategory, string(category)), subscription.Modified, storedKey)
	}
	for _, label := range subscription.Labels {
		_ = conn.Send(ZADD, CreateKey(SubscriptionCollectionLabel, label), subscription.Modified, storedKey)
	}
	_ = conn.Send(ZADD, CreateKey(SubscriptionCollectionReceiver, subscription.Receiver), subscription.Modified, storedKey)
	return nil
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

	ts := pkgCommon.MakeTimestamp()
	if subscription.Created == 0 {
		subscription.Created = ts
	}
	subscription.Modified = ts

	storedKey := subscriptionStoredKey(subscription.Id)
	_ = conn.Send(MULTI)
	edgeXerr = sendAddSubscriptionCmd(conn, storedKey, subscription)
	if edgeXerr != nil {
		return subscription, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "subscription creation failed", err)
	}

	return subscription, edgeXerr
}

// allSubscriptions queries subscriptions by offset and limit
func allSubscriptions(conn redis.Conn, offset, limit int) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	objects, edgeXerr := getObjectsByRevRange(conn, SubscriptionCollection, offset, limit)
	if edgeXerr != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	subscriptions = make([]models.Subscription, len(objects))
	for i, o := range objects {
		s := models.Subscription{}
		err := json.Unmarshal(o, &s)
		if err != nil {
			return []models.Subscription{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "subscription format parsing failed from the database", err)
		}
		subscriptions[i] = s
	}
	return subscriptions, nil
}

// subscriptionById query subscription by id from DB
func subscriptionById(conn redis.Conn, id string) (subscription models.Subscription, edgexErr errors.EdgeX) {
	edgexErr = getObjectById(conn, subscriptionStoredKey(id), &subscription)
	if edgexErr != nil {
		return subscription, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	return
}

// subscriptionByName queries subscription by name
func subscriptionByName(conn redis.Conn, name string) (subscription models.Subscription, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectByHash(conn, SubscriptionCollectionName, name, &subscription)
	if edgeXerr != nil {
		return subscription, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return
}

// subscriptionsByCategory queries subscriptions by offset, limit, and category
func subscriptionsByCategory(conn redis.Conn, offset int, limit int, category string) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(SubscriptionCollectionCategory, category), offset, limit)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToSubscriptions(objects)
}

// subscriptionsByLabel queries subscriptions by offset, limit, and label
func subscriptionsByLabel(conn redis.Conn, offset int, limit int, label string) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(SubscriptionCollectionLabel, label), offset, limit)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToSubscriptions(objects)
}

// subscriptionsByReceiver queries subscriptions by offset, limit, and receiver
func subscriptionsByReceiver(conn redis.Conn, offset int, limit int, receiver string) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(SubscriptionCollectionReceiver, receiver), offset, limit)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToSubscriptions(objects)
}

func convertObjectsToSubscriptions(objects [][]byte) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	subscriptions = make([]models.Subscription, len(objects))
	for i, o := range objects {
		s := models.Subscription{}
		err := json.Unmarshal(o, &s)
		if err != nil {
			return []models.Subscription{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "subscription format parsing failed from the database", err)
		}
		subscriptions[i] = s
	}
	return subscriptions, nil
}

// deleteSubscriptionByName deletes the subscription by name
func deleteSubscriptionByName(conn redis.Conn, name string) errors.EdgeX {
	subscription, err := subscriptionByName(conn, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = deleteSubscription(conn, subscription)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// sendDeleteSubscriptionCmd sends redis command to delete a subscription
func sendDeleteSubscriptionCmd(conn redis.Conn, storedKey string, subscription models.Subscription) {
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, SubscriptionCollection, storedKey)
	_ = conn.Send(HDEL, SubscriptionCollectionName, subscription.Name)
	for _, category := range subscription.Categories {
		_ = conn.Send(ZREM, CreateKey(SubscriptionCollectionCategory, string(category)), storedKey)
	}
	for _, label := range subscription.Labels {
		_ = conn.Send(ZREM, CreateKey(SubscriptionCollectionLabel, label), storedKey)
	}
	_ = conn.Send(ZREM, CreateKey(SubscriptionCollectionReceiver, subscription.Receiver), storedKey)
}

// deleteSubscription deletes a subscription
func deleteSubscription(conn redis.Conn, subscription models.Subscription) errors.EdgeX {
	storedKey := subscriptionStoredKey(subscription.Id)
	_ = conn.Send(MULTI)
	sendDeleteSubscriptionCmd(conn, storedKey, subscription)
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "subscription deletion failed", err)
	}
	return nil
}

// updateSubscription updates an subscription
func updateSubscription(conn redis.Conn, subscription models.Subscription) errors.EdgeX {
	oldSubscription, edgeXerr := subscriptionByName(conn, subscription.Name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	subscription.Modified = pkgCommon.MakeTimestamp()
	storedKey := subscriptionStoredKey(subscription.Id)

	_ = conn.Send(MULTI)
	sendDeleteSubscriptionCmd(conn, storedKey, oldSubscription)
	edgeXerr = sendAddSubscriptionCmd(conn, storedKey, subscription)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "subscription update failed", err)
	}
	return nil
}

func subscriptionsByCategoriesAndLabels(conn redis.Conn, offset int, limit int, categories []string, labels []string) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	var redisKeys []string
	for _, c := range categories {
		redisKeys = append(redisKeys, CreateKey(SubscriptionCollectionCategory, c))
	}
	for _, label := range labels {
		redisKeys = append(redisKeys, CreateKey(SubscriptionCollectionLabel, label))
	}

	objects, err := intersectionObjectsByKeys(conn, offset, limit, redisKeys...)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}
	return convertObjectsToSubscriptions(objects)
}
