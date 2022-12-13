//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/google/uuid"
)

// TransmissionById invokes the infrastructure layer function to query transmission by ID
func TransmissionById(id string, dic *di.Container) (trans dtos.Transmission, edgeXerr errors.EdgeX) {
	if id == "" {
		return trans, errors.NewCommonEdgeX(errors.KindContractInvalid, "ID is empty", nil)
	}
	if _, err := uuid.Parse(id); err != nil {
		return trans, errors.NewCommonEdgeX(errors.KindContractInvalid, "ID is not a valid UUID", err)
	}

	dbClient := container.DBClientFrom(dic.Get)
	transModel, edgeXerr := dbClient.TransmissionById(id)
	if edgeXerr != nil {
		return trans, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	trans = dtos.FromTransmissionModelToDTO(transModel)
	return trans, nil
}

// TransmissionsByTimeRange query transmissions with offset, limit and time range
func TransmissionsByTimeRange(start int, end int, offset int, limit int, dic *di.Container) (transmissions []dtos.Transmission, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	models, err := dbClient.TransmissionsByTimeRange(start, end, offset, limit)
	if err == nil {
		totalCount, err = dbClient.TransmissionCountByTimeRange(start, end)
	}
	if err != nil {
		return transmissions, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	transmissions = make([]dtos.Transmission, len(models))
	for i, trans := range models {
		transmissions[i] = dtos.FromTransmissionModelToDTO(trans)
	}
	return transmissions, totalCount, nil
}

// AllTransmissions queries transmissions by offset and limit
func AllTransmissions(offset, limit int, dic *di.Container) (transmissions []dtos.Transmission, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	models, err := dbClient.AllTransmissions(offset, limit)
	if err == nil {
		totalCount, err = dbClient.TransmissionTotalCount()
	}
	if err != nil {
		return transmissions, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	transmissions = make([]dtos.Transmission, len(models))
	for i, trans := range models {
		transmissions[i] = dtos.FromTransmissionModelToDTO(trans)
	}
	return transmissions, totalCount, nil
}

// TransmissionsByStatus queries transmissions with offset, limit, and status
func TransmissionsByStatus(offset, limit int, status string, dic *di.Container) (transmissions []dtos.Transmission, totalCount uint32, err errors.EdgeX) {
	if status == "" {
		return transmissions, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "status is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	transModels, err := dbClient.TransmissionsByStatus(offset, limit, status)
	if err == nil {
		totalCount, err = dbClient.TransmissionCountByStatus(status)
	}
	if err != nil {
		return transmissions, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromTransmissionModelsToDTOs(transModels), totalCount, nil
}

// DeleteProcessedTransmissionsByAge invokes the infrastructure layer function to remove the processed transmissions that are older than age.
// Age is supposed in milliseconds since created timestamp.
func DeleteProcessedTransmissionsByAge(age int64, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)

	err := dbClient.DeleteProcessedTransmissionsByAge(age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// TransmissionsBySubscriptionName queries transmissions with offset, limit, and subscription name
func TransmissionsBySubscriptionName(offset, limit int, subscriptionName string, dic *di.Container) (transmissions []dtos.Transmission, totalCount uint32, err errors.EdgeX) {
	if subscriptionName == "" {
		return transmissions, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "subscription name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	transModels, err := dbClient.TransmissionsBySubscriptionName(offset, limit, subscriptionName)
	if err == nil {
		totalCount, err = dbClient.TransmissionCountBySubscriptionName(subscriptionName)
	}
	if err != nil {
		return transmissions, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromTransmissionModelsToDTOs(transModels), totalCount, nil
}

// TransmissionsByNotificationId queries transmissions with offset, limit, and notification id
func TransmissionsByNotificationId(offset, limit int, notificationId string, dic *di.Container) (transmissions []dtos.Transmission, totalCount uint32, err errors.EdgeX) {
	if notificationId == "" {
		return transmissions, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "notification id is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	transModels, err := dbClient.TransmissionsByNotificationId(offset, limit, notificationId)
	if err == nil {
		totalCount, err = dbClient.TransmissionCountByNotificationId(notificationId)
	}
	if err != nil {
		return transmissions, totalCount, errors.NewCommonEdgeXWrapper(err)
	} else {
		return dtos.FromTransmissionModelsToDTOs(transModels), totalCount, nil
	}
}
