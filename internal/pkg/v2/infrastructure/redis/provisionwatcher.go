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
