/********************************************************************************
 *  Copyright 2020 Dell Inc.
 *  Copyright (c) 2021 Intel Corporation
 *  Copyright (c) 2023 IOTech Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

// Package redis contains a RedisClient which leverages go-redis to interact with a Redis server.
package redis

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"

	goRedis "github.com/go-redis/redis/v7"
)

// goRedisWrapper implements RedisClient and uses a underlying 'go-redis' client to communicate with a Redis server.
//
// This functionality was abstracted out from Client so that unit testing can be done easily. The functionality provided
// by this struct can be complex to test and has been tested in the integration test.
type goRedisWrapper struct {
	wrappedClient      *goRedis.Client
	subscriptions      map[string]*goRedis.PubSub
	subscriptionsMutex *sync.Mutex
}

// NewGoRedisClientWrapper creates a RedisClient implementation which uses a 'go-redis' Client to achieve the necessary
// functionality.
func NewGoRedisClientWrapper(redisServerURL string, password string, tlsConfig *tls.Config) (RedisClient, error) {
	options, err := goRedis.ParseURL(redisServerURL)
	if err != nil {
		return nil, err
	}

	options.Password = password
	options.TLSConfig = tlsConfig

	return &goRedisWrapper{
		wrappedClient:      goRedis.NewClient(options),
		subscriptions:      make(map[string]*goRedis.PubSub),
		subscriptionsMutex: &sync.Mutex{},
	}, nil
}

// Send sends the provided message to a topic.
func (g *goRedisWrapper) Send(topic string, message types.MessageEnvelope) error {
	encoded, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = g.wrappedClient.Publish(topic, encoded).Result()
	if err != nil {
		return err
	}

	return nil
}

// Subscribe creates the subscription in Redis
func (g *goRedisWrapper) Subscribe(topic string) error {
	_, err := g.getSubscription(topic)
	return err
}

// Unsubscribe closes the subscription in Redis and removes it.
func (g *goRedisWrapper) Unsubscribe(topic string) {
	g.subscriptionsMutex.Lock()
	defer g.subscriptionsMutex.Unlock()

	subscription := g.subscriptions[topic]
	if subscription == nil {
		return
	}

	_ = subscription.Close()
	delete(g.subscriptions, topic)
}

// Receive retrieves the next message from the specified topic. This operation blocks indefinitely until a
// message is received for the topic.
func (g *goRedisWrapper) Receive(topic string) (*types.MessageEnvelope, error) {
	subscription, err := g.getSubscription(topic)
	if err != nil {
		return nil, err
	}

	data, err := subscription.ReceiveMessage()
	if err != nil {
		return nil, err
	}

	message := &types.MessageEnvelope{}
	payload := []byte(data.Payload)
	err = json.Unmarshal(payload, message)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal payload: %w", err)
	}

	message.ReceivedTopic = data.Channel

	return message, nil
}

// Close closes the subscriptions and the underlying 'go-redis' client.
func (g *goRedisWrapper) Close() error {
	g.subscriptionsMutex.Lock()
	defer g.subscriptionsMutex.Unlock()

	for _, subscription := range g.subscriptions {
		_ = subscription.Close()
	}

	return g.wrappedClient.Close()
}

func (g *goRedisWrapper) getSubscription(topic string) (*goRedis.PubSub, error) {
	g.subscriptionsMutex.Lock()
	defer g.subscriptionsMutex.Unlock()
	subscription, exists := g.subscriptions[topic]
	if !exists {
		// Redis Pub/Sub wildcard doesn't cover empty sub channel level, to match MQTT multi-level wildcard,
		// subscribe additional channel for empty level if the suffix is multiple wildcard
		// for example, subscribing channels a.b and a.b.* is equal to MQTT topic a/b/#
		if strings.HasSuffix(topic, RedisTopicSeparator+RedisWildcard) {
			subscription = g.wrappedClient.PSubscribe(topic, strings.TrimSuffix(topic, RedisTopicSeparator+RedisWildcard))
		} else {
			subscription = g.wrappedClient.PSubscribe(topic)
		}

		res, err := subscription.Receive()
		if err != nil {
			return nil, err
		}

		sub, ok := res.(*goRedis.Subscription)
		if !ok {
			return nil, errors.New("first received message is not redis Subscription")
		}

		if !strings.HasPrefix(sub.String(), "psubscribe") {
			return nil, errors.New("first received message is not subscribe confirmation")
		}

		g.subscriptions[topic] = subscription
	}

	return subscription, nil
}
