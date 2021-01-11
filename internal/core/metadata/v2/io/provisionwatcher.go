//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"encoding/json"
	"io"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	dtoRequest "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

// ProvisionWatcherReader unmarshals a request body into an array of ProvisionWatcher type
type ProvisionWatcherReader interface {
	ReadAddProvisionWatcherRequest(reader io.Reader) ([]dtoRequest.AddProvisionWatcherRequest, errors.EdgeX)
	ReadUpdateProvisionWatcherRequest(reader io.Reader) ([]dtoRequest.UpdateProvisionWatcherRequest, errors.EdgeX)
}

// NewProvisionWatcherRequestReader returns a BodyReader capable of processing the request body
func NewProvisionWatcherRequestReader() ProvisionWatcherReader {
	return NewJsonProvisionWatcherReader()
}

// NewJsonProvisionWatcherReader creates a new instance of jsonProvisionWatcherReader
func NewJsonProvisionWatcherReader() jsonProvisionWatcherReader {
	return jsonProvisionWatcherReader{}
}

// jsonProvisionWatcherReader unmarshals the JSON request body payload
type jsonProvisionWatcherReader struct{}

// ReadAddProvisionWatcherRequest reads a request and then converts its JSON data into an array of AddProvisionWatcherRequest struct
func (jsonProvisionWatcherReader) ReadAddProvisionWatcherRequest(reader io.Reader) ([]dtoRequest.AddProvisionWatcherRequest, errors.EdgeX) {
	var addProvisionWatchers []dtoRequest.AddProvisionWatcherRequest
	err := json.NewDecoder(reader).Decode(&addProvisionWatchers)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "provision watcher json decoding failed", err)
	}

	return addProvisionWatchers, nil
}

// ReadUpdateProvisionWatcherRequest reads a request and then converts its JSON data into an array of UpdateProvisionWatcherRequest struct
func (jsonProvisionWatcherReader) ReadUpdateProvisionWatcherRequest(reader io.Reader) ([]dtoRequest.UpdateProvisionWatcherRequest, errors.EdgeX) {
	var updateProvisionWatchers []dtoRequest.UpdateProvisionWatcherRequest
	err := json.NewDecoder(reader).Decode(&updateProvisionWatchers)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "provision watcher json decoding failed", err)
	}

	return updateProvisionWatchers, nil
}
