//
// Copyright (C) 2021 IOTech Ltd
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
	NotificationCollection         = "sn|notif"
	NotificationCollectionCategory = NotificationCollection + DBKeySeparator + v2.Category
	NotificationCollectionLabel    = NotificationCollection + DBKeySeparator + v2.Label
	NotificationCollectionSender   = NotificationCollection + DBKeySeparator + v2.Sender
	NotificationCollectionSeverity = NotificationCollection + DBKeySeparator + v2.Severity
	NotificationCollectionStatus   = NotificationCollection + DBKeySeparator + v2.Status
)

// notificationStoredKey return the notification's stored key which combines the collection name and object id
func notificationStoredKey(id string) string {
	return CreateKey(NotificationCollection, id)
}

// addNotification adds a new notification into DB
func addNotification(conn redis.Conn, notification models.Notification) (models.Notification, errors.EdgeX) {
	exists, edgeXerr := objectIdExists(conn, notificationStoredKey(notification.Id))
	if edgeXerr != nil {
		return notification, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return notification, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("notification id %s already exists", notification.Id), edgeXerr)
	}

	ts := common.MakeTimestamp()
	if notification.Created == 0 {
		notification.Created = ts
	}
	notification.Modified = ts

	notifJSONBytes, err := json.Marshal(notification)
	if err != nil {
		return notification, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal notification for Redis persistence", err)
	}

	redisKey := notificationStoredKey(notification.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, redisKey, notifJSONBytes)
	_ = conn.Send(ZADD, NotificationCollection, 0, redisKey)
	if len(notification.Category) > 0 {
		_ = conn.Send(ZADD, CreateKey(NotificationCollectionCategory, notification.Category), notification.Modified, redisKey)
	}
	for _, label := range notification.Labels {
		_ = conn.Send(ZADD, CreateKey(NotificationCollectionLabel, label), notification.Modified, redisKey)
	}
	_ = conn.Send(ZADD, CreateKey(NotificationCollectionSender, notification.Sender), notification.Modified, redisKey)
	_ = conn.Send(ZADD, CreateKey(NotificationCollectionSeverity, string(notification.Severity)), notification.Modified, redisKey)
	_ = conn.Send(ZADD, CreateKey(NotificationCollectionStatus, string(notification.Status)), notification.Modified, redisKey)
	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "notification creation failed", err)
	}

	return notification, edgeXerr
}
