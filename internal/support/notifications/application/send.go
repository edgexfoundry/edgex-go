//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"
	"time"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/application/channel"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// firstSend sends the notification and return the transmission
func firstSend(dic *di.Container, n models.Notification, trans models.Transmission) models.Transmission {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	record := sendNotificationViaChannel(dic, n, trans.Channel)
	trans.Records = append(trans.Records, record)
	trans.Status = record.Status
	lc.Debugf("sent the notification to %s with address %v, transmission status %s", trans.SubscriptionName, trans.Channel.GetBaseAddress(), trans.Status)
	return trans
}

// reSend sends the Critical notification and return the transmission
func reSend(dic *di.Container, n models.Notification, sub models.Subscription, trans models.Transmission) (models.Transmission, errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

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
		if record.Status == models.Failed {
			// fail to transmit the notification, keep resending
			trans.Status = models.RESENDING
		} else {
			trans.Status = record.Status
		}
		trans.ResendCount = trans.ResendCount + 1
		trans.Records = append(trans.Records, record)
		err = dbClient.UpdateTransmission(trans)
		if err != nil {
			return trans, errors.NewCommonEdgeXWrapper(err)
		}

		if trans.Status == models.RESENDING {
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
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

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
		go transmit(dic, escalated, sub, address) // nolint:errcheck
	}
	return nil
}

func escalatedNotification(n models.Notification, trans models.Transmission) models.Notification {
	n.Id = ""
	n.Created = 0
	n.Content = fmt.Sprintf("[%s %s] %s", models.EscalatedContentNotice, trans.Id, n.Content)
	n.ContentType = common.ContentTypeText
	n.Status = models.Escalated
	return n
}

// sendNotificationViaChannel sends notification via address and return the transmission record. The record status should be SENT or FAILED.
func sendNotificationViaChannel(dic *di.Container, n models.Notification, address models.Address) (transRecord models.TransmissionRecord) {
	var err errors.EdgeX
	transRecord.Status = models.Sent
	switch address.GetBaseAddress().Type {
	case common.REST:
		restSender := channel.RESTSenderFrom(dic.Get)
		transRecord.Response, err = restSender.Send(n, address)
	case common.EMAIL:
		emailSender := channel.EmailSenderFrom(dic.Get)
		transRecord.Response, err = emailSender.Send(n, address)
	default:
		transRecord.Response = fmt.Sprintf("unsupported address type: %s", address.GetBaseAddress().Type)
		return transRecord
	}

	if err != nil {
		transRecord.Status = models.Failed
		transRecord.Response = err.Error()
	}
	transRecord.Sent = pkgCommon.MakeTimestamp()
	return transRecord
}
