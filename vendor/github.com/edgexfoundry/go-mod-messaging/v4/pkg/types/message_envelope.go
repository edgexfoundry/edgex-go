//
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2022-2025 IOTech Ltd
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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
)

const (
	envMsgBase64Payload = "EDGEX_MSG_BASE64_PAYLOAD"
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
	Payload any `json:"payload"`
	// ContentType is the marshaled type of payload, i.e. application/json, application/xml, application/cbor, etc
	ContentType string `json:"contentType"`
	// QueryParams is optionally provided key/value pairs.
	QueryParams map[string]string `json:"queryParams,omitempty"`
}

// NewMessageEnvelope creates a new MessageEnvelope for the specified payload with attributes from the specified context
func NewMessageEnvelope(payload any, ctx context.Context) MessageEnvelope {
	envelope := MessageEnvelope{
		Versionable:   commonDTO.NewVersionable(),
		CorrelationID: fromContext(ctx, common.CorrelationHeader),
		ContentType:   fromContext(ctx, common.ContentType),
		Payload:       payload,
		QueryParams:   make(map[string]string),
	}

	if IsMsgBase64Payload() || envelope.ContentType == common.ContentTypeCBOR {
		err := envelope.ConvertMsgPayloadToByteArray()
		if err != nil {
			fmt.Println("convert message payload to bytes failed, err: " + err.Error())
		}
	}
	return envelope
}

// NewMessageEnvelopeForRequest creates a new MessageEnvelope for sending request to EdgeX via internal
// MessageBus to target Device Service. Used when request is from internal App Service via command client.
func NewMessageEnvelopeForRequest(payload any, queryParams map[string]string) MessageEnvelope {
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

	if IsMsgBase64Payload() || envelope.ContentType == common.ContentTypeCBOR {
		err := envelope.ConvertMsgPayloadToByteArray()
		if err != nil {
			fmt.Println("convert message payload to bytes failed, err: " + err.Error())
		}
	}
	return envelope
}

// NewMessageEnvelopeForResponse creates a new MessageEnvelope for sending response from Device Service back to Core Command.
func NewMessageEnvelopeForResponse(payload any, requestId string, correlationId string, contentType string) (MessageEnvelope, error) {
	var err error
	if _, err = uuid.Parse(requestId); err != nil {
		return MessageEnvelope{}, err
	}
	if _, err = uuid.Parse(correlationId); err != nil {
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

	if IsMsgBase64Payload() || envelope.ContentType == common.ContentTypeCBOR {
		err = envelope.ConvertMsgPayloadToByteArray()
		if err != nil {
			return MessageEnvelope{}, fmt.Errorf("failed to convert payload to []byte: %w", err)
		}
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
	envelope := MessageEnvelope{
		CorrelationID: uuid.NewString(),
		Versionable:   commonDTO.NewVersionable(),
		RequestID:     requestId,
		ErrorCode:     1,
		Payload:       errorMessage,
		ContentType:   common.ContentTypeText,
		QueryParams:   make(map[string]string),
	}

	return envelope
}

func fromContext(ctx context.Context, key string) string {
	hdr, ok := ctx.Value(key).(string)
	if !ok {
		hdr = ""
	}
	return hdr
}

func IsMsgBase64Payload() bool {
	return os.Getenv(envMsgBase64Payload) == common.ValueTrue
}

// ConvertMsgPayloadToByteArray converts the MessageEnvelope's payload to a byte array.
func (msg *MessageEnvelope) ConvertMsgPayloadToByteArray() error {
	payload, err := ConvertMsgPayloadToByteArray(msg.ContentType, msg.Payload)
	if err != nil {
		return err
	}
	msg.Payload = payload
	return nil
}

func ConvertMsgPayloadToByteArray(contentType string, payload any) (result []byte, err error) {
	switch v := payload.(type) {
	case []byte:
		result = v
	case string:
		result = []byte(v)
	default:
		result, err = marshalMsgPayload(contentType, payload)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// GetMsgPayload handles different payload types and attempts to convert them to the desired type T.
func GetMsgPayload[T any](msg MessageEnvelope) (res T, err error) {
	unmarshalErrStr := "failed to unmarshal to %T, error: %w"
	switch v := msg.Payload.(type) {
	case T:
		res = v
	case []byte:
		err = unmarshalMsgPayload(msg.ContentType, v, &res)
		if err != nil {
			return res, fmt.Errorf(unmarshalErrStr, res, err)
		}
	case string:
		// Check if payload is base64 string
		decodeValue, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return res, fmt.Errorf("failed to decode base64 string to %T: %w", res, err)
		}
		// If T is []byte
		if reflect.TypeOf(res).String() == reflect.TypeOf(decodeValue).String() {
			return any(decodeValue).(T), nil
		}

		err = unmarshalMsgPayload(msg.ContentType, decodeValue, &res)
		if err != nil {
			return res, fmt.Errorf(unmarshalErrStr, res, err)
		}
	default:
		bytes, err := marshalMsgPayload(msg.ContentType, v)
		if err != nil {
			return res, fmt.Errorf("failed to marshal to []byte: %w", err)
		}
		err = unmarshalMsgPayload(msg.ContentType, bytes, &res)
		if err != nil {
			return res, fmt.Errorf(unmarshalErrStr, res, err)
		}
	}

	return res, nil
}

func marshalMsgPayload(contentType string, payload any) (bytes []byte, err error) {
	// no need to marshal nil payload
	if payload == nil {
		return nil, nil
	}
	switch contentType {
	case common.ContentTypeJSON:
		bytes, err = json.Marshal(payload)
		if err != nil {
			return bytes, fmt.Errorf("failed to marshal to JSON, error: %w", err)
		}
	case common.ContentTypeCBOR:
		bytes, err = cbor.Marshal(payload)
		if err != nil {
			return bytes, fmt.Errorf("failed to marshal to CBOR, error: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	return bytes, nil
}

func unmarshalMsgPayload(contentType string, payload []byte, result any) (err error) {
	switch contentType {
	case common.ContentTypeJSON:
		err = json.Unmarshal(payload, result)
		if err != nil {
			return err
		}
	case common.ContentTypeCBOR:
		err = cbor.Unmarshal(payload, result)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
	return nil
}
