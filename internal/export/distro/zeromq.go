//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	zmq "github.com/pebbe/zmq4"
)

func ZeroMQReceiver(eventCh chan *models.Event) {
	go initZmq(eventCh)
}

func initZmq(eventCh chan *models.Event) {
	q, _ := zmq.NewSocket(zmq.SUB)
	defer q.Close()

	LoggingClient.Info("Connecting to zmq...")
	q.Connect(Configuration.MessageQueue.Uri())
	LoggingClient.Info("Connected to zmq")
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
