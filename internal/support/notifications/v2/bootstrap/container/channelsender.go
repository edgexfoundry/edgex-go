//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/application/channel"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// RESTSenderName contains the name of the channel.RESTSender implementation in the DIC.
var RESTSenderName = di.TypeInstanceToName(channel.RESTSender{})

// EmailSenderName contains the name of the channel.EmailSender implementation in the DIC.
var EmailSenderName = di.TypeInstanceToName(channel.EmailSender{})

// RESTSenderFrom helper function queries the DIC and returns the channel.Sender implementation.
func RESTSenderFrom(get di.Get) channel.Sender {
	return get(RESTSenderName).(channel.Sender)
}

// EmailSenderFrom helper function queries the DIC and returns the channel.Sender implementation.
func EmailSenderFrom(get di.Get) channel.Sender {
	return get(EmailSenderName).(channel.Sender)
}
