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

//go:build include_nats_messaging

package nats

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v4/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v4/internal/pkg/nats/interfaces"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/hashicorp/go-multierror"
	"github.com/nats-io/nats.go"
)

func newConnection(cc ClientConfig) (interfaces.Connection, error) {
	co, err := cc.ConnectOpt()

	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(cc.BrokerURL, co...)

	if err != nil {
		return nil, err
	}

	return nc, nil
}

// NewClient initializes creates a new client using NATS core
func NewClient(cfg types.MessageBusConfig) (*Client, error) {
	return NewClientWithConnectionFactory(cfg, newConnection)
}

// NewClientWithConnectionFactory creates a new client that uses the specified function to establish pub/sub connections
func NewClientWithConnectionFactory(cfg types.MessageBusConfig, connectionFactory ConnectNats) (*Client, error) {
	if connectionFactory == nil {
		return nil, fmt.Errorf("connectionFactory is required")
	}

	var m interfaces.MarshallerUnmarshaller

	cc, err := CreateClientConfiguration(cfg)

	if err != nil {
		return nil, err
	}

	switch strings.ToLower(cc.Format) {
	case "json":
		m = &jsonMarshaller{opts: cc}
	default:
		m = &natsMarshaller{opts: cc}
	}

	return &Client{
		config:                cc,
		connect:               connectionFactory,
		m:                     m,
		existingSubscriptions: make(map[string]*nats.Subscription),
		subscriptionMutex:     new(sync.Mutex),
	}, nil
}

// Client provides NATS MessageBus implementations per the underlying connection
type Client struct {
	connect               ConnectNats
	connection            interfaces.Connection
	m                     interfaces.MarshallerUnmarshaller
	config                ClientConfig
	existingSubscriptions map[string]*nats.Subscription
	subscriptionMutex     *sync.Mutex
}

// Connect establishes the connections to publish and subscribe hosts
func (c *Client) Connect() error {
	if c.connection != nil {
		return fmt.Errorf("already connected to NATS")
	}

	if c.connect == nil {
		return fmt.Errorf("connection function not specified")
	}

	conn, err := c.connect(c.config)

	if err != nil {
		return err
	}

	c.connection = conn

	return nil
}

// Publish publishes EdgeX messages to NATS
func (c *Client) Publish(message types.MessageEnvelope, topic string) error {
	if c.connection == nil {
		return fmt.Errorf("cannot publish with disconnected client")
	}

	if topic == "" {
		return fmt.Errorf("cannot publish to empty topic")
	}

	msg, err := c.m.Marshal(message, topic)

	if err != nil {
		return err
	}

	return c.connection.PublishMsg(msg)
}

// PublishWithSizeLimit checks the message size and publishes EdgeX messages to NATS
func (c *Client) PublishWithSizeLimit(message types.MessageEnvelope, topic string, limit int64) error {
	if c.connection == nil {
		return fmt.Errorf("cannot publish with disconnected client")
	}

	if topic == "" {
		return fmt.Errorf("cannot publish to empty topic")
	}

	msg, err := c.m.Marshal(message, topic)

	if err != nil {
		return err
	}

	if limit > 0 && int64(msg.Size()) > limit*1024 {
		return fmt.Errorf("message size exceed limit(%d KB)", limit)
	}

	return c.connection.PublishMsg(msg)
}

// Subscribe establishes NATS subscriptions for the given topics
func (c *Client) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	if c.connection == nil {
		return fmt.Errorf("cannot subscribe with disconnected client")
	}

	c.subscriptionMutex.Lock()
	defer c.subscriptionMutex.Unlock()

	for _, tc := range topics {
		s := TopicToSubject(tc.Topic)

		subscription, err := c.connection.QueueSubscribe(s, c.config.QueueGroup, func(msg *nats.Msg) {
			env := types.MessageEnvelope{}
			err := c.m.Unmarshal(msg, &env)
			if err != nil {
				messageErrors <- err
			} else {
				tc.Messages <- env
			}

			// core nats messages without reply do not need to be ack'd
			if msg.Reply != "" {
				var ackErr error
				if c.config.ExactlyOnce {
					// AckSync carries a performance penalty
					// but is needed for subscribe side of ExactlyOnce
					ackErr = msg.AckSync()
				} else {
					ackErr = msg.Ack()
				}
				if ackErr != nil {
					messageErrors <- ackErr
				}
			}
		})

		if err != nil {
			return err
		}

		c.existingSubscriptions[tc.Topic] = subscription
	}

	return nil
}

// Request publishes a request and waits for a response
func (c *Client) Request(message types.MessageEnvelope, requestTopic string, responseTopicPrefix string, timeout time.Duration) (*types.MessageEnvelope, error) {
	return pkg.DoRequest(c.Subscribe, c.Unsubscribe, c.Publish, message, requestTopic, responseTopicPrefix, timeout)
}

// Unsubscribe to unsubscribe from the specified topics.
func (c *Client) Unsubscribe(topics ...string) error {
	if c.connection == nil {
		return fmt.Errorf("cannot unsubscribe with disconnected client")
	}

	c.subscriptionMutex.Lock()
	defer c.subscriptionMutex.Unlock()

	var errs error
	for _, topic := range topics {
		subscription, exists := c.existingSubscriptions[topic]
		if !exists {
			continue
		}

		// If the subscription doesn't exist, not need to and can't unsubscribe from it.
		if subscription != nil {
			err := subscription.Unsubscribe()
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("unable to unsubscribe to topic '%s': %w", topic, err))
				continue
			}
		}

		delete(c.existingSubscriptions, topic)
	}

	return errs
}

// Disconnect drains open subscriptions before closing
func (c *Client) Disconnect() error {
	if c.connection == nil {
		return nil
	}
	return c.connection.Drain()
}

func (c *Client) PublishBinaryData(data []byte, topic string) error {
	return fmt.Errorf("not supported PublishBinaryData func")
}

func (c *Client) SubscribeBinaryData(topics []types.TopicChannel, messageErrors chan error) error {
	return fmt.Errorf("not supported SubscribeBinaryData func")
}
