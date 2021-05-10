//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// asyncDistributeNotification distributes notification to associate subscriptions
func asyncDistributeNotification(dic *di.Container, n models.Notification) {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	n.Status = models.Processed
	err := dbClient.UpdateNotification(n)
	if err != nil {
		lc.Errorf("fail to update notification status to processed", err)
		return
	}

	var categories []string
	if n.Category != "" {
		categories = append(categories, n.Category)
	}
	subs, err := dbClient.SubscriptionsByCategoriesAndLabels(0, -1, categories, n.Labels)
	if err != nil {
		lc.Errorf("fail to query subscriptions to distribute notification", err)
		return
	}

	for _, sub := range subs {
		for _, address := range sub.Channels {
			go asyncHandleNotification(dic, n, sub, address)
		}
	}
}

// asyncHandleNotification handle the notification sending with specified subscription and address
func asyncHandleNotification(dic *di.Container, n models.Notification, sub models.Subscription, address models.Address) {
	lc := container.LoggingClientFrom(dic.Get)
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)

	trans := models.NewTransmission(sub.Name, address, n.Id)
	trans, err := normalSend(dic, n, trans)
	if err != nil {
		lc.Errorf("fail to handle the notification sending for the subscription %s with address %v, err: %v", sub.Name, address.GetBaseAddress(), err)
		return
	}
	trans, err = dbClient.AddTransmission(trans)
	if err != nil {
		lc.Errorf(err.Message())
		return
	}

	if n.Status == models.Escalated {
		// Do not resend if the notification status is Escalated
		return
	}

	// Resend the critical notification if the transmission is failed.
	if n.Severity == models.Critical && trans.Status == models.Failed {
		trans, err = criticalSend(dic, n, sub, trans)
		if err != nil {
			lc.Errorf("fail to handle the critical notification sending for the subscription %s with address %v, err: %v", sub.Name, address.GetBaseAddress(), err)
			return
		}
	}
	// Trigger a escalated notification if the transmission is Escalated
	if trans.Status == models.Escalated {
		err = escalatedSend(dic, n, trans)
		if err != nil {
			lc.Errorf("fail to handle the escalated notification sending, err: %v", err)
		}
	}
}
