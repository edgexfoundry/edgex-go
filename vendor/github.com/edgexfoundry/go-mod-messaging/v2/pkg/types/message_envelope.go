//
// Copyright (c) 2019 Intel Corporation
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
)

const (
	checksum      = "payload-checksum"
	correlationId = "X-Correlation-ID"
	contentType   = "Content-Type"
)

// MessageEnvelope is the data structure for messages. It wraps the generic message payload with attributes.
type MessageEnvelope struct {
	// ReceivedTopic is the topic that the message was receive on.
	ReceivedTopic string
	// CorrelationID is an object id to identify the envelop
	CorrelationID string
	// Payload is byte representation of the data being transferred.
	Payload []byte
	// ContentType is the marshaled type of payload, i.e. application/json, application/xml, application/cbor, etc
	ContentType string
}

// NewMessageEnvelope creates a new MessageEnvelope for the specified payload with attributes from the specified context
func NewMessageEnvelope(payload []byte, ctx context.Context) MessageEnvelope {
	envelope := MessageEnvelope{
		CorrelationID: fromContext(ctx, correlationId),
		ContentType:   fromContext(ctx, contentType),
		Payload:       payload,
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
