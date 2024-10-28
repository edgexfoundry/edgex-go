//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/gomodule/redigo/redis"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"
)

const RegistrationCollection = "kp|r"
const RegistrationCollectionName = RegistrationCollection + DBKeySeparator + common.Name

func registrationStoredKey(id string) string {
	return CreateKey(RegistrationCollection, id)
}

func addRegistration(conn redis.Conn, r models.Registration) (models.Registration, errors.EdgeX) {
	exists, edgeXerr := serviceIdExists(conn, r.ServiceId)
	if edgeXerr != nil {
		return models.Registration{}, errors.NewCommonEdgeXWrapper(edgeXerr)
	} else if exists {
		return models.Registration{}, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("serviceId %s already exists", r.ServiceId), edgeXerr)
	}

	ts := pkgCommon.MakeTimestamp()
	if r.Created == 0 {
		r.Created = ts
	}
	r.Modified = ts

	storedKey := registrationStoredKey(r.ServiceId)
	_ = conn.Send(MULTI)

	edgexErr := sendAddRegistrationCmd(conn, storedKey, r)
	if edgexErr != nil {
		return models.Registration{}, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return models.Registration{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "registration creation failed", err)
	}

	return r, nil
}

func deleteRegistrationByServiceId(conn redis.Conn, id string) errors.EdgeX {
	r, err := registrationById(conn, id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	err = deleteRegistration(conn, r)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

func serviceIdExists(conn redis.Conn, id string) (bool, errors.EdgeX) {
	exists, err := objectIdExists(conn, registrationStoredKey(id))
	if err != nil {
		return false, errors.NewCommonEdgeX(errors.KindDatabaseError, "registration existence check by serviceId failed", err)
	}

	return exists, nil
}

func registrationById(conn redis.Conn, id string) (registration models.Registration, edgexErr errors.EdgeX) {
	edgexErr = getObjectById(conn, registrationStoredKey(id), &registration)
	if edgexErr != nil {
		return registration, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	return
}

func registrations(conn redis.Conn, start, end int) ([]models.Registration, errors.EdgeX) {
	objects, edgeXerr := getObjectsByRange(conn, RegistrationCollection, start, end)
	if edgeXerr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	registries := make([]models.Registration, len(objects))
	for i, in := range objects {
		r := models.Registration{}
		err := json.Unmarshal(in, &r)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "device format parsing failed from the database", err)
		}
		registries[i] = r
	}
	return registries, nil
}

func deleteRegistration(conn redis.Conn, r models.Registration) errors.EdgeX {
	storedKey := registrationStoredKey(r.ServiceId)
	_ = conn.Send(MULTI)
	sendDeleteRegistrationCmd(conn, storedKey, r)
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "registration deletion failed", err)
	}
	return nil
}

func updateRegistration(conn redis.Conn, r models.Registration) errors.EdgeX {
	oldRegistration, edgexErr := registrationById(conn, r.ServiceId)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	if r.Created == 0 {
		r.Created = oldRegistration.Created
	}
	r.Modified = pkgCommon.MakeTimestamp()
	storedKey := registrationStoredKey(r.ServiceId)
	_ = conn.Send(MULTI)
	sendDeleteRegistrationCmd(conn, storedKey, oldRegistration)
	edgexErr = sendAddRegistrationCmd(conn, storedKey, r)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "registration update failed", err)
	}

	return nil
}

func sendAddRegistrationCmd(conn redis.Conn, storedKey string, r models.Registration) errors.EdgeX {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal registration for Redis persistence", err)
	}

	_ = conn.Send(SET, storedKey, jsonBytes)
	_ = conn.Send(ZADD, RegistrationCollection, 0, storedKey)
	_ = conn.Send(HSET, RegistrationCollectionName, r.ServiceId, storedKey)
	return nil
}

func sendDeleteRegistrationCmd(conn redis.Conn, storedKey string, r models.Registration) {
	_ = conn.Send(DEL, storedKey)
	_ = conn.Send(ZREM, RegistrationCollection, storedKey)
	_ = conn.Send(HDEL, RegistrationCollectionName, r.ServiceId)
}
