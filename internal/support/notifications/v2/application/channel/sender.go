//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/infrastructure/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// RESTSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via REST
type RESTSender struct {
	lc logger.LoggingClient
}

// NewRESTSender creates the RESTSender instance
func NewRESTSender(lc logger.LoggingClient) interfaces.ChannelSender {
	return &RESTSender{lc: lc}
}

// Send sends the REST request to the specified address
func (sender *RESTSender) Send(content string, contentType string, address models.Address) (res string, err errors.EdgeX) {
	restAddress, ok := address.(models.RESTAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to RESTAddress", nil)
	}
	return utils.SendRequestWithRESTAddress(sender.lc, content, contentType, restAddress)
}

// EmailSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via email
type EmailSender struct {
	lc logger.LoggingClient
}

// NewEmailSender creates the EmailSender instance
func NewEmailSender(lc logger.LoggingClient) interfaces.ChannelSender {
	return &EmailSender{lc: lc}
}

// Send sends the email to the specified address
func (sender *EmailSender) Send(content string, contentType string, address models.Address) (res string, err errors.EdgeX) {
	emailAddress, ok := address.(models.EmailAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to EmailAddress", nil)
	}
	return utils.SendEmailWithAddress(sender.lc, content, contentType, emailAddress)
}
