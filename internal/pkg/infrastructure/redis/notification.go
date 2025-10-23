//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"
	"strconv"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const (
	NotificationCollection         = "sn|notif"
	NotificationCollectionCategory = NotificationCollection + DBKeySeparator + common.Category
	NotificationCollectionLabel    = NotificationCollection + DBKeySeparator + common.Label
	NotificationCollectionSender   = NotificationCollection + DBKeySeparator + common.Sender
	NotificationCollectionSeverity = NotificationCollection + DBKeySeparator + common.Severity
	NotificationCollectionStatus   = NotificationCollection + DBKeySeparator + common.Status
	NotificationCollectionCreated  = NotificationCollection + DBKeySeparator + common.Created
	NotificationCollectionAck      = NotificationCollection + DBKeySeparator + common.Ack
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
	_ = conn.Send(ZADD, CreateKey(NotificationCollectionAck, strconv.FormatBool(n.Acknowledged)), n.Modified, storedKey)
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
func notificationsByCategory(conn redis.Conn, offset int, limit int, ack, category string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	redisKey := CreateKey(NotificationCollectionCategory, category)
	return getNotificationsByRedisKeyAndAck(conn, offset, limit, ack, redisKey)
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
func notificationsByLabel(conn redis.Conn, offset int, limit int, ack, label string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	redisKey := CreateKey(NotificationCollectionLabel, label)
	return getNotificationsByRedisKeyAndAck(conn, offset, limit, ack, redisKey)
}

// notificationById query notification by id from DB
func notificationById(conn redis.Conn, id string) (notification models.Notification, edgexErr errors.EdgeX) {
	edgexErr = getObjectById(conn, notificationStoredKey(id), &notification)
	if edgexErr != nil {
		return notification, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	return
}

// notificationByIds query notification by ids from DB
func notificationByIds(conn redis.Conn, ids []string) (notifications []models.Notification, edgexErr errors.EdgeX) {
	var storeKeys []string
	for _, id := range ids {
		storeKeys = append(storeKeys, notificationStoredKey(id))
	}
	objects, edgexErr := getObjectsByIds(conn, pkgCommon.ConvertStringsToInterfaces(storeKeys))
	if edgexErr != nil {
		return notifications, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	return convertObjectsToNotifications(objects)
}

// notificationsByStatus queries notifications by offset, limit, and status
func notificationsByStatus(conn redis.Conn, offset int, limit int, ack, status string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	redisKey := CreateKey(NotificationCollectionStatus, status)
	return getNotificationsByRedisKeyAndAck(conn, offset, limit, ack, redisKey)
}

// notificationsByTimeRange query notifications by time range, offset, limit, and ack
func notificationsByTimeRange(conn redis.Conn, startTime int64, endTime int64, offset int, limit int, ack string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	redisKey := NotificationCollectionCreated
	if len(ack) > 0 {
		args := redis.Args{}
		cacheSet := uuid.New().String()
		defer func() {
			// delete cache set
			_, _ = conn.Do(DEL, cacheSet)
		}()
		command := ZINTERSTORE
		args = args.Add(cacheSet, 2, redisKey, CreateKey(NotificationCollectionAck, ack), WEIGHTS, 1, 0)
		_, err := conn.Do(command, args...)
		if err != nil {
			return notifications, errors.NewCommonEdgeX(errors.KindDatabaseError,
				fmt.Sprintf("failed to execute %s command with args %v", command, args), err)
		}
		redisKey = cacheSet
	}

	objects, edgeXerr := getObjectsByScoreRange(conn, redisKey, startTime, endTime, offset, limit)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	notifications, edgeXerr = convertObjectsToNotifications(objects)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return notifications, nil
}

// notificationByQueryConditions query notifications by offset, limit, categories and time range
func notificationByQueryConditions(conn redis.Conn, offset, limit int, condition requests.NotificationQueryCondition,
	ack string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	if len(condition.Category) == 0 {
		return notificationsByTimeRange(conn, condition.Start, condition.End, offset, limit, ack)
	}

	cacheSetUnionCategory := uuid.New().String()
	cacheSetIntersectionCreatedAndCategory := uuid.New().String()
	cacheSetIntersectionAck := uuid.New().String()
	cacheSet := cacheSetIntersectionCreatedAndCategory
	defer func() {
		// delete cache set
		_, _ = conn.Do(DEL, cacheSetUnionCategory, cacheSetIntersectionCreatedAndCategory, cacheSetIntersectionAck)
	}()

	var redisKeys []string
	for _, c := range condition.Category {
		redisKeys = append(redisKeys, CreateKey(NotificationCollectionCategory, c))
	}
	args := redis.Args{}
	args = args.Add(cacheSetUnionCategory, len(redisKeys))
	for _, key := range redisKeys {
		args = args.Add(key)
	}
	// find all notifications by category and store the result to cache
	command := ZUNIONSTORE
	_, err := conn.Do(command, args...)
	if err != nil {
		return notifications, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to execute %s command with args %v", command, args), err)
	}

	if len(ack) > 0 {
		cacheSet = cacheSetIntersectionAck
	}

	objects, edgeXerr := getObjectsByScoreRange(conn, cacheSet, condition.Start, condition.End, offset, limit)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	notifications, edgeXerr = convertObjectsToNotifications(objects)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return notifications, nil
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
	_ = conn.Send(ZREM, CreateKey(NotificationCollectionAck, strconv.FormatBool(n.Acknowledged)), storedKey)
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

// deleteNotificationByIds deletes the notification by id and all of its associated transmissions
func deleteNotificationByIds(conn redis.Conn, ids []string) errors.EdgeX {
	notifications, edgexErr := notificationByIds(conn, ids)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	var transmissions []models.Transmission
	for _, notification := range notifications {
		trans, edgexErr := transmissionsByNotificationId(conn, 0, -1, notification.Id)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
		transmissions = append(transmissions, trans...)
	}
	_ = conn.Send(MULTI)
	for _, notification := range notifications {
		sendDeleteNotificationCmd(conn, notificationStoredKey(notification.Id), notification)
	}
	for _, transmission := range transmissions {
		sendDeleteTransmissionCmd(conn, transmissionStoredKey(transmission.Id), transmission)
	}
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

// updateNotificationAckStatus bulk updates acknowledgement status
func updateNotificationAckStatus(conn redis.Conn, ack bool, notifications []models.Notification) errors.EdgeX {
	_ = conn.Send(MULTI)
	for _, n := range notifications {
		storedKey := notificationStoredKey(n.Id)
		sendDeleteNotificationCmd(conn, storedKey, n)
		n.Modified = pkgCommon.MakeTimestamp()
		n.Acknowledged = ack
		edgexErr := sendAddNotificationCmd(conn, storedKey, n)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "notification acknowledgement status update failed", err)
	}
	return nil
}

func notificationsByCategoriesAndLabels(conn redis.Conn, offset int, limit int, categories []string, labels []string, ack string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	cacheSetUnion := uuid.New().String()
	cacheSetIntersection := uuid.New().String()
	cacheSet := cacheSetUnion
	defer func() {
		// delete cache set
		_, _ = conn.Do(DEL, cacheSetUnion, cacheSetIntersection)
	}()

	var redisKeys []string
	for _, c := range categories {
		redisKeys = append(redisKeys, CreateKey(NotificationCollectionCategory, c))
	}
	for _, label := range labels {
		redisKeys = append(redisKeys, CreateKey(NotificationCollectionLabel, label))
	}

	args := redis.Args{}
	args = args.Add(cacheSetUnion, len(redisKeys))
	for _, key := range redisKeys {
		args = args.Add(key)
	}

	if len(ack) > 0 {
		var redisKeys []string
		redisKeys = append(redisKeys, cacheSetUnion, CreateKey(NotificationCollectionAck, ack))
		args := redis.Args{}
		args = args.Add(cacheSetIntersection, len(redisKeys))
		for _, key := range redisKeys {
			args = args.Add(key)
		}
		cacheSet = cacheSetIntersection
	}

	objects, edgeXerr := getObjectsByRevRange(conn, cacheSet, offset, limit)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	notifications, edgeXerr = convertObjectsToNotifications(objects)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return notifications, nil
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
	defer closeRedisConnection(conn, c.loggingClient)

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
	defer closeRedisConnection(conn, c.loggingClient)

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
	defer closeRedisConnection(conn, c.loggingClient)
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
	defer closeRedisConnection(conn, c.loggingClient)
	collectionKey := CreateKey(NotificationCollectionStatus, models.Processed)
	ncStoreKeys, transStoreKeys, err := notificationAndTransmissionStoreKeys(conn, collectionKey, age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	go c.asyncDeleteNotificationByStoreKeys(ncStoreKeys)
	go c.asyncDeleteTransmissionByStoreKeys(transStoreKeys)
	return nil
}

func getNotificationsByRedisKeyAndAck(conn redis.Conn, offset, limit int, ack, redisKey string) (notifications []models.Notification, edgeXerr errors.EdgeX) {
	var objects [][]byte
	if len(ack) > 0 {
		objects, edgeXerr = intersectionObjectsByKeys(conn, offset, limit, redisKey, CreateKey(NotificationCollectionAck, ack))
		if edgeXerr != nil {
			return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	} else {
		objects, edgeXerr = getObjectsByRevRange(conn, redisKey, offset, limit)
		if edgeXerr != nil {
			return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}
	return convertObjectsToNotifications(objects)
}

func latestNotificationByOffset(conn redis.Conn, offset int) (notification models.Notification, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, NotificationCollection, offset, 1)
	if err != nil {
		return notification, errors.NewCommonEdgeXWrapper(err)
	}
	notifications, err := convertObjectsToNotifications(objects)
	if err != nil {
		return notification, errors.NewCommonEdgeXWrapper(err)
	}
	if len(notifications) > 1 {
		return notification, errors.NewCommonEdgeX(errors.KindServerError, "the query result should not greater than one notification", nil)
	}
	if len(notifications) == 0 {
		return notification, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("notification not found from the offset %d", offset), nil)
	}
	return notifications[0], nil
}
