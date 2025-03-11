//
// Copyright (c) 2023 Intel Corporation
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

package messaging

import (
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// MessageClient is the messaging interface for publisher-subscriber pattern
type MessageClient interface {
	// Connect to messaging host specified in MessageBus config
	// returns error if not able to connect
	Connect() error

	// Publish is to send message to the message bus
	// the message contains data payload to send to the message bus
	Publish(message types.MessageEnvelope, topic string) error

	// PublishWithSizeLimit checks the message size in kilobytes after marshall and send it to the message bus
	// the message contains data payload to send to the message bus
	PublishWithSizeLimit(message types.MessageEnvelope, topic string, limit int64) error

	// Subscribe is to receive messages from topic channels
	// if message does not require a topic, then use empty string ("") for topic
	// the topic channel contains subscribed message channel and topic to associate with it
	// the channel is used for multiple threads of subscribers for 1 publisher (1-to-many)
	// the messageErrors channel returns the message errors from the caller
	// since subscriber works in asynchronous fashion
	// the function returns error for any subscribe error
	Subscribe(topics []types.TopicChannel, messageErrors chan error) error

	// Request publishes a request containing a RequestID to the specified topic,
	// then subscribes to a response topic which contains the RequestID. Once the response is received, the
	// response topic is unsubscribed and the response data is returned. If no response is received within
	// the timeout period, a timed out  error returned.
	Request(message types.MessageEnvelope, requestTopic string, responseTopicPrefix string, timeout time.Duration) (*types.MessageEnvelope, error)

	// PublishBinaryData sends binary data to the message bus
	PublishBinaryData(data []byte, topic string) error

	// SubscribeBinaryData receives binary data from the specified topic, and wrap it in MessageEnvelope.
	SubscribeBinaryData(topics []types.TopicChannel, messageErrors chan error) error

	// Unsubscribe to unsubscribe from the specified topics.
	Unsubscribe(topics ...string) error

	// Disconnect is to close all connections on the message bus
	// and TopicChannel will also be closed
	Disconnect() error
}
