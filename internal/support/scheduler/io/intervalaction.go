//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"encoding/json"
	"io"

	dtoRequest "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// IntervalActionReader unmarshals a request body into an array of IntervalAction type
type IntervalActionReader interface {
	ReadAddIntervalActionRequest(reader io.Reader) ([]dtoRequest.AddIntervalActionRequest, errors.EdgeX)
	ReadUpdateIntervalActionRequest(reader io.Reader) ([]dtoRequest.UpdateIntervalActionRequest, errors.EdgeX)
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewIntervalActionRequestReader() IntervalActionReader {
	return NewJsonIntervalActionReader()
}

// NewJsonIntervalActionReader creates a new instance of jsonIntervalActionReader
func NewJsonIntervalActionReader() jsonIntervalActionReader {
	return jsonIntervalActionReader{}
}

// jsonIntervalActionReader unmarshals the JSON request body payload
type jsonIntervalActionReader struct{}

// ReadAddIntervalActionRequest reads a request and then converts its JSON data into an array of AddIntervalActionRequest struct
func (jsonIntervalActionReader) ReadAddIntervalActionRequest(reader io.Reader) ([]dtoRequest.AddIntervalActionRequest, errors.EdgeX) {
	var addIntervalActions []dtoRequest.AddIntervalActionRequest
	err := json.NewDecoder(reader).Decode(&addIntervalActions)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "intervalAction json decoding failed", err)
	}
	return addIntervalActions, nil
}

// ReadUpdateIntervalActionRequest reads a request and then converts its JSON data into an array of UpdateIntervalActionRequest struct
func (jsonIntervalActionReader) ReadUpdateIntervalActionRequest(reader io.Reader) ([]dtoRequest.UpdateIntervalActionRequest, errors.EdgeX) {
	var updateIntervalActions []dtoRequest.UpdateIntervalActionRequest
	err := json.NewDecoder(reader).Decode(&updateIntervalActions)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "intervalAction json decoding failed", err)
	}
	return updateIntervalActions, nil
}
