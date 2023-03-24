//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/gomodule/redigo/redis"
)

const (
	ProvisionWatcherCollection            = "md|pw"
	ProvisionWatcherCollectionName        = ProvisionWatcherCollection + DBKeySeparator + common.Name
	ProvisionWatcherCollectionLabel       = ProvisionWatcherCollection + DBKeySeparator + common.Label
	ProvisionWatcherCollectionServiceName = ProvisionWatcherCollectionName + DBKeySeparator + common.Service + DBKeySeparator + common.Name
	ProvisionWatcherCollectionProfileName = ProvisionWatcherCollectionName + DBKeySeparator + common.Profile + DBKeySeparator + common.Name
)

// provisionWatcherStoredKey return the provision watcher's stored key which combines the collection name and object id
func provisionWatcherStoredKey(id string) string {
	return CreateKey(ProvisionWatcherCollection, id)
}

// sendAddProvisionWatcherCmd send redis command for adding provision watcher
func sendAddProvisionWatcherCmd(conn redis.Conn, storedKey string, pw models.ProvisionWatcher) errors.EdgeX {
	m, err := json.Marshal(pw)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal provision watcher for Redis persistence", err)
	}
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(HSET, ProvisionWatcherCollectionName, pw.Name, storedKey)
	_ = conn.Send(ZADD, ProvisionWatcherCollection, pw.Modified, storedKey)
	_ = conn.Send(ZADD, CreateKey(ProvisionWatcherCollectionServiceName, pw.DiscoveredDevice.ServiceName), pw.Modified, storedKey)
	_ = conn.Send(ZADD, CreateKey(ProvisionWatcherCollectionProfileName, pw.DiscoveredDevice.ProfileName), pw.Modified, storedKey)
	for _, label := range pw.Labels {
		_ = conn.Send(ZADD, CreateKey(ProvisionWatcherCollectionLabel, label), pw.Modified, storedKey)
	}
	return nil
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

	// check the associated ServiceName and ProfileName existence
	exists, edgexErr = deviceServiceNameExist(conn, pw.DiscoveredDevice.ServiceName)
	if edgexErr != nil {
		return addedProvisionWatcher, errors.NewCommonEdgeXWrapper(edgexErr)
	} else if !exists {
		return addedProvisionWatcher, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exists", pw.DiscoveredDevice.ServiceName), edgexErr)
	}
	exists, edgexErr = deviceProfileNameExists(conn, pw.DiscoveredDevice.ProfileName)
	if edgexErr != nil {
		return addedProvisionWatcher, errors.NewCommonEdgeXWrapper(edgexErr)
	} else if !exists {
		return addedProvisionWatcher, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exists", pw.DiscoveredDevice.ProfileName), edgexErr)
	}

	ts := pkgCommon.MakeTimestamp()
	if pw.Created == 0 {
		pw.Created = ts
	}
	// query API will sort the result based on Modified, so even newly created provision watcher shall specify Modified as Created
	pw.Modified = ts
	storedKey := provisionWatcherStoredKey(pw.Id)
	_ = conn.Send(MULTI)
	edgexErr = sendAddProvisionWatcherCmd(conn, storedKey, pw)
	_, err := conn.Do(EXEC)
	if err != nil {
		edgexErr = errors.NewCommonEdgeX(errors.KindDatabaseError, "provision watcher creation failed", err)
	}

	return pw, edgexErr
}

// provisionWatcherById query provision watcher by id from DB
func provisionWatcherById(conn redis.Conn, id string) (provisionWatcher models.ProvisionWatcher, edgexErr errors.EdgeX) {
	edgexErr = getObjectById(conn, provisionWatcherStoredKey(id), &provisionWatcher)
	if edgexErr != nil {
		return provisionWatcher, errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("fail to query provision watcher by id %s", id), edgexErr)
	}

	return
}

// provisionWatcherByName query provision watcher by name from DB
func provisionWatcherByName(conn redis.Conn, name string) (provisionWatcher models.ProvisionWatcher, edgexErr errors.EdgeX) {
	edgexErr = getObjectByHash(conn, ProvisionWatcherCollectionName, name, &provisionWatcher)
	if edgexErr != nil {
		return provisionWatcher, errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("fail to query provision watcher by name %s", name), edgexErr)
	}

	return
}

// provisionWatchersByServiceName query provision watchers by offset, limit and service name
func provisionWatchersByServiceName(conn redis.Conn, offset int, limit int, name string) (provisionWatchers []models.ProvisionWatcher, edgexErr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(ProvisionWatcherCollectionServiceName, name), offset, limit)
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
	objects, err := getObjectsByRevRange(conn, CreateKey(ProvisionWatcherCollectionProfileName, name), offset, limit)
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
	objects, err := getObjectsByLabelsAndSomeRange(conn, ZREVRANGE, ProvisionWatcherCollection, labels, offset, limit)
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

// sendDeleteProvisionWatcherCmd send redis command for deleting provision watcher
func sendDeleteProvisionWatcherCmd(conn redis.Conn, storedKey string, pw models.ProvisionWatcher) {
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(HDEL, ProvisionWatcherCollectionName, pw.Name)
	_ = conn.Send(ZREM, ProvisionWatcherCollection, storedKey)
	_ = conn.Send(ZREM, CreateKey(ProvisionWatcherCollectionServiceName, pw.DiscoveredDevice.ServiceName), storedKey)
	_ = conn.Send(ZREM, CreateKey(ProvisionWatcherCollectionProfileName, pw.DiscoveredDevice.ProfileName), storedKey)
	for _, label := range pw.Labels {
		_ = conn.Send(ZREM, CreateKey(ProvisionWatcherCollectionLabel, label), storedKey)
	}
}

// deleteProvisionWatcher deletes a provision watcher
func deleteProvisionWatcher(conn redis.Conn, pw models.ProvisionWatcher) errors.EdgeX {
	storedKey := provisionWatcherStoredKey(pw.Id)
	_ = conn.Send(MULTI)
	sendDeleteProvisionWatcherCmd(conn, storedKey, pw)
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "provision watcher deletion failed", err)
	}

	return nil
}

func updateProvisionWatcher(conn redis.Conn, pw models.ProvisionWatcher) errors.EdgeX {
	exists, edgeXerr := deviceServiceNameExist(conn, pw.DiscoveredDevice.ServiceName)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("device service '%s' existence check failed", pw.DiscoveredDevice.ServiceName), edgeXerr)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exist", pw.DiscoveredDevice.ServiceName), nil)
	}
	exists, edgeXerr = deviceProfileNameExists(conn, pw.DiscoveredDevice.ProfileName)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("device profile '%s' existence check failed", pw.DiscoveredDevice.ProfileName), edgeXerr)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exist", pw.DiscoveredDevice.ProfileName), nil)
	}

	oldProvisionWatcher, edgexErr := provisionWatcherByName(conn, pw.Name)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	pw.Modified = pkgCommon.MakeTimestamp()
	storedKey := provisionWatcherStoredKey(pw.Id)
	_ = conn.Send(MULTI)
	sendDeleteProvisionWatcherCmd(conn, storedKey, oldProvisionWatcher)
	edgexErr = sendAddProvisionWatcherCmd(conn, storedKey, pw)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "provision watcher update failed", err)
	}

	return nil
}
