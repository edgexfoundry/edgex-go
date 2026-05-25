//
// Copyright (c) 2022 One Track Consulting
// Copyright (c) 2022 IOTech Ltd
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

//go:build include_nats_messaging

package nats

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

type jsonMarshaller struct {
	opts ClientConfig
}

func (jm *jsonMarshaller) Marshal(v types.MessageEnvelope, publishTopic string) (*nats.Msg, error) {
	var err error

	subject := TopicToSubject(publishTopic)

	out := nats.NewMsg(subject)
	out.Data, err = json.Marshal(v)

	if jm.opts.ExactlyOnce {
		// the broker should only accept a message once per publishing service / correlation ID
		out.Header.Set(nats.MsgIdHdr, fmt.Sprintf("%s-%s", jm.opts.ClientId, v.CorrelationID))
	}

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (jm *jsonMarshaller) Unmarshal(msg *nats.Msg, target *types.MessageEnvelope) error {
	topic := subjectToTopic(msg.Subject)

	if err := json.Unmarshal(msg.Data, target); err != nil {
		return err
	}
	target.ReceivedTopic = topic
	return nil
}

const (
	contentTypeHeader   = "Content-Type"
	correlationIDHeader = "X-Correlation-ID"
	requestIDHeader     = "RequestId"
	apiVersionHeader    = "ApiVersion"
	errorCodeHeader     = "ErrorCode"
	queryParamsHeader   = "QueryParams"
)

type natsMarshaller struct {
	opts ClientConfig
}

func (nm *natsMarshaller) Marshal(v types.MessageEnvelope, publishTopic string) (*nats.Msg, error) {
	subject := TopicToSubject(publishTopic)

	out := nats.NewMsg(subject)
	payload, err := types.ConvertMsgPayloadToByteArray(v.ContentType, v.Payload)
	if err != nil {
		return nil, err
	}
	out.Data = payload
	out.Header.Set(correlationIDHeader, v.CorrelationID)
	out.Header.Set(contentTypeHeader, v.ContentType)
	out.Header.Set(requestIDHeader, v.RequestID)
	out.Header.Set(apiVersionHeader, v.ApiVersion)
	out.Header.Set(errorCodeHeader, strconv.Itoa(v.ErrorCode))
	if len(v.QueryParams) > 0 {
		for key, value := range v.QueryParams {
			query := key + ":" + value
			out.Header.Add(queryParamsHeader, query)
		}
	}
	if nm.opts.ExactlyOnce {
		// the broker should only accept a message once per publishing service / correlation ID
		out.Header.Set(nats.MsgIdHdr, fmt.Sprintf("%s-%s", nm.opts.ClientId, v.CorrelationID))
	}

	return out, nil
}

func (nm *natsMarshaller) Unmarshal(msg *nats.Msg, target *types.MessageEnvelope) error {
	topic := subjectToTopic(msg.Subject)

	target.ReceivedTopic = topic

	target.Payload = msg.Data
	target.CorrelationID = msg.Header.Get(correlationIDHeader)
	target.ContentType = msg.Header.Get(contentTypeHeader)
	target.RequestID = msg.Header.Get(requestIDHeader)
	target.ApiVersion = msg.Header.Get(apiVersionHeader)

	errorCode := msg.Header.Get(errorCodeHeader)
	if errorCode != "" {
		ec, err := strconv.Atoi(errorCode)
		if err != nil {
			return err
		}
		target.ErrorCode = ec
	}

	target.QueryParams = make(map[string]string)
	query := msg.Header.Values(queryParamsHeader)
	if len(query) > 0 {
		for _, q := range query {
			kv := strings.Split(q, ":")
			target.QueryParams[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return nil
}
