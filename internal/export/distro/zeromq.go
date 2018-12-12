//
// Copyright (c) 2017 Cavium
// Copyright (c) 2018 Dell Technologies, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	zmq "github.com/pebbe/zmq4"
)

// ZeroMQ implementation of the event publisher
type zeroMQEventPublisher struct {
	publisher *zmq.Socket
	mux       sync.Mutex
}

func newZeroMQEventPublisher() sender {
	newPublisher, _ := zmq.NewSocket(zmq.PUB)
	LoggingClient.Info("Connecting to analytics 0MQ at: " + Configuration.AnalyticsQueue.Uri())
	newPublisher.Bind(Configuration.AnalyticsQueue.Uri())
	LoggingClient.Info("Connected to analytics outbound 0MQ")
	sender := &zeroMQEventPublisher{
		publisher: newPublisher,
	}
	return sender
}

func (sender *zeroMQEventPublisher) Send(data []byte, event *models.Event) bool {
	sender.mux.Lock()
	defer sender.mux.Unlock()
	LoggingClient.Debug("Sending data to 0MQ: " + string(data[:]))
	_, err := sender.publisher.SendBytes(data, 0)
	if err != nil {
		LoggingClient.Error("Issue trying to publish to 0MQ...")
		return false
	}
	return true
}

func ZeroMQReceiver(eventCh chan *models.Event) {
	go initZmq(eventCh)
}

func initZmq(eventCh chan *models.Event) {
	q, _ := zmq.NewSocket(zmq.SUB)
	defer q.Close()

	LoggingClient.Info("Connecting to incoming 0MQ at: " + Configuration.MessageQueue.Uri())
	q.Connect(Configuration.MessageQueue.Uri())
	LoggingClient.Info("Connected to inbound 0MQ")
	q.SetSubscribe("")

	for {
		msg, err := q.RecvMessage(0)
		if err != nil {
			id, _ := q.GetIdentity()
			LoggingClient.Error(fmt.Sprintf("Error getting message %s", id))
		} else {
			for _, str := range msg {
				event := parseEvent(str)
				LoggingClient.Info(fmt.Sprintf("Event received: %s", str))
				eventCh <- event
			}
		}
	}
}

func parseEvent(str string) *models.Event {
	event := models.Event{}

	if err := json.Unmarshal([]byte(str), &event); err != nil {
		LoggingClient.Error(err.Error())
		LoggingClient.Warn("Failed to parse event")
		return nil
	}
	return &event
}
