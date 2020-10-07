//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"context"
	"encoding/json"
	"io"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	dtoRequest "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

// DeviceServiceReader unmarshals a request body into an array of DeviceService type
type DeviceServiceReader interface {
	ReadAddDeviceServiceRequest(reader io.Reader, ctx *context.Context) ([]dtoRequest.AddDeviceServiceRequest, errors.EdgeX)
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewDeviceServiceRequestReader() DeviceServiceReader {
	return NewJsonDeviceServiceReader()
}

// NewJsonDeviceServiceReader creates a new instance of jsonDeviceServiceReader
func NewJsonDeviceServiceReader() jsonDeviceServiceReader {
	return jsonDeviceServiceReader{}
}

// jsonDeviceServiceReader unmarshals the JSON request body payload
type jsonDeviceServiceReader struct{}

// ReadAddDeviceServiceRequest reads a request and then converts its JSON data into an array of AddDeviceServiceRequest struct
func (jsonDeviceServiceReader) ReadAddDeviceServiceRequest(reader io.Reader, ctx *context.Context) ([]dtoRequest.AddDeviceServiceRequest, errors.EdgeX) {
	c := context.WithValue(*ctx, clients.ContentType, clients.ContentTypeJSON)
	*ctx = c
	var addDeviceServices []dtoRequest.AddDeviceServiceRequest
	err := json.NewDecoder(reader).Decode(&addDeviceServices)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "device profile json decoding failed", err)
	}
	return addDeviceServices, nil
}
