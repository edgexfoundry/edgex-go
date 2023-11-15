//
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2022-2023 IOTech Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	"github.com/google/uuid"
)

// MessageEnvelope is the data structure for messages. It wraps the generic message payload with attributes.
type MessageEnvelope struct {
	// ApiVersion (from Versionable) shows the API version for the message envelope.
	commonDTO.Versionable
	// ReceivedTopic is the topic that the message was received on.
	ReceivedTopic string `json:"receivedTopic"`
	// CorrelationID is an object id to identify the envelope.
	CorrelationID string `json:"correlationID"`
	// RequestID is an object id to identify the request.
	RequestID string `json:"requestID"`
	// ErrorCode provides the indication of error. '0' indicates no error, '1' indicates error.
	// Additional codes may be added in the future. If non-0, the payload will contain the error.
	ErrorCode int `json:"errorCode"`
	// Payload is byte representation of the data being transferred.
	Payload []byte `json:"payload"`
	// ContentType is the marshaled type of payload, i.e. application/json, application/xml, application/cbor, etc
	ContentType string `json:"contentType"`
	// QueryParams is optionally provided key/value pairs.
	QueryParams map[string]string `json:"queryParams,omitempty"`
}

// NewMessageEnvelope creates a new MessageEnvelope for the specified payload with attributes from the specified context
func NewMessageEnvelope(payload []byte, ctx context.Context) MessageEnvelope {
	envelope := MessageEnvelope{
		Versionable:   commonDTO.NewVersionable(),
		CorrelationID: fromContext(ctx, common.CorrelationHeader),
		ContentType:   fromContext(ctx, common.ContentType),
		Payload:       payload,
		QueryParams:   make(map[string]string),
	}

	return envelope
}

// NewMessageEnvelopeForRequest creates a new MessageEnvelope for sending request to EdgeX via internal
// MessageBus to target Device Service. Used when request is from internal App Service via command client.
func NewMessageEnvelopeForRequest(payload []byte, queryParams map[string]string) MessageEnvelope {
	envelope := MessageEnvelope{
		CorrelationID: uuid.NewString(),
		Versionable:   commonDTO.NewVersionable(),
		RequestID:     uuid.NewString(),
		ErrorCode:     0,
		Payload:       payload,
		ContentType:   common.ContentTypeJSON,
		QueryParams:   make(map[string]string),
	}

	if len(queryParams) > 0 {
		envelope.QueryParams = queryParams
	}

	return envelope
}

// NewMessageEnvelopeForResponse creates a new MessageEnvelope for sending response from Device Service back to Core Command.
func NewMessageEnvelopeForResponse(payload []byte, requestId string, correlationId string, contentType string) (MessageEnvelope, error) {
	if _, err := uuid.Parse(requestId); err != nil {
		return MessageEnvelope{}, err
	}
	if _, err := uuid.Parse(correlationId); err != nil {
		return MessageEnvelope{}, err
	}
	if contentType == "" {
		return MessageEnvelope{}, errors.New("ContentType is empty")
	}

	envelope := MessageEnvelope{
		CorrelationID: correlationId,
		Versionable:   commonDTO.NewVersionable(),
		RequestID:     requestId,
		ErrorCode:     0,
		Payload:       payload,
		ContentType:   contentType,
		QueryParams:   make(map[string]string),
	}

	return envelope, nil
}

// NewMessageEnvelopeFromJSON creates a new MessageEnvelope by decoding the message payload
// received from external MQTT in order to send request via internal MessageBus.
func NewMessageEnvelopeFromJSON(message []byte) (MessageEnvelope, error) {
	var envelope MessageEnvelope
	err := json.Unmarshal(message, &envelope)
	if err != nil {
		return MessageEnvelope{}, err
	}

	if envelope.ApiVersion != common.ApiVersion {
		return MessageEnvelope{}, fmt.Errorf("api version '%s' is required", common.ApiVersion)
	}

	if _, err = uuid.Parse(envelope.RequestID); err != nil {
		return MessageEnvelope{}, fmt.Errorf("error parsing RequestID: %s", err.Error())
	}

	if _, err = uuid.Parse(envelope.CorrelationID); err != nil {
		if envelope.CorrelationID != "" {
			return MessageEnvelope{}, fmt.Errorf("error parsing CorrelationID: %s", err.Error())
		}

		envelope.CorrelationID = uuid.NewString()
	}

	if envelope.ContentType != common.ContentTypeJSON {
		return envelope, fmt.Errorf("ContentType is not %s", common.ContentTypeJSON)
	}

	if envelope.QueryParams == nil {
		envelope.QueryParams = make(map[string]string)
	}

	return envelope, nil
}

// NewMessageEnvelopeWithError creates a new MessageEnvelope with ErrorCode set to 1 indicating there's error
// and the payload contains message string about the error.
func NewMessageEnvelopeWithError(requestId string, errorMessage string) MessageEnvelope {
	return MessageEnvelope{
		CorrelationID: uuid.NewString(),
		Versionable:   commonDTO.NewVersionable(),
		RequestID:     requestId,
		ErrorCode:     1,
		Payload:       []byte(errorMessage),
		ContentType:   common.ContentTypeText,
		QueryParams:   make(map[string]string),
	}
}

func fromContext(ctx context.Context, key string) string {
	hdr, ok := ctx.Value(key).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}
