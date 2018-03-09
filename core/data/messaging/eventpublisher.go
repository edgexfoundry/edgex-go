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
	"github.com/edgexfoundry/edgex-go/core/data/errors"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

// Types of messaging protocols
const (
	ZEROMQ int = iota
	MQTT
)

// Publisher to send events to northbound services
type EventPublisher struct {
	protocol int
	zmq      zeroMQEventPublisher
}

func NewZeroMQPublisher(configuration ZeroMQConfiguration) *EventPublisher {
	return &EventPublisher{protocol: ZEROMQ, zmq: newZeroMQEventPublisher(configuration)}
}

// Send the event
func (ep *EventPublisher) SendEventMessage(e models.Event) error {
	// Switch based on the protocol you're using
	switch ep.protocol {
	case ZEROMQ:
		return ep.zmq.SendEventMessage(e)
	default:
		return errors.UnsupportedPublisher{}
	}
}
