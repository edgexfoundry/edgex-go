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
	dto "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
)

// EventReader unmarshals a request body into an Event type
type EventReader interface {
	ReadAddEventRequest(reader io.Reader, ctx *context.Context) ([]dto.AddEventRequest, errors.EdgeX)
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewEventRequestReader() EventReader {
	return NewJsonReader()
}

// jsonReader handles unmarshaling of a JSON request body payload
type jsonEventReader struct{}

// NewJsonReader creates a new instance of jsonReader.
func NewJsonReader() jsonEventReader {
	return jsonEventReader{}
}

// Read reads and converts the request's JSON event data into an Event struct
func (jsonEventReader) ReadAddEventRequest(reader io.Reader, ctx *context.Context) ([]dto.AddEventRequest, errors.EdgeX) {
	c := context.WithValue(*ctx, clients.ContentType, clients.ContentTypeJSON)
	*ctx = c
	var addEvents []dto.AddEventRequest
	err := json.NewDecoder(reader).Decode(&addEvents)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "event json decoding failed", err)
	}
	return addEvents, nil
}
