//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

// +build zeromq

package distro

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/export-go"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
)

const (
	zeroMQPort = 5563
)

func ZeroMQReceiver(eventCh chan *export.Event) {
	go initZmq(eventCh)
}

func initZmq(eventCh chan *export.Event) {
	q, _ := zmq.NewSocket(zmq.SUB)
	defer q.Close()

	logger.Info("Connecting to zmq...")
	url := fmt.Sprintf("tcp://%s:%d", cfg.DataHost, zeroMQPort)
	q.Connect(url)
	logger.Info("Connected to zmq")
	q.SetSubscribe("")

	for {
		msg, err := q.RecvMessage(0)
		if err != nil {
			id, _ := q.GetIdentity()
			logger.Error("Error getting mesage", zap.String("id", id))
		} else {
			for _, str := range msg {
				event := parseEvent(str)
				logger.Info("Event received", zap.Any("event", event))
				eventCh <- event
			}
		}
	}
}

func parseEvent(str string) *export.Event {
	event := export.Event{}

	if err := json.Unmarshal([]byte(str), &event); err != nil {
		logger.Error("Failed to parse event", zap.Error(err))
		return nil
	}
	return &event
}
