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

package distro

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-messaging/messaging"
	"github.com/edgexfoundry/go-mod-messaging/pkg/types"
)

func initMessaging(client messaging.MessageClient) (chan error, chan types.MessageEnvelope, error) {
	if err := client.Connect(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to message bus: %s ", err.Error())
	}

	errs := make(chan error, 2)
	messages := make(chan types.MessageEnvelope, 10)

	topics := []types.TopicChannel{
		{
			Topic:    Configuration.MessageQueue.Topic,
			Messages: messages,
		},
	}

	LoggingClient.Info("Connecting to incoming message bus at: " + Configuration.MessageQueue.Uri())

	err := client.Subscribe(topics, errs)
	if err != nil {
		close(errs)
		close(messages)
		return nil, nil, fmt.Errorf("failed to subscribe for event messages: %s", err.Error())
	}

	LoggingClient.Info("Connected to inbound event messages for topic: " + Configuration.MessageQueue.Topic)

	return errs, messages, nil
}
