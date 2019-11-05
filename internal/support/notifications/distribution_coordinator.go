/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package notifications

import (
	notificationsConfig "github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func distribute(
	n models.Notification,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) error {

	loggingClient.Debug("DistributionCoordinator start distributing notification: " + n.Slug)
	var categories []string
	categories = append(categories, string(n.Category))
	subs, err := dbClient.GetSubscriptionByCategoriesLabels(categories, n.Labels)
	if err != nil {
		loggingClient.Error("Unable to get subscriptions to distribute notification:" + n.Slug)
		return err
	}
	for _, sub := range subs {
		send(n, sub, loggingClient, dbClient, config)
	}
	return nil
}

func resend(
	t models.Transmission,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	loggingClient.Debug("Resending transmission: " + t.ID + " for: " + t.Notification.Slug)
	resendViaChannel(t, loggingClient, dbClient, config)
}

func send(
	n models.Notification,
	s models.Subscription,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	for _, ch := range s.Channels {
		sendViaChannel(n, ch, s.Receiver, loggingClient, dbClient, config)
	}
}

func criticalSeverityResend(
	t models.Transmission,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	loggingClient.Info("Critical severity resend scheduler is triggered.")
	resend(t, loggingClient, dbClient, config)
}
