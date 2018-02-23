/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package messaging

import (
	"encoding/json"
	"sync"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	zmq "github.com/pebbe/zmq4"
)

// Configuration struct for ZeroMQ
type ZeroMQConfiguration struct {
	AddressPort string
}

// ZeroMQ implementation of the event publisher
type zeroMQEventPublisher struct {
	publisher *zmq.Socket
	mux       sync.Mutex
}

func newZeroMQEventPublisher(config ZeroMQConfiguration) zeroMQEventPublisher {
	newPublisher, _ := zmq.NewSocket(zmq.PUB)
	newPublisher.Bind(config.AddressPort)

	return zeroMQEventPublisher{
		publisher: newPublisher,
	}
}

func (zep *zeroMQEventPublisher) SendEventMessage(e models.Event) error {
	s, err := json.Marshal(&e)
	if err != nil {
		return err
	}
	zep.mux.Lock()
	defer zep.mux.Unlock()
	_, err = zep.publisher.SendBytes(s, 0)
	if err != nil {
		return err
	}

	return nil
}
