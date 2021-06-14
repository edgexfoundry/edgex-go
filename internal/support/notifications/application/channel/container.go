//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// RESTSenderName contains the name of the channel.RESTSender implementation in the DIC.
var RESTSenderName = di.TypeInstanceToName(RESTSender{})

// EmailSenderName contains the name of the channel.EmailSender implementation in the DIC.
var EmailSenderName = di.TypeInstanceToName(EmailSender{})

// RESTSenderFrom helper function queries the DIC and returns the channel.Sender implementation.
func RESTSenderFrom(get di.Get) Sender {
	return get(RESTSenderName).(Sender)
}

// EmailSenderFrom helper function queries the DIC and returns the channel.Sender implementation.
func EmailSenderFrom(get di.Get) Sender {
	return get(EmailSenderName).(Sender)
}
