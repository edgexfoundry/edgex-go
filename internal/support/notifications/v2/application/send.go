//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/common"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// firstSend sends the notification and return the transmission
func firstSend(dic *di.Container, n models.Notification, trans models.Transmission) models.Transmission {
	lc := container.LoggingClientFrom(dic.Get)

	record := sendNotificationViaChannel(dic, n, trans.Channel)
	trans.Records = append(trans.Records, record)
	trans.Status = record.Status
	lc.Debugf("sent the notification to %s with address %v, transmission status %s", trans.SubscriptionName, trans.Channel.GetBaseAddress(), trans.Status)
	return trans
}

// reSend sends the Critical notification and return the transmission
func reSend(dic *di.Container, n models.Notification, sub models.Subscription, trans models.Transmission) (models.Transmission, errors.EdgeX) {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)
	config := notificationContainer.ConfigurationFrom(dic.Get)

	resendLimit, resendInterval, err := resendLimitAndInterval(config, sub)
	if err != nil {
		return trans, errors.NewCommonEdgeXWrapper(err)
	}
	for i := 1; i <= resendLimit; i++ {
		// Since this sending process is triggered for the critical notification which is failed to send to the subscription at the first time,
		// so we wait seconds and retry to send the notification again.
		time.Sleep(resendInterval)
		lc.Warn("fail to send the critical notification. Retry to send again...")

		record := sendNotificationViaChannel(dic, n, trans.Channel)
		trans.ResendCount = trans.ResendCount + 1
		trans.Status = record.Status
		trans.Records = append(trans.Records, record)
		err = dbClient.UpdateTransmission(trans)
		if err != nil {
			return trans, errors.NewCommonEdgeXWrapper(err)
		}
		if trans.Status == models.Failed {
			continue
		}
		lc.Debugf("success to send the critical notification to %s with address %v, transmission Id: %s", trans.SubscriptionName, trans.Channel.GetBaseAddress(), trans.Id)
		return trans, nil
	}

	lc.Warn("Resend count exceeds the configurable limit, escalate the transmission.")
	trans.Status = models.Escalated
	err = dbClient.UpdateTransmission(trans)
	if err != nil {
		return trans, errors.NewCommonEdgeXWrapper(err)
	}
	return trans, nil
}

func resendLimitAndInterval(config *config.ConfigurationStruct, sub models.Subscription) (int, time.Duration, errors.EdgeX) {
	resendLimit := config.Writable.ResendLimit
	if sub.ResendLimit > 0 {
		resendLimit = sub.ResendLimit
	}
	resendInterval := config.Writable.ResendInterval
	if sub.ResendInterval != "" {
		resendInterval = sub.ResendInterval
	}
	resendIntervalDuration, err := time.ParseDuration(resendInterval)
	if err != nil {
		return 0, time.Second, errors.NewCommonEdgeX(errors.KindContractInvalid, "fail to parse resendInterval", err)
	}
	return resendLimit, resendIntervalDuration, nil
}

// escalatedSend handle the escalated notification for the ESCALATION subscription
func escalatedSend(dic *di.Container, n models.Notification, trans models.Transmission) errors.EdgeX {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	sub, err := dbClient.SubscriptionByName(models.EscalationSubscriptionName)
	if err != nil {
		lc.Warnf(fmt.Sprintf("subscription %s does not exists, skip the escalated notification sending", models.EscalationSubscriptionName))
		return nil
	}

	escalated := escalatedNotification(n, trans)
	escalated, err = dbClient.AddNotification(escalated)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "fail to create the escalated notification", err)
	}

	for _, address := range sub.Channels {
		go transmit(dic, escalated, sub, address)
	}
	return nil
}

func escalatedNotification(n models.Notification, trans models.Transmission) models.Notification {
	n.Id = ""
	n.Created = 0
	n.Content = fmt.Sprintf("[%s %s] %s", models.EscalatedContentNotice, trans.Id, n.Content)
	n.ContentType = clients.ContentTypeText
	n.Status = models.Escalated
	return n
}

// sendNotificationViaChannel sends notification via address and return the transmission record. The record status should be SENT or FAILED.
func sendNotificationViaChannel(dic *di.Container, n models.Notification, channel models.Address) (transRecord models.TransmissionRecord) {
	var err errors.EdgeX
	transRecord.Status = models.Sent
	switch channel.GetBaseAddress().Type {
	case v2.REST:
		restSender := v2NotificationsContainer.RESTSenderFrom(dic.Get)
		transRecord.Response, err = restSender.Send(n, channel)
	case v2.EMAIL:
		emailSender := v2NotificationsContainer.EmailSenderFrom(dic.Get)
		transRecord.Response, err = emailSender.Send(n, channel)
	default:
		transRecord.Response = fmt.Sprintf("unsupported address type: %s", channel.GetBaseAddress().Type)
		return transRecord
	}

	if err != nil {
		transRecord.Status = models.Failed
		transRecord.Response = err.Error()
	}
	transRecord.Sent = common.MakeTimestamp()
	return transRecord
}
