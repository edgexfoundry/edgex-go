//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"

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

	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	transModel, edgeXerr := dbClient.TransmissionById(id)
	if edgeXerr != nil {
		return trans, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	trans = dtos.FromTransmissionModelToDTO(transModel)
	return trans, nil
}

// TransmissionsByTimeRange query transmissions with offset, limit and time range
func TransmissionsByTimeRange(start int, end int, offset int, limit int, dic *di.Container) (transmissions []dtos.Transmission, err errors.EdgeX) {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	models, err := dbClient.TransmissionsByTimeRange(start, end, offset, limit)
	if err != nil {
		return transmissions, errors.NewCommonEdgeXWrapper(err)
	}
	transmissions = make([]dtos.Transmission, len(models))
	for i, trans := range models {
		transmissions[i] = dtos.FromTransmissionModelToDTO(trans)
	}
	return transmissions, nil
}

// AllTransmissions queries transmissions by offset and limit
func AllTransmissions(offset, limit int, dic *di.Container) (transmissions []dtos.Transmission, err errors.EdgeX) {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	models, err := dbClient.AllTransmissions(offset, limit)
	if err != nil {
		return transmissions, errors.NewCommonEdgeXWrapper(err)
	}
	transmissions = make([]dtos.Transmission, len(models))
	for i, trans := range models {
		transmissions[i] = dtos.FromTransmissionModelToDTO(trans)
	}
	return transmissions, nil
}

// TransmissionsByStatus queries transmissions with offset, limit, and status
func TransmissionsByStatus(offset, limit int, status string, dic *di.Container) (transmissions []dtos.Transmission, err errors.EdgeX) {
	if status == "" {
		return transmissions, errors.NewCommonEdgeX(errors.KindContractInvalid, "status is empty", nil)
	}
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	transModels, err := dbClient.TransmissionsByStatus(offset, limit, status)
	if err != nil {
		return transmissions, errors.NewCommonEdgeXWrapper(err)
	}
	return dtos.FromTransmissionModelsToDTOs(transModels), nil
}

// DeleteProcessedTransmissionsByAge invokes the infrastructure layer function to remove the processed transmissions that are older than age.
// Age is supposed in milliseconds since created timestamp.
func DeleteProcessedTransmissionsByAge(age int64, dic *di.Container) errors.EdgeX {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)

	err := dbClient.DeleteProcessedTransmissionsByAge(age)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}
