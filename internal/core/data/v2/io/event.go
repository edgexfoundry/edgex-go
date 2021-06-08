//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	dto "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/fxamacker/cbor/v2"
)

// To avoid large data causing unexpected memory exhaustion when decoding CBOR payload, defaultMaxEventSize was introduced as
// a reasonable limit appropriate for handling CBOR payload in edgex-go.  More details could be found at
// https://github.com/edgexfoundry/edgex-go/issues/2439
// TODO Make MaxEventSize a service configuration setting, so that users could adjust the limit per systems requirements
// https://github.com/edgexfoundry/edgex-go/issues/3237
const defaultMaxEventSize = int64(25 * 1e6) // 25 MB

// EventReader unmarshals a request body into an Event type
type EventReader interface {
	ReadAddEventRequest(bytes []byte) (dto.AddEventRequest, errors.EdgeX)
}

// NewRequestReader returns a BodyReader capable of processing the request body
func NewEventRequestReader(contentType string) EventReader {
	switch strings.ToLower(contentType) {
	case clients.ContentTypeCBOR:
		return NewCborReader()
	default:
		return NewJsonReader()
	}
}

// cborEventReader handles unmarshaling of a CBOR encoded request body payload
type cborEventReader struct{}

// NewCborReader creates a new instance of cborEventReader.
func NewCborReader() cborEventReader {
	return cborEventReader{}
}

// ReadDataInBytes reads and converts the request's CBOR encoded add event request into an AddEventRequest struct
func (cborEventReader) ReadAddEventRequest(bytes []byte) (dto.AddEventRequest, errors.EdgeX) {
	var addEvent dto.AddEventRequest
	err := cbor.Unmarshal(bytes, &addEvent)
	if err != nil {
		return addEvent, errors.NewCommonEdgeX(errors.KindContractInvalid, "cbor AddEventRequest decoding failed", err)
	}

	return addEvent, nil
}

// jsonReader handles unmarshaling of a JSON request body payload
type jsonEventReader struct{}

// NewJsonReader creates a new instance of jsonReader.
func NewJsonReader() jsonEventReader {
	return jsonEventReader{}
}

// ReadDataInBytes reads and converts the request's JSON encoded add event request into an AddEventRequest struct
func (jsonEventReader) ReadAddEventRequest(bytes []byte) (dto.AddEventRequest, errors.EdgeX) {
	var addEvent dto.AddEventRequest
	err := json.Unmarshal(bytes, &addEvent)
	if err != nil {
		return addEvent, errors.NewCommonEdgeX(errors.KindContractInvalid, "event json decoding failed", err)
	}
	return addEvent, nil
}

func ReadDataInBytes(reader io.Reader) ([]byte, errors.EdgeX) {
	// use LimitReader with defaultMaxEventSize to avoid unexpected memory exhaustion
	bytes, err := io.ReadAll(io.LimitReader(reader, defaultMaxEventSize))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindIOError, "AddEventRequest I/O reading failed", err)
	}
	return bytes, nil
}
