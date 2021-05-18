//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/application/channel"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/infrastructure/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// RESTSenderName contains the name of the application.RESTSender implementation in the DIC.
var RESTSenderName = di.TypeInstanceToName(channel.RESTSender{})

// EmailSenderName contains the name of the application.EmailSender implementation in the DIC.
var EmailSenderName = di.TypeInstanceToName(channel.EmailSender{})

// RESTSenderFrom helper function queries the DIC and returns the interfaces.ChannelSender implementation.
func RESTSenderFrom(get di.Get) interfaces.ChannelSender {
	return get(RESTSenderName).(interfaces.ChannelSender)
}

// EmailSenderFrom helper function queries the DIC and returns the interfaces.ChannelSender implementation.
func EmailSenderFrom(get di.Get) interfaces.ChannelSender {
	return get(EmailSenderName).(interfaces.ChannelSender)
}
