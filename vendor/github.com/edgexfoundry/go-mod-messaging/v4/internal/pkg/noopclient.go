//
// Copyright (c) 2022 One Track Consulting
// Copyright (c) 2023 Intel Corporation
// Copyright (c) 2025 IOTech Ltd
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

package pkg

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

type NoopClient struct{}

func (n NoopClient) Request(message types.MessageEnvelope, targetServiceName string, requestTopic string, timeout time.Duration) (*types.MessageEnvelope, error) {
	panic("implement me")
}

func (n NoopClient) Unsubscribe(topics ...string) error {
	panic("implement me")
}

func (n NoopClient) Connect() error {
	panic("implement me")
}

func (n NoopClient) Publish(message types.MessageEnvelope, topic string) error {
	panic("implement me")
}

func (n NoopClient) PublishWithSizeLimit(message types.MessageEnvelope, topic string, limit int64) error {
	panic("implement me")
}

func (n NoopClient) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	panic("implement me")
}

func (n NoopClient) Disconnect() error {
	panic("implement me")
}

func (n NoopClient) PublishBinaryData(data []byte, topic string) error {
	return fmt.Errorf("not supported PublishBinaryData func")
}
func (n NoopClient) SubscribeBinaryData(topics []types.TopicChannel, messageErrors chan error) error {
	return fmt.Errorf("not supported SubscribeBinaryData func")
}
