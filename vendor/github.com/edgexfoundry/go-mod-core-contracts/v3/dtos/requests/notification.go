//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// AddNotificationRequest defines the Request Content for POST Notification DTO.
type AddNotificationRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	Notification          dtos.Notification `json:"notification"`
}

// Validate satisfies the Validator interface
func (request AddNotificationRequest) Validate() error {
	err := common.Validate(request)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the AddNotificationRequest type
func (request *AddNotificationRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		Notification dtos.Notification
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*request = AddNotificationRequest(alias)

	// validate AddNotificationRequest DTO
	if err := request.Validate(); err != nil {
		return err
	}
	return nil
}

// AddNotificationReqToNotificationModels transforms the AddNotificationRequest DTO array to the AddNotificationRequest model array
func AddNotificationReqToNotificationModels(reqs []AddNotificationRequest) (n []models.Notification) {
	for _, req := range reqs {
		d := dtos.ToNotificationModel(req.Notification)
		n = append(n, d)
	}
	return n
}

func NewAddNotificationRequest(dto dtos.Notification) AddNotificationRequest {
	return AddNotificationRequest{
		BaseRequest:  dtoCommon.NewBaseRequest(),
		Notification: dto,
	}
}
