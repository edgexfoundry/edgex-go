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

// IntervalReader unmarshals a request body into an array of Interval type
type IntervalReader interface {
	ReadAddIntervalRequest(reader io.Reader) ([]dtoRequest.AddIntervalRequest, errors.EdgeX)
	ReadUpdateIntervalRequest(reader io.Reader) ([]dtoRequest.UpdateIntervalRequest, errors.EdgeX)
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewIntervalRequestReader() IntervalReader {
	return NewJsonIntervalReader()
}

// NewJsonIntervalReader creates a new instance of jsonIntervalReader
func NewJsonIntervalReader() jsonIntervalReader {
	return jsonIntervalReader{}
}

// jsonIntervalReader unmarshals the JSON request body payload
type jsonIntervalReader struct{}

// ReadAddIntervalRequest reads a request and then converts its JSON data into an array of AddIntervalRequest struct
func (jsonIntervalReader) ReadAddIntervalRequest(reader io.Reader) ([]dtoRequest.AddIntervalRequest, errors.EdgeX) {
	var addIntervals []dtoRequest.AddIntervalRequest
	err := json.NewDecoder(reader).Decode(&addIntervals)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "interval json decoding failed", err)
	}
	return addIntervals, nil
}

// ReadUpdateIntervalRequest reads a request and then converts its JSON data into an array of UpdateIntervalRequest struct
func (jsonIntervalReader) ReadUpdateIntervalRequest(reader io.Reader) ([]dtoRequest.UpdateIntervalRequest, errors.EdgeX) {
	var updateIntervals []dtoRequest.UpdateIntervalRequest
	err := json.NewDecoder(reader).Decode(&updateIntervals)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "interval json decoding failed", err)
	}
	return updateIntervals, nil
}
