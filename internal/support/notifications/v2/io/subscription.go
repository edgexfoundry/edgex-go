//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"encoding/json"
	"io"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	dtoRequest "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
)

// SubscriptionReader unmarshals a request body into an array of Subscription type
type SubscriptionReader interface {
	ReadAddSubscriptionRequest(reader io.Reader) ([]dtoRequest.AddSubscriptionRequest, errors.EdgeX)
	ReadUpdateSubscriptionRequest(reader io.Reader) ([]dtoRequest.UpdateSubscriptionRequest, errors.EdgeX)
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewSubscriptionRequestReader() SubscriptionReader {
	return NewJsonSubscriptionReader()
}

// NewJsonSubscriptionReader creates a new instance of jsonSubscriptionReader
func NewJsonSubscriptionReader() jsonSubscriptionReader {
	return jsonSubscriptionReader{}
}

// jsonSubscriptionReader unmarshals the JSON request body payload
type jsonSubscriptionReader struct{}

// ReadAddSubscriptionRequest reads a request and then converts its JSON data into an array of AddSubscriptionRequest struct
func (jsonSubscriptionReader) ReadAddSubscriptionRequest(reader io.Reader) ([]dtoRequest.AddSubscriptionRequest, errors.EdgeX) {
	var addSubscriptions []dtoRequest.AddSubscriptionRequest
	err := json.NewDecoder(reader).Decode(&addSubscriptions)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "subscription json decoding failed", err)
	}
	return addSubscriptions, nil
}

// ReadUpdateSubscriptionRequest reads a request and then converts its JSON data into an array of UpdateSubscriptionRequest struct
func (jsonSubscriptionReader) ReadUpdateSubscriptionRequest(reader io.Reader) ([]dtoRequest.UpdateSubscriptionRequest, errors.EdgeX) {
	var updateSubscriptions []dtoRequest.UpdateSubscriptionRequest
	err := json.NewDecoder(reader).Decode(&updateSubscriptions)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "subscription json decoding failed", err)
	}

	return updateSubscriptions, nil
}
