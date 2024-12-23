// Copyright (C) 2024-2025 IOTech Ltd

package channel

import (
	"fmt"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	zmq "github.com/pebbe/zmq4"
)

const (
	ProtocolTCP = "tcp"
)

// prepareZeroMQTClient creates a new client or load the exist client from cache
func (sender *ZeroMQSender) prepareZeroMQTClient(address models.ZeroMQAddress) (*zmq.Socket, errors.EdgeX) {
	client := sender.loadClient(address)
	if client != nil {
		return client, nil
	}

	client, err := sender.createClient(address)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	return client, nil
}

func (sender *ZeroMQSender) loadClient(address models.ZeroMQAddress) *zmq.Socket {
	sender.mutex.RLock()
	defer sender.mutex.RUnlock()
	key := strconv.Itoa(address.Port)
	socket, ok := sender.clientCache[key]
	if ok {
		return socket
	}
	return nil
}

// createMqttClient creates a new ZeroMQ client
func (sender *ZeroMQSender) createClient(address models.ZeroMQAddress) (*zmq.Socket, errors.EdgeX) {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	// Check the cache before creating new MQTT client
	key := strconv.Itoa(address.Port)
	socket, ok := sender.clientCache[key]
	if ok {
		return socket, nil
	}

	socket, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	url := fmt.Sprintf("%s://%s:%d", ProtocolTCP, address.Host, address.Port)
	err = socket.Bind(url)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	// allow some time for socket binding before start publishing
	// TODO: It's known as "slow joiner" symptom. https://zguide.zeromq.org/docs/chapter1/#Getting-the-Message-Out
	//  We should avoid using any hardcoded sleep. Read https://zguide.zeromq.org/docs/chapter2/#sockets-and-patterns to see how to do this right.
	time.Sleep(300 * time.Millisecond)

	sender.clientCache[key] = socket

	return socket, nil
}
