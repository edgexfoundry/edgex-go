//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"encoding/json"
	"io"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	dtoRequest "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
)

// DeviceReader unmarshals a request body into an array of Device type
type DeviceReader interface {
	ReadAddDeviceRequest(reader io.Reader) ([]dtoRequest.AddDeviceRequest, errors.EdgeX)
	ReadUpdateDeviceRequest(reader io.Reader) ([]dtoRequest.UpdateDeviceRequest, errors.EdgeX)
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewDeviceRequestReader() DeviceReader {
	return NewJsonDeviceReader()
}

// NewJsonDeviceReader creates a new instance of jsonDeviceReader
func NewJsonDeviceReader() jsonDeviceReader {
	return jsonDeviceReader{}
}

// jsonDeviceReader unmarshals the JSON request body payload
type jsonDeviceReader struct{}

// ReadAddDeviceRequest reads a request and then converts its JSON data into an array of AddDeviceRequest struct
func (jsonDeviceReader) ReadAddDeviceRequest(reader io.Reader) ([]dtoRequest.AddDeviceRequest, errors.EdgeX) {
	var addDevices []dtoRequest.AddDeviceRequest
	err := json.NewDecoder(reader).Decode(&addDevices)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "device json decoding failed", err)
	}
	return addDevices, nil
}

// ReadUpdateDeviceRequest reads a request and then converts its JSON data into an array of UpdateDeviceRequest struct
func (jsonDeviceReader) ReadUpdateDeviceRequest(reader io.Reader) ([]dtoRequest.UpdateDeviceRequest, errors.EdgeX) {
	var updateDevices []dtoRequest.UpdateDeviceRequest
	err := json.NewDecoder(reader).Decode(&updateDevices)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "device json decoding failed", err)
	}
	return updateDevices, nil
}
