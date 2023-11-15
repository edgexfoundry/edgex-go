/********************************************************************************
 *  Copyright 2020 Dell Inc.
 *  Copyright (c) 2023 Intel Corporation
 *  Copyright (C) 2023 IOTech Ltd
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

package redis

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg"
	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
)

const (
	StandardTopicSeparator = "/"
	RedisTopicSeparator    = "."
	StandardWildcard       = "#"
	SingleLevelWildcard    = "+"
	RedisWildcard          = "*"
)

// Client MessageClient implementation which provides functionality for sending and receiving messages using
// Redis Pub/Sub.
type Client struct {
	redisClient RedisClient

	// Used to avoid multiple subscriptions to the same topic
	existingTopics map[string]bool
	mapMutex       *sync.Mutex
}

// NewClient creates a new Client based on the provided configuration.
func NewClient(messageBusConfig types.MessageBusConfig) (Client, error) {
	return NewClientWithCreator(messageBusConfig, NewGoRedisClientWrapper, tls.X509KeyPair, tls.LoadX509KeyPair,
		x509.ParseCertificate, os.ReadFile, pem.Decode)
}

// NewClientWithCreator creates a new Client based on the provided configuration while allowing more control on the
// creation of the underlying entities such as certs, keys, and Redis clients
func NewClientWithCreator(
	messageBusConfig types.MessageBusConfig,
	creator RedisClientCreator,
	pairCreator pkg.X509KeyPairCreator,
	keyLoader pkg.X509KeyLoader,
	caCertCreator pkg.X509CaCertCreator,
	caCertLoader pkg.X509CaCertLoader,
	pemDecoder pkg.PEMDecoder) (Client, error) {

	// Parse Optional configuration properties
	optionalClientConfiguration, err := NewClientConfiguration(messageBusConfig)
	if err != nil {
		return Client{}, err
	}

	// Parse TLS configuration properties
	tlsConfigurationOptions := pkg.TlsConfigurationOptions{}
	err = pkg.Load(messageBusConfig.Optional, &tlsConfigurationOptions)
	if err != nil {
		return Client{}, err
	}

	var client RedisClient

	// Create underlying client to use when publishing
	if !messageBusConfig.Broker.IsHostInfoEmpty() {
		client, err = createRedisClient(
			messageBusConfig.Broker.GetHostURL(),
			optionalClientConfiguration,
			tlsConfigurationOptions,
			creator,
			pairCreator,
			keyLoader,
			caCertCreator,
			caCertLoader,
			pemDecoder)

		if err != nil {
			return Client{}, err
		}
	}

	return Client{
		redisClient:    client,
		existingTopics: make(map[string]bool),
		mapMutex:       new(sync.Mutex),
	}, nil
}

// Connect noop as preemptive connections are not needed.
func (c Client) Connect() error {
	// No need to connect, connection pooling is handled by the underlying client.
	return nil
}

// Publish sends the provided message to appropriate Redis Pub/Sub.
func (c Client) Publish(message types.MessageEnvelope, topic string) error {
	if c.redisClient == nil {
		return pkg.NewMissingConfigurationErr("Broker", "Unable to create a connection for publishing")
	}

	if topic == "" {
		// Empty topics are not allowed for Redis
		return pkg.NewInvalidTopicErr("", "Unable to publish to the invalid topic")
	}

	topic = convertToRedisTopicScheme(topic)
	var err error
	if err = c.redisClient.Send(topic, message); err != nil && strings.Contains(err.Error(), "EOF") {
		// Redis may have been restarted and the first attempt will fail with EOF, so need to try again
		err = c.redisClient.Send(topic, message)
	}

	return err
}

// Subscribe creates background processes which reads messages from the appropriate Redis Pub/Sub and sends to the
// provided channels
func (c Client) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	if c.redisClient == nil {
		return pkg.NewMissingConfigurationErr("Broker", "Unable to create a connection for subscribing")
	}

	err := c.validateTopics(topics)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for i := range topics {
		wg.Add(1)
		go func(topic types.TopicChannel) {
			topicName := convertToRedisTopicScheme(topic.Topic)
			messageChannel := topic.Messages
			var previousErr error

			err := c.redisClient.Subscribe(topicName)
			wg.Done()
			if err != nil {
				messageErrors <- err
				return
			}

			for {

				message, err := c.redisClient.Receive(topicName)

				// Make sure the topic is still subscribed before processing the message.
				// If not cleanup and exit from the go func
				c.mapMutex.Lock()
				subscribed := c.existingTopics[topic.Topic]
				if !subscribed {
					delete(c.existingTopics, topic.Topic)
					// Do the actual unsubscribe from Redis with the Redis style topic now that this go func can exit.
					c.redisClient.Unsubscribe(topicName)
					c.mapMutex.Unlock()
					return
				}
				c.mapMutex.Unlock()

				if err != nil {
					// This handles case when getting same repeated error due to Redis connectivity issue
					// Avoids starving of other threads/processes and recipient spamming the log file.
					if previousErr != nil && reflect.DeepEqual(err, previousErr) {
						time.Sleep(1 * time.Millisecond) // Sleep allows other threads to get time
						continue
					}
					messageErrors <- err

					previousErr = err
					continue
				}

				previousErr = nil
				message.ReceivedTopic = convertFromRedisTopicScheme(message.ReceivedTopic)

				messageChannel <- *message
			}
		}(topics[i])
	}

	// Wait for all the subscribe go funcs to be spun up
	// This is needed for the Request API since the subscribe must be spun up prior to the response being published.
	wg.Wait()

	return nil

}

func (c Client) Request(message types.MessageEnvelope, requestTopic string, responseTopicPrefix string, timeout time.Duration) (*types.MessageEnvelope, error) {
	return pkg.DoRequest(c.Subscribe, c.Unsubscribe, c.Publish, message, requestTopic, responseTopicPrefix, timeout)
}

func (c Client) Unsubscribe(topics ...string) error {
	c.mapMutex.Lock()

	for _, topic := range topics {
		c.existingTopics[topic] = false
	}

	c.mapMutex.Unlock()

	for _, topic := range topics {
		// Must publish a dummy message to kick the subscription go func out of the Receive() call
		_ = c.Publish(types.MessageEnvelope{}, topic)
	}

	return nil
}

// Disconnect closes connections to the Redis server.
func (c Client) Disconnect() error {
	var disconnectErrors []string
	if c.redisClient != nil {
		err := c.redisClient.Close()
		if err != nil {
			disconnectErrors = append(disconnectErrors, fmt.Sprintf("Unable to disconnect publish client: %v", err))
		}
	}

	if len(disconnectErrors) > 0 {
		return NewDisconnectErr(disconnectErrors)
	}

	return nil
}

func (c Client) validateTopics(topics []types.TopicChannel) error {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()

	// First validate all the topics are unique, i.e. not existing subscription
	for _, topic := range topics {
		_, exists := c.existingTopics[topic.Topic]
		if exists {
			return fmt.Errorf("subscription for '%s' topic already exists, must be unique", topic.Topic)
		}

		c.existingTopics[topic.Topic] = true
	}

	return nil
}

// createRedisClient helper function for creating RedisClient implementations.
func createRedisClient(
	redisServerURL string,
	optionalClientConfiguration OptionalClientConfiguration,
	tlsConfigurationOptions pkg.TlsConfigurationOptions,
	creator RedisClientCreator,
	pairCreator pkg.X509KeyPairCreator,
	keyLoader pkg.X509KeyLoader,
	caCertCreator pkg.X509CaCertCreator,
	caCertLoader pkg.X509CaCertLoader,
	pemDecoder pkg.PEMDecoder) (RedisClient, error) {

	tlsConfig, err := pkg.GenerateTLSForClientClientOptions(
		redisServerURL,
		tlsConfigurationOptions,
		pairCreator,
		keyLoader,
		caCertCreator,
		caCertLoader,
		pemDecoder)

	if err != nil {
		return nil, err
	}

	return creator(redisServerURL, optionalClientConfiguration.Password, tlsConfig)
}

func convertToRedisTopicScheme(topic string) string {
	// Redis Pub/Sub uses "." for separator and "*" for wild cards.
	// Since we have standardized on the MQTT style scheme of "/" & "#" & "+" we need to
	// convert it to the Redis Pub/Sub scheme.
	topic = strings.Replace(topic, StandardTopicSeparator, RedisTopicSeparator, -1)
	topic = strings.Replace(topic, StandardWildcard, RedisWildcard, -1)
	topic = strings.Replace(topic, SingleLevelWildcard, RedisWildcard, -1)

	return topic
}

func convertFromRedisTopicScheme(topic string) string {
	// Redis Pub/Sub uses "." for separator and "*" for wild cards.
	// Since we have standardized on the MQTT style scheme of "/" & "#" & "+" we need to
	// convert it from the Redis Pub/Sub scheme.
	topic = strings.Replace(topic, RedisWildcard+RedisTopicSeparator, SingleLevelWildcard+StandardTopicSeparator, -1)
	topic = strings.Replace(topic, RedisTopicSeparator, StandardTopicSeparator, -1)
	topic = strings.Replace(topic, RedisWildcard, StandardWildcard, -1)

	return topic
}
