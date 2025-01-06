//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// RESTSenderName contains the name of the channel.RESTSender implementation in the DIC.
var RESTSenderName = di.TypeInstanceToName(RESTSender{})

// EmailSenderName contains the name of the channel.EmailSender implementation in the DIC.
var EmailSenderName = di.TypeInstanceToName(EmailSender{})

// MQTTSenderName contains the name of the channel.MQTTSender implementation in the DIC.
var MQTTSenderName = di.TypeInstanceToName(MQTTSender{})

// ZeroMQTSenderName contains the name of the channel.ZeroMQSender implementation in the DIC.
var ZeroMQTSenderName = di.TypeInstanceToName(ZeroMQSender{})

// RESTSenderFrom helper function queries the DIC and returns the channel.Sender implementation.
func RESTSenderFrom(get di.Get) Sender {
	return get(RESTSenderName).(Sender)
}

// EmailSenderFrom helper function queries the DIC and returns the channel.Sender implementation.
func EmailSenderFrom(get di.Get) Sender {
	return get(EmailSenderName).(Sender)
}

// MQTTSenderFrom helper function queries the DIC and returns the channel.Sender implementation.
func MQTTSenderFrom(get di.Get) Sender {
	return get(MQTTSenderName).(Sender)
}

// ZeroMQSenderFrom helper function queries the DIC and returns the channel.Sender implementation.
func ZeroMQSenderFrom(get di.Get) Sender {
	return get(ZeroMQTSenderName).(Sender)
}
