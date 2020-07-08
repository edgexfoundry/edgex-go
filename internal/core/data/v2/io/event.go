//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"context"
	"encoding/json"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	dto "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"io"
)

// EventReader unmarshals a request body into an Event type
type EventReader interface {
	ReadAddEventRequest(reader io.Reader, ctx *context.Context) ([]dto.AddEventRequest, error)
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
func (jsonEventReader) ReadAddEventRequest(reader io.Reader, ctx *context.Context) (events []dto.AddEventRequest, err error) {
	c := context.WithValue(*ctx, clients.ContentType, clients.ContentTypeJSON)
	*ctx = c
	var addEvents []dto.AddEventRequest
	err = json.NewDecoder(reader).Decode(&addEvents)
	if err != nil {
		return nil, err
	}
	return addEvents, nil
}
