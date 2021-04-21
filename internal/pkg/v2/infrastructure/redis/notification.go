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
	NotificationCollectionCreated  = NotificationCollection + DBKeySeparator + v2.Created
)

// notificationStoredKey return the notification's stored key which combines the collection name and object id
func notificationStoredKey(id string) string {
	return CreateKey(NotificationCollection, id)
}

// sendAddNotificationCmd sends redis command for adding notification
func sendAddNotificationCmd(conn redis.Conn, storedKey string, n models.Notification) errors.EdgeX {
	m, err := json.Marshal(n)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal notification for Redis persistence", err)
	}
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, NotificationCollection, 0, storedKey)
	_ = conn.Send(ZADD, NotificationCollectionCreated, n.Created, storedKey)
	if len(n.Category) > 0 {
		_ = conn.Send(ZADD, CreateKey(NotificationCollectionCategory, n.Category), n.Modified, storedKey)
	}
	for _, label := range n.Labels {
		_ = conn.Send(ZADD, CreateKey(NotificationCollectionLabel, label), n.Modified, storedKey)
	}
	_ = conn.Send(ZADD, CreateKey(NotificationCollectionSender, n.Sender), n.Modified, storedKey)
	_ = conn.Send(ZADD, CreateKey(NotificationCollectionSeverity, string(n.Severity)), n.Modified, storedKey)
	_ = conn.Send(ZADD, CreateKey(NotificationCollectionStatus, string(n.Status)), n.Modified, storedKey)
	return nil
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

	storedKey := notificationStoredKey(notification.Id)
	_ = conn.Send(MULTI)
	edgeXerr = sendAddNotificationCmd(conn, storedKey, notification)
	if edgeXerr != nil {
		return notification, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "notification creation failed", err)
	}

	return notification, edgeXerr
}

// notificationsByCategory queries notifications by offset, limit, and category
func notificationsByCategory(conn redis.Conn, offset int, limit int, category string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(NotificationCollectionCategory, category), offset, limit)
	if err != nil {
		return notifications, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToNotifications(objects)
}

func convertObjectsToNotifications(objects [][]byte) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	notifications = make([]models.Notification, len(objects))
	for i, o := range objects {
		s := models.Notification{}
		err := json.Unmarshal(o, &s)
		if err != nil {
			return []models.Notification{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "notification format parsing failed from the database", err)
		}
		notifications[i] = s
	}
	return notifications, nil
}

// notificationsByLabel queries notifications by offset, limit, and label
func notificationsByLabel(conn redis.Conn, offset int, limit int, label string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(NotificationCollectionLabel, label), offset, limit)
	if err != nil {
		return notifications, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToNotifications(objects)
}

// notificationById query notification by id from DB
func notificationById(conn redis.Conn, id string) (notification models.Notification, edgexErr errors.EdgeX) {
	edgexErr = getObjectById(conn, notificationStoredKey(id), &notification)
	if edgexErr != nil {
		return notification, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	return
}

// notificationsByStatus queries notifications by offset, limit, and status
func notificationsByStatus(conn redis.Conn, offset int, limit int, status string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(NotificationCollectionStatus, status), offset, limit)
	if err != nil {
		return notifications, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToNotifications(objects)
}

// notificationsByTimeRange query notifications by time range, offset, and limit
func notificationsByTimeRange(conn redis.Conn, startTime int, endTime int, offset int, limit int) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	objects, edgeXerr := getObjectsByScoreRange(conn, NotificationCollectionCreated, startTime, endTime, offset, limit)
	if edgeXerr != nil {
		return notifications, edgeXerr
	}
	return convertObjectsToNotifications(objects)
}

// sendDeleteNotificationCmd sends redis command to delete a notification
func sendDeleteNotificationCmd(conn redis.Conn, storedKey string, n models.Notification) {
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, NotificationCollection, storedKey)
	_ = conn.Send(ZREM, NotificationCollectionCreated, storedKey)
	if len(n.Category) > 0 {
		_ = conn.Send(ZREM, CreateKey(NotificationCollectionCategory, n.Category), storedKey)
	}
	for _, label := range n.Labels {
		_ = conn.Send(ZREM, CreateKey(NotificationCollectionLabel, label), storedKey)
	}
	_ = conn.Send(ZREM, CreateKey(NotificationCollectionSender, n.Sender), storedKey)
	_ = conn.Send(ZREM, CreateKey(NotificationCollectionSeverity, string(n.Severity)), storedKey)
	_ = conn.Send(ZREM, CreateKey(NotificationCollectionStatus, string(n.Status)), storedKey)
}

// deleteNotificationById deletes the notification by id
func deleteNotificationById(conn redis.Conn, id string) errors.EdgeX {
	notification, edgexErr := notificationById(conn, id)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	storedKey := notificationStoredKey(notification.Id)
	_ = conn.Send(MULTI)
	sendDeleteNotificationCmd(conn, storedKey, notification)
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "notification deletion failed", err)
	}
	return nil
}

// updateNotification updates a notification
func updateNotification(conn redis.Conn, n models.Notification) errors.EdgeX {
	oldNotification, edgeXerr := notificationById(conn, n.Id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	n.Modified = common.MakeTimestamp()
	storedKey := notificationStoredKey(n.Id)

	_ = conn.Send(MULTI)
	sendDeleteNotificationCmd(conn, storedKey, oldNotification)
	edgeXerr = sendAddNotificationCmd(conn, storedKey, n)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "notification update failed", err)
	}
	return nil
}

func notificationsByCategoriesAndLabels(conn redis.Conn, offset int, limit int, categories []string, labels []string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	var redisKeys []string
	for _, c := range categories {
		redisKeys = append(redisKeys, CreateKey(NotificationCollectionCategory, c))
	}
	for _, label := range labels {
		redisKeys = append(redisKeys, CreateKey(NotificationCollectionLabel, label))
	}

	objects, err := unionObjectsByKeys(conn, offset, limit, redisKeys...)
	if err != nil {
		return notifications, errors.NewCommonEdgeXWrapper(err)
	}
	return convertObjectsToNotifications(objects)
}
