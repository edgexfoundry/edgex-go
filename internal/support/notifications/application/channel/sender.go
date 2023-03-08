//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// Sender abstracts the notification sending via specified channel
type Sender interface {
	Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX)
}

// RESTSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via REST
type RESTSender struct {
	dic *di.Container
}

// NewRESTSender creates the RESTSender instance
func NewRESTSender(dic *di.Container) Sender {
	return &RESTSender{dic: dic}
}

// Send sends the REST request to the specified address
func (sender *RESTSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	lc := container.LoggingClientFrom(sender.dic.Get)

	restAddress, ok := address.(models.RESTAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to RESTAddress", nil)
	}
	// NOTE: Not currently passing an AuthenticationInjector here;
	// no current notifications are calling EdgeX services
	return utils.SendRequestWithRESTAddress(lc, notification.Content, notification.ContentType, restAddress, nil)
}

// EmailSender is the implementation of the interfaces.ChannelSender, which is used to send the notifications via email
type EmailSender struct {
	dic *di.Container
}

// NewEmailSender creates the EmailSender instance
func NewEmailSender(dic *di.Container) Sender {
	return &EmailSender{dic: dic}
}

// Send sends the email to the specified address
func (sender *EmailSender) Send(notification models.Notification, address models.Address) (res string, err errors.EdgeX) {
	smtpInfo := notificationContainer.ConfigurationFrom(sender.dic.Get).Smtp

	emailAddress, ok := address.(models.EmailAddress)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to cast Address to EmailAddress", nil)
	}

	msg := buildSmtpMessage(notification.Sender, smtpInfo.Subject, emailAddress.Recipients, notification.ContentType, notification.Content)
	auth, err := deduceAuth(sender.dic, smtpInfo)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	err = sendEmail(smtpInfo, auth, emailAddress.Recipients, msg)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	return "", nil
}
