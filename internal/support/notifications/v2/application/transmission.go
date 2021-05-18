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
