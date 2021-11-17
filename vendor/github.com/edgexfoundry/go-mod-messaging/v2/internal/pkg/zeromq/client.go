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

package zeromq

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	zmq "github.com/pebbe/zmq4"
)

const (
	singleMessagePayloadIndex = 0
	multiMessageTopicIndex    = 0
	multiMessagePayloadIndex  = 1
)

const (
	maxZeroMqSubscribeTopics = 10
)

type zeromqClient struct {
	config        types.MessageBusConfig
	lock          sync.Mutex
	publisher     *zmq.Socket
	subscribers   []*zeromqSubscriber
	messageErrors chan error
	quitSubscribe chan interface{}
}

// NewZeroMqClient instantiates a new zeromq client instance based on the configuration
func NewZeroMqClient(msgConfig types.MessageBusConfig) (*zeromqClient, error) {

	client := zeromqClient{config: msgConfig}

	return &client, nil
}

// Connect implements connect to 0mq
// Since 0mq pub-sub pattern has different pub socket type and sub socket one
// the socket initialization and connection are delayed to Publish and Subscribe calls, respectively
func (client *zeromqClient) Connect() error {
	return nil
}

// Publish sends a message via zeromq with the specified topic
func (client *zeromqClient) Publish(message types.MessageEnvelope, topic string) error {

	var err error

	// Safely binds to port if not already done.
	err = client.bindToPort(client.config.PublishHost.GetHostURL())
	if err != nil {
		return err
	}

	marshaledMsg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	totalLength, err := client.sendMessage(topic, marshaledMsg)

	if err != nil {
		return err
	} else if totalLength != len(topic)+len(marshaledMsg) {
		return errors.New("the length of the sent messages does not match the expected length")
	}

	return err
}

// Subscribe subscribes to each of the specified topics and posts messages receive to the topic's channel. Any
// errors encountered while waiting/processing message are sent to the messageErrors channel
func (client *zeromqClient) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {

	if len(topics) == 0 {
		return errors.New("no topic(s) specified")
	} else if len(topics) > maxZeroMqSubscribeTopics {
		return fmt.Errorf("number of topics(%d) exceeds the maximum capacity(%d)", len(topics), maxZeroMqSubscribeTopics)
	}

	client.quitSubscribe = make(chan interface{})
	client.messageErrors = messageErrors
	client.subscribers = make([]*zeromqSubscriber, len(topics))
	var errorsSubscribe []error
	var err error

	for idx, topic := range topics {
		client.subscribers[idx], err = client.subscribeTopic(&topic)
		if err != nil {
			errorsSubscribe = append(errorsSubscribe, err)
		}
	}

	if len(errorsSubscribe) == 0 {
		return nil
	}

	var errorStr string
	for _, err := range errorsSubscribe {
		errorStr = errorStr + fmt.Sprintf("%s  ", err.Error())
	}
	return errors.New(errorStr)
}

func (client *zeromqClient) subscribeTopic(topic *types.TopicChannel) (*zeromqSubscriber, error) {

	subscriber := zeromqSubscriber{}
	msgQueueURL := client.config.SubscribeHost.GetHostURL()
	if err := subscriber.init(msgQueueURL, topic); err != nil {
		return nil, err
	}

	if err := subscriber.connection.SetSubscribe(topic.Topic); err != nil {
		return nil, fmt.Errorf("error subscribing to topic, %v", err)
	}

	go func(topic *types.TopicChannel) {

		for {
			payloadMsg, err := subscriber.connection.RecvMessage(0)

			select {
			case <-client.quitSubscribe:
				client.lock.Lock()
				_ = subscriber.connection.SetLinger(time.Duration(0))
				err = subscriber.connection.Close()
				if err != nil {
					client.messageErrors <- fmt.Errorf("unable to close socket on subscribe quit, %v", err)
				}
				client.lock.Unlock()
				return
			default:
			}

			// This may occur if try to receive before publisher has established the socket. Just try again.
			if err != nil && err.Error() != "resource temporarily unavailable" {
				continue
			}

			var receivedTopic string

			var payload []byte
			switch msgLen := len(payloadMsg); msgLen {
			case 1:
				payload = []byte(payloadMsg[singleMessagePayloadIndex])
			case 2:
				receivedTopic = payloadMsg[multiMessageTopicIndex]
				payload = []byte(payloadMsg[multiMessagePayloadIndex])
			default:
				client.messageErrors <- fmt.Errorf("found more than 2 incoming messages (1 is no topic, 2 is topic and message), but found: %d", msgLen)
				continue
			}

			msgEnvelope := types.MessageEnvelope{}
			err = json.Unmarshal(payload, &msgEnvelope)
			if err != nil {
				client.messageErrors <- err
				continue
			}

			if len(receivedTopic) == 0 {
				// Publish topic was empty, aka published to any topic,
				// then we have to use the subscribe topic as the received topic
				receivedTopic = topic.Topic
			}

			msgEnvelope.ReceivedTopic = receivedTopic
			topic.Messages <- msgEnvelope
		}
	}(&subscriber.topic)

	return &subscriber, nil
}

func (client *zeromqClient) Disconnect() error {

	var disconnectErrs []error

	if client.quitSubscribe != nil {
		close(client.quitSubscribe)
	}

	for _, subscriber := range client.subscribers {
		// This forces the receive calls to unblock
		err := subscriber.context.Term()

		if err != nil {
			disconnectErrs = append(disconnectErrs, err)
		}
	}

	if client.publisher != nil {
		_ = client.publisher.SetLinger(time.Duration(0))
		err := client.publisher.Close()
		if err != nil {
			disconnectErrs = append(disconnectErrs, err)
		}
	}

	if client.messageErrors != nil {
		close(client.messageErrors)
	}

	for _, subscriber := range client.subscribers {
		if subscriber.topic.Messages != nil {
			close(subscriber.topic.Messages)
		}
	}

	if len(disconnectErrs) == 0 {
		return nil
	}

	var errorStr string
	for _, err := range disconnectErrs {
		if err != nil {
			errorStr = errorStr + fmt.Sprintf("%s  ", err.Error())
		}
	}
	return errors.New(errorStr)
}

func (client *zeromqClient) bindToPort(msgQueueURL string) (err error) {
	client.lock.Lock()
	defer client.lock.Unlock()
	if client.publisher == nil {
		if client.publisher, err = zmq.NewSocket(zmq.PUB); err != nil {
			return
		}
		if conErr := client.publisher.Bind(msgQueueURL); conErr != nil {
			// wrapping the error with msgQueueURL info:
			return fmt.Errorf("error: %v [%s]", conErr, msgQueueURL)
		}

		// allow some time for socket binding before start publishing
		time.Sleep(300 * time.Millisecond)
	}
	return
}

func (client *zeromqClient) sendMessage(topic string, message []byte) (int, error) {
	client.lock.Lock()
	defer client.lock.Unlock()
	return client.publisher.SendMessage(topic, message)
}
