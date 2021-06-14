//
// Copyright (C) 2021 IOTech Ltd
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
	NotificationCollection         = "sn|notif"
	NotificationCollectionCategory = NotificationCollection + DBKeySeparator + common.Category
	NotificationCollectionLabel    = NotificationCollection + DBKeySeparator + common.Label
	NotificationCollectionSender   = NotificationCollection + DBKeySeparator + common.Sender
	NotificationCollectionSeverity = NotificationCollection + DBKeySeparator + common.Severity
	NotificationCollectionStatus   = NotificationCollection + DBKeySeparator + common.Status
	NotificationCollectionCreated  = NotificationCollection + DBKeySeparator + common.Created
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
	_ = conn.Send(ZADD, NotificationCollection, n.Modified, storedKey)
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

	ts := pkgCommon.MakeTimestamp()
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

// deleteNotificationById deletes the notification by id and all of its associated transmissions
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

	// Remove the associated transmissions
	transStoreKeys, err := redis.Strings(conn.Do(ZRANGE, CreateKey(TransmissionCollectionNotificationId, notification.Id), 0, -1))
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "fail to retrieve transmission storeKeys", err)
	}
	for _, storeKey := range transStoreKeys {
		err = deleteTransmissionById(conn, idFromStoredKey(storeKey))
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	return nil
}

// updateNotification updates a notification
func updateNotification(conn redis.Conn, n models.Notification) errors.EdgeX {
	oldNotification, edgeXerr := notificationById(conn, n.Id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	n.Modified = pkgCommon.MakeTimestamp()
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

// notificationAndTransmissionStoreKeys return the store keys of the notification and transmission that are older than age.
func notificationAndTransmissionStoreKeys(conn redis.Conn, collectionKey string, age int64) ([]string, []string, errors.EdgeX) {
	expireTimestamp := pkgCommon.MakeTimestamp() - age

	ncStoreKeys, err := redis.Strings(conn.Do(ZRANGEBYSCORE, collectionKey, 0, expireTimestamp))
	if err != nil {
		return nil, nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("retrieve nitification storeKeys by %s failed", NotificationCollection), err)
	}

	var transStoreKeys []string
	for _, ncStoreKey := range ncStoreKeys {
		keys, err := redis.Strings(conn.Do(ZRANGE, CreateKey(TransmissionCollectionNotificationId, idFromStoredKey(ncStoreKey)), 0, -1))
		if err != nil {
			return nil, nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "fail to retrieve transmission storeKeys", err)
		}
		transStoreKeys = append(transStoreKeys, keys...)
	}

	return ncStoreKeys, transStoreKeys, nil
}

// asyncDeleteNotificationByStoreKeys deletes all notifications with given storeKeys. This function is implemented to be run as a
// separate goroutine in the background to achieve better performance, so this function return nothing.  When
// encountering any errors during deletion, this function will simply log the error.
func (c *Client) asyncDeleteNotificationByStoreKeys(storeKeys []string) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, edgeXerr := getObjectsByIds(conn, pkgCommon.ConvertStringsToInterfaces(storeKeys))
	if edgeXerr != nil {
		c.loggingClient.Errorf("Deleted notifications failed while retrieving objects by storeKeys, %v", edgeXerr)
		return
	}

	// cmdSize is used to count the notification deletion command
	cmdSize := 0
	// start the transaction before iteration
	_ = conn.Send(MULTI)
	// iterate each notifications for deletion in batch
	for i, o := range objects {
		nc := models.Notification{}
		err := json.Unmarshal(o, &nc)
		if err != nil {
			c.loggingClient.Errorf("unable to marshal notification.  Err: %s", err.Error())
			continue
		}
		sendDeleteNotificationCmd(conn, notificationStoredKey(nc.Id), nc)
		cmdSize++

		if cmdSize >= c.BatchSize {
			_, err = conn.Do(EXEC)
			if err != nil {
				c.loggingClient.Errorf("unable to execute batch notification deletion, %v", err)
				continue
			}
			// reset cmdSize to zero if EXEC is successfully executed without error
			cmdSize = 0
			// rerun another transaction when iteration is not finished
			if i < len(objects)-1 {
				_ = conn.Send(MULTI)
			}
		}
	}

	// Iteration is finished but there are commands need to execute
	if cmdSize > 0 {
		_, err := conn.Do(EXEC)
		if err != nil {
			c.loggingClient.Errorf("unable to execute batch notification deletion, %v", err)
		}
	}
}

// asyncDeleteTransmissionByStoreKeys deletes all transmissions with given storeKeys. This function is implemented to be run as a
// separate goroutine in the background to achieve better performance, so this function return nothing.  When
// encountering any errors during deletion, this function will simply log the error.
func (c *Client) asyncDeleteTransmissionByStoreKeys(storeKeys []string) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, edgeXerr := getObjectsByIds(conn, pkgCommon.ConvertStringsToInterfaces(storeKeys))
	if edgeXerr != nil {
		c.loggingClient.Errorf("Deleted transmissions failed while retrieving objects by storeKeys, %v", edgeXerr)
		return
	}

	// cmdSize is used to count the transmission deletion command
	cmdSize := 0
	// start the transaction before iteration
	_ = conn.Send(MULTI)
	// iterate each notifications for deletion in batch
	for i, o := range objects {
		trans := models.Transmission{}
		err := json.Unmarshal(o, &trans)
		if err != nil {
			c.loggingClient.Errorf("unable to marshal transmission.  Err: %s", err.Error())
			continue
		}
		sendDeleteTransmissionCmd(conn, transmissionStoredKey(trans.Id), trans)
		cmdSize++

		if cmdSize >= c.BatchSize {
			_, err = conn.Do(EXEC)
			if err != nil {
				c.loggingClient.Errorf("unable to execute batch transmission deletion, %v", err)
				continue
			}
			// reset cmdSize to zero if EXEC is successfully executed without error
			cmdSize = 0
			// rerun another transaction when iteration is not finished
			if i < len(objects)-1 {
				_ = conn.Send(MULTI)
			}
		}
	}

	// Iteration is finished but there are commands need to execute
	if cmdSize > 0 {
		_, err := conn.Do(EXEC)
		if err != nil {
			c.loggingClient.Errorf("unable to execute batch transmission deletion, %v", err)
		}
	}
}

// CleanupNotificationsByAge deletes notifications and their corresponding transmissions that are older than age.
// This function is implemented to starts up two goroutines to delete transmissions and notifications in the background to achieve better performance.
func (c *Client) CleanupNotificationsByAge(age int64) (err errors.EdgeX) {
	conn := c.Pool.Get()
	defer conn.Close()
	ncStoreKeys, transStoreKeys, err := notificationAndTransmissionStoreKeys(conn, NotificationCollection, age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	go c.asyncDeleteNotificationByStoreKeys(ncStoreKeys)
	go c.asyncDeleteTransmissionByStoreKeys(transStoreKeys)
	return nil
}

// DeleteProcessedNotificationsByAge deletes processed notifications and their corresponding transmissions that are older than age.
// This function is implemented to starts up two goroutines to delete transmissions and notifications in the background to achieve better performance.
func (c *Client) DeleteProcessedNotificationsByAge(age int64) (err errors.EdgeX) {
	conn := c.Pool.Get()
	defer conn.Close()
	collectionKey := CreateKey(NotificationCollectionStatus, models.Processed)
	ncStoreKeys, transStoreKeys, err := notificationAndTransmissionStoreKeys(conn, collectionKey, age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	go c.asyncDeleteNotificationByStoreKeys(ncStoreKeys)
	go c.asyncDeleteTransmissionByStoreKeys(transStoreKeys)
	return nil
}
