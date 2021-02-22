//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"encoding/json"
	"io"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	dtoRequest "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
)

// NotificationReader unmarshals a request body into an array of Notification type
type NotificationReader interface {
	ReadAddNotificationRequest(reader io.Reader) ([]dtoRequest.AddNotificationRequest, errors.EdgeX)
}

// NewNotificationRequestReader returns a BodyReader capable of processing the request body
func NewNotificationRequestReader() NotificationReader {
	return NewJsonNotificationReader()
}

// NewJsonNotificationReader creates a new instance of jsonNotificationReader
func NewJsonNotificationReader() jsonNotificationReader {
	return jsonNotificationReader{}
}

// jsonNotificationReader unmarshals the JSON request body payload
type jsonNotificationReader struct{}

// ReadAddNotificationRequest reads a request and then converts its JSON data into an array of AddNotificationRequest struct
func (jsonNotificationReader) ReadAddNotificationRequest(reader io.Reader) ([]dtoRequest.AddNotificationRequest, errors.EdgeX) {
	var addNotifications []dtoRequest.AddNotificationRequest
	err := json.NewDecoder(reader).Decode(&addNotifications)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "notification json decoding failed", err)
	}
	return addNotifications, nil
}
