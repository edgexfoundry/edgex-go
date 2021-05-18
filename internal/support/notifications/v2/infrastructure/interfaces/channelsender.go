//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// ChannelSender abstracts the notification sending via specified channel
type ChannelSender interface {
	Send(content string, contentType string, address models.Address) (res string, err errors.EdgeX)
}
