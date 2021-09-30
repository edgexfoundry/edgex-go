//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// distribute distributes notification to associate subscriptions
func distribute(dic *di.Container, n models.Notification) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	var categories []string
	if n.Category != "" {
		categories = append(categories, n.Category)
	}
	subs, err := dbClient.SubscriptionsByCategoriesAndLabels(0, -1, categories, n.Labels)
	if err != nil {
		lc.Errorf("fail to query subscriptions to distribute notification", err)
		return errors.NewCommonEdgeXWrapper(err)
	}

	for _, sub := range subs {
		if sub.AdminState == models.Locked {
			lc.Debugf("subscription %s is locked, skip the notification transmission", sub.Name)
			continue
		}
		for _, address := range sub.Channels {
			// Async transmit the notification to improve the performance
			go transmit(dic, n, sub, address) // nolint:errcheck
		}
	}

	n.Status = models.Processed
	err = dbClient.UpdateNotification(n)
	if err != nil {
		lc.Errorf("fail to update notification status to processed", err)
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// transmit transmits the notification with specified subscription and address
func transmit(dic *di.Container, n models.Notification, sub models.Subscription, address models.Address) (models.Transmission, errors.EdgeX) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)

	trans := models.NewTransmission(sub.Name, address, n.Id)
	trans = firstSend(dic, n, trans)
	trans, err := dbClient.AddTransmission(trans)
	if err != nil {
		lc.Error(err.Message())
		return trans, errors.NewCommonEdgeXWrapper(err)
	}

	if n.Status == models.Escalated {
		// Do not resend if the notification status is Escalated
		return trans, nil
	}

	// Resend the critical notification if the transmission is failed.
	if n.Severity == models.Critical && trans.Status == models.Failed {
		// Change the transmission status to RESENDING which means this transmission process is resending the notification and should not be removed.
		trans.Status = models.RESENDING
		err = dbClient.UpdateTransmission(trans)
		if err != nil {
			lc.Error(err.Message())
			return trans, errors.NewCommonEdgeXWrapper(err)
		}
		trans, err = reSend(dic, n, sub, trans)
		if err != nil {
			lc.Errorf("fail to handle the critical notification sending for the subscription %s with address %v, err: %v", sub.Name, address.GetBaseAddress(), err)
			return trans, errors.NewCommonEdgeXWrapper(err)
		}
	}
	// Trigger a escalated notification if the transmission is Escalated
	if trans.Status == models.Escalated {
		err = escalatedSend(dic, n, trans)
		if err != nil {
			lc.Errorf("fail to handle the escalated notification sending, err: %v", err)
			return trans, errors.NewCommonEdgeXWrapper(err)
		}
	}
	return trans, nil
}
