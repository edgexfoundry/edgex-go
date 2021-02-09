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

// allSubscriptions queries subscriptions by offset and limit
func allSubscriptions(conn redis.Conn, offset, limit int) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, edgeXerr := getObjectsBySomeRange(conn, ZREVRANGE, SubscriptionCollection, offset, end)
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
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByRevRange(conn, CreateKey(SubscriptionCollectionCategory, category), offset, end)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToSubscriptions(objects)
}

// subscriptionsByLabel queries subscriptions by offset, limit, and label
func subscriptionsByLabel(conn redis.Conn, offset int, limit int, label string) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByRevRange(conn, CreateKey(SubscriptionCollectionLabel, label), offset, end)
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToSubscriptions(objects)
}

// subscriptionsByReceiver queries subscriptions by offset, limit, and receiver
func subscriptionsByReceiver(conn redis.Conn, offset int, limit int, receiver string) (subscriptions []models.Subscription, edgeXerr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByRevRange(conn, CreateKey(SubscriptionCollectionReceiver, receiver), offset, end)
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
