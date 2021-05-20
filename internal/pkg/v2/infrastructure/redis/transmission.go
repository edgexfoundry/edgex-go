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
	TransmissionCollection                 = "sn|trans"
	TransmissionCollectionStatus           = TransmissionCollection + DBKeySeparator + v2.Status
	TransmissionCollectionSubscriptionName = TransmissionCollection + DBKeySeparator + v2.Subscription + DBKeySeparator + v2.Name
	TransmissionCollectionNotificationId   = TransmissionCollection + DBKeySeparator + v2.Notification + DBKeySeparator + v2.Id
	TransmissionCollectionCreated          = TransmissionCollection + DBKeySeparator + v2.Created
)

// notificationStoredKey return the transmission's stored key which combines the collection name and object id
func transmissionStoredKey(id string) string {
	return CreateKey(TransmissionCollection, id)
}

// transmissionById query transmission by id from DB
func transmissionById(conn redis.Conn, id string) (trans models.Transmission, edgexErr errors.EdgeX) {
	edgexErr = getObjectById(conn, transmissionStoredKey(id), &trans)
	if edgexErr != nil {
		return trans, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	return
}

// sendAddTransmissionCmd sends redis command for adding transmission
func sendAddTransmissionCmd(conn redis.Conn, storedKey string, trans models.Transmission) errors.EdgeX {
	m, err := json.Marshal(trans)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal transmission for Redis persistence", err)
	}
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, TransmissionCollection, trans.Created, storedKey)
	_ = conn.Send(ZADD, TransmissionCollectionCreated, trans.Created, storedKey)
	_ = conn.Send(ZADD, CreateKey(TransmissionCollectionStatus, string(trans.Status)), trans.Created, storedKey)
	_ = conn.Send(ZADD, CreateKey(TransmissionCollectionSubscriptionName, trans.SubscriptionName), trans.Created, storedKey)
	_ = conn.Send(ZADD, CreateKey(TransmissionCollectionNotificationId, trans.NotificationId), trans.Created, storedKey)
	return nil
}

// addTransmission adds a new transmission into DB
func addTransmission(conn redis.Conn, trans models.Transmission) (models.Transmission, errors.EdgeX) {
	exists, edgeXerr := objectIdExists(conn, transmissionStoredKey(trans.Id))
	if edgeXerr != nil {
		return trans, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return trans, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("transmission id %s already exists", trans.Id), edgeXerr)
	}

	ts := common.MakeTimestamp()
	if trans.Created == 0 {
		trans.Created = ts
	}

	storedKey := transmissionStoredKey(trans.Id)
	_ = conn.Send(MULTI)
	edgeXerr = sendAddTransmissionCmd(conn, storedKey, trans)
	if edgeXerr != nil {
		return trans, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "transmission creation failed", err)
	}

	return trans, edgeXerr
}

// sendDeleteTransmissionCmd sends redis command to delete a transmission
func sendDeleteTransmissionCmd(conn redis.Conn, storedKey string, trans models.Transmission) {
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, TransmissionCollection, storedKey)
	_ = conn.Send(ZREM, TransmissionCollectionCreated, storedKey)
	_ = conn.Send(ZREM, CreateKey(TransmissionCollectionStatus, string(trans.Status)), storedKey)
	_ = conn.Send(ZREM, CreateKey(TransmissionCollectionSubscriptionName, trans.SubscriptionName), storedKey)
	_ = conn.Send(ZREM, CreateKey(TransmissionCollectionNotificationId, trans.NotificationId), storedKey)
}

// updateTransmission updates a transmission
func updateTransmission(conn redis.Conn, trans models.Transmission) errors.EdgeX {
	oldTransmission, edgeXerr := transmissionById(conn, trans.Id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	storedKey := transmissionStoredKey(trans.Id)

	_ = conn.Send(MULTI)
	sendDeleteTransmissionCmd(conn, storedKey, oldTransmission)
	edgeXerr = sendAddTransmissionCmd(conn, storedKey, trans)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "transmission update failed", err)
	}
	return nil
}

// transmissionsByTimeRange query transmissions by time range, offset, and limit
func transmissionsByTimeRange(conn redis.Conn, startTime int, endTime int, offset int, limit int) (transmissions []models.Transmission, edgeXerr errors.EdgeX) {
	objects, edgeXerr := getObjectsByScoreRange(conn, TransmissionCollectionCreated, startTime, endTime, offset, limit)
	if edgeXerr != nil {
		return transmissions, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return objectsToTransmissions(objects)
}

// allTransmissions queries transmissions by offset and limit
func allTransmissions(conn redis.Conn, offset, limit int) (transmissions []models.Transmission, edgeXerr errors.EdgeX) {
	objects, edgeXerr := getObjectsByRevRange(conn, TransmissionCollection, offset, limit)
	if edgeXerr != nil {
		return transmissions, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return objectsToTransmissions(objects)
}

// transmissionsByStatus queries transmissions by offset, limit, and status
func transmissionsByStatus(conn redis.Conn, offset int, limit int, status string) (transmissions []models.Transmission, err errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(TransmissionCollectionStatus, status), offset, limit)
	if err != nil {
		return transmissions, errors.NewCommonEdgeXWrapper(err)
	}

	return objectsToTransmissions(objects)
}

func objectsToTransmissions(objects [][]byte) (transmissions []models.Transmission, edgeXerr errors.EdgeX) {
	transmissions = make([]models.Transmission, len(objects))
	for i, o := range objects {
		trans := models.Transmission{}
		err := json.Unmarshal(o, &trans)
		if err != nil {
			return transmissions, errors.NewCommonEdgeX(errors.KindDatabaseError, "transmission format parsing failed from the database", err)
		}
		transmissions[i] = trans
	}
	return transmissions, nil
}

// DeleteProcessedTransmissionsByAge deletes the processed transmissions((ACKNOWLEDGED, SENT, ESCALATED) that are older than age.
// This function is implemented to starts up goroutines to delete processed transmissions in the background to achieve better performance.
func (c *Client) DeleteProcessedTransmissionsByAge(age int64) (err errors.EdgeX) {
	conn := c.Pool.Get()
	defer conn.Close()
	acknowledgedStoreKeys, err := transmissionStoreKeys(conn, CreateKey(TransmissionCollectionStatus, models.Acknowledged), age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	sentStoreKeys, err := transmissionStoreKeys(conn, CreateKey(TransmissionCollectionStatus, models.Sent), age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	escalatedStoreKeys, err := transmissionStoreKeys(conn, CreateKey(TransmissionCollectionStatus, models.Escalated), age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	go c.asyncDeleteTransmissionByStoreKeys(acknowledgedStoreKeys)
	go c.asyncDeleteTransmissionByStoreKeys(sentStoreKeys)
	go c.asyncDeleteTransmissionByStoreKeys(escalatedStoreKeys)
	return nil
}

// transmissionStoreKeys return the store keys of the transmission that are older than age.
func transmissionStoreKeys(conn redis.Conn, collectionKey string, age int64) ([]string, errors.EdgeX) {
	expireTimestamp := common.MakeTimestamp() - age

	storeKeys, err := redis.Strings(conn.Do(ZRANGEBYSCORE, collectionKey, 0, expireTimestamp))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("retrieve transmission storeKeys by %s failed", collectionKey), err)
	}
	return storeKeys, nil
}
