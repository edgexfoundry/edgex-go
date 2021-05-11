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
