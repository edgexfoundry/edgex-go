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

package messaging

import (
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

// MessageClient is the messaging interface for publisher-subscriber pattern
type MessageClient interface {
	// Connect to messaging host specified in MessageBus config
	// returns error if not able to connect
	Connect() error

	// Publish is to send message to the message bus
	// the message contains data payload to send to the message queue
	Publish(message types.MessageEnvelope, topic string) error

	// Subscribe is to receive messages from topic channels
	// if message does not require a topic, then use empty string ("") for topic
	// the topic channel contains subscribed message channel and topic to associate with it
	// the channel is used for multiple threads of subscribers for 1 publisher (1-to-many)
	// the messageErrors channel returns the message errors from the caller
	// since subscriber works in asynchronous fashion
	// the function returns error for any subscribe error
	Subscribe(topics []types.TopicChannel, messageErrors chan error) error

	// Disconnect is to close all connections on the message bus
	// and TopicChannel will also be closed
	Disconnect() error
}
