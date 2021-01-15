//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/common"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	v2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gomodule/redigo/redis"
)

const (
	ProvisionWatcherCollection            = "md|pw"
	ProvisionWatcherCollectionName        = ProvisionWatcherCollection + DBKeySeparator + v2.Name
	ProvisionWatcherCollectionLabel       = ProvisionWatcherCollection + DBKeySeparator + v2.Label
	ProvisionWatcherCollectionServiceName = ProvisionWatcherCollectionName + DBKeySeparator + v2.Service + DBKeySeparator + v2.Name
	ProvisionWatcherCollectionProfileName = ProvisionWatcherCollectionName + DBKeySeparator + v2.Profile + DBKeySeparator + v2.Name
)

// provisionWatcherStoredKey return the provision watcher's stored key which combines the collection name and object id
func provisionWatcherStoredKey(id string) string {
	return CreateKey(ProvisionWatcherCollection, id)
}

// addProvisionWatcher adds a new provision watcher into DB
func addProvisionWatcher(conn redis.Conn, pw models.ProvisionWatcher) (addedProvisionWatcher models.ProvisionWatcher, edgexErr errors.EdgeX) {
	// retrieve provision watcher by Id first to ensure there is no Id conflict; when Id exists, return duplicate error
	exists, edgexErr := objectIdExists(conn, provisionWatcherStoredKey(pw.Id))
	if edgexErr != nil {
		return addedProvisionWatcher, errors.NewCommonEdgeXWrapper(edgexErr)
	} else if exists {
		return addedProvisionWatcher, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher id %s already exists", pw.Id), edgexErr)
	}

	// verify if provision watcher name is unique or not
	exists, edgexErr = objectNameExists(conn, ProvisionWatcherCollectionName, pw.Name)
	if edgexErr != nil {
		return addedProvisionWatcher, errors.NewCommonEdgeXWrapper(edgexErr)
	} else if exists {
		return addedProvisionWatcher, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("provision watcher name %s already exists", pw.Name), edgexErr)
	}

	ts := common.MakeTimestamp()
	if pw.Created == 0 {
		pw.Created = ts
	}
	// query API will sort the result based on Modified, so even newly created device service shall specify Modified as Created
	pw.Modified = ts

	dsJSONBytes, err := json.Marshal(pw)
	if err != nil {
		return addedProvisionWatcher, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal provision watcher for Redis persistence", err)
	}

	redisKey := provisionWatcherStoredKey(pw.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(SET, redisKey, dsJSONBytes)
	_ = conn.Send(HSET, ProvisionWatcherCollectionName, pw.Name, redisKey)
	_ = conn.Send(ZADD, ProvisionWatcherCollection, pw.Modified, redisKey)
	_ = conn.Send(ZADD, CreateKey(ProvisionWatcherCollectionServiceName, pw.ServiceName), pw.Modified, redisKey)
	_ = conn.Send(ZADD, CreateKey(ProvisionWatcherCollectionProfileName, pw.ProfileName), pw.Modified, redisKey)
	for _, label := range pw.Labels {
		_ = conn.Send(ZADD, CreateKey(ProvisionWatcherCollectionLabel, label), pw.Modified, redisKey)
	}
	_, err = conn.Do(EXEC)
	if err != nil {
		edgexErr = errors.NewCommonEdgeX(errors.KindDatabaseError, "provision watcher creation failed", err)
	}

	return pw, edgexErr
}

// provisionWatcherById query provision watcher by id from DB
func provisionWatcherById(conn redis.Conn, id string) (provisionWatcher models.ProvisionWatcher, edgexErr errors.EdgeX) {
	edgexErr = getObjectById(conn, provisionWatcherStoredKey(id), &provisionWatcher)
	if edgexErr != nil {
		return provisionWatcher, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	return
}

// provisionWatcherByName query provision watcher by name from DB
func provisionWatcherByName(conn redis.Conn, name string) (provisionWatcher models.ProvisionWatcher, edgexErr errors.EdgeX) {
	edgexErr = getObjectByHash(conn, ProvisionWatcherCollectionName, name, &provisionWatcher)
	if edgexErr != nil {
		return provisionWatcher, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	return
}

// provisionWatchersByServiceName query provision watchers by offset, limit and service name
func provisionWatchersByServiceName(conn redis.Conn, offset int, limit int, name string) (provisionWatchers []models.ProvisionWatcher, edgexErr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { // -1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByRevRange(conn, CreateKey(ProvisionWatcherCollectionServiceName, name), offset, end)
	if err != nil {
		return provisionWatchers, errors.NewCommonEdgeXWrapper(err)
	}

	provisionWatchers = make([]models.ProvisionWatcher, len(objects))
	for i, in := range objects {
		pw := models.ProvisionWatcher{}
		err := json.Unmarshal(in, &pw)
		if err != nil {
			return []models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "provision watcher format parsing failed from the database", err)
		}
		provisionWatchers[i] = pw
	}

	return provisionWatchers, nil
}

// provisionWatchersByProfileName query provision watchers by offset, limit and profile name
func provisionWatchersByProfileName(conn redis.Conn, offset int, limit int, name string) (provisionWatchers []models.ProvisionWatcher, edgexErr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { // -1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByRevRange(conn, CreateKey(ProvisionWatcherCollectionProfileName, name), offset, end)
	if err != nil {
		return []models.ProvisionWatcher{}, errors.NewCommonEdgeXWrapper(err)
	}

	provisionWatchers = make([]models.ProvisionWatcher, len(objects))
	for i, in := range objects {
		pw := models.ProvisionWatcher{}
		err := json.Unmarshal(in, &pw)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "provision watcher format parsing failed from the database", err)
		}
		provisionWatchers[i] = pw
	}

	return provisionWatchers, nil
}

// provisionWatchersByLabels query provision watchers by offset, limit and labels
func provisionWatchersByLabels(conn redis.Conn, offset int, limit int, labels []string) (provisionWatchers []models.ProvisionWatcher, edgexErr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { // -1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByLabelsAndSomeRange(conn, ZREVRANGE, ProvisionWatcherCollection, labels, offset, end)
	if err != nil {
		return provisionWatchers, errors.NewCommonEdgeXWrapper(err)
	}

	provisionWatchers = make([]models.ProvisionWatcher, len(objects))
	for i, in := range objects {
		pw := models.ProvisionWatcher{}
		err := json.Unmarshal(in, &pw)
		if err != nil {
			return []models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "provision watcher format parsing failed from the database", err)
		}
		provisionWatchers[i] = pw
	}

	return
}

// deleteProvisionWatcherByName deletes the provision watcher by name
func deleteProvisionWatcherByName(conn redis.Conn, name string) errors.EdgeX {
	provisionWatcher, err := provisionWatcherByName(conn, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	err = deleteProvisionWatcher(conn, provisionWatcher)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// deleteProvisionWatcher deletes a provision watcher
func deleteProvisionWatcher(conn redis.Conn, pw models.ProvisionWatcher) errors.EdgeX {
	redisKey := provisionWatcherStoredKey(pw.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(DEL, redisKey)
	_ = conn.Send(HDEL, ProvisionWatcherCollectionName, pw.Name)
	_ = conn.Send(ZREM, ProvisionWatcherCollection, redisKey)
	_ = conn.Send(ZREM, CreateKey(ProvisionWatcherCollectionServiceName, pw.ServiceName), redisKey)
	_ = conn.Send(ZREM, CreateKey(ProvisionWatcherCollectionProfileName, pw.ProfileName), redisKey)
	for _, label := range pw.Labels {
		_ = conn.Send(ZREM, CreateKey(ProvisionWatcherCollectionLabel, label), redisKey)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "provision watcher deletion failed", err)
	}

	return nil
}
