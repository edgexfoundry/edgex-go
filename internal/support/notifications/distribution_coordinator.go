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
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func distribute(n models.Notification) error {
	LoggingClient.Debug("DistributionCoordinator start distributing notification: " + n.Slug)
	var categories []string
	categories = append(categories, string(n.Category))
	subs, err := dbClient.GetSubscriptionByCategoriesLabels(categories, n.Labels)
	if err != nil {
		LoggingClient.Error("Unable to get subcriptions to distribute notification:" + n.Slug)
		return err
	}
	for _, sub := range subs {
		send(n, sub)
	}
	return nil
}

func resend(t models.Transmission) {
	LoggingClient.Debug("Resending transmission: " + t.ID + " for: " + t.Notification.Slug)
	resendViaChannel(t)
}

func send(n models.Notification, s models.Subscription) {
	for _, ch := range s.Channels {
		sendViaChannel(n, ch, s.Receiver)
	}
}

func criticalSeverityResend(t models.Transmission) {
	LoggingClient.Info("Critical severity resend scheduler is triggered.")
	LoggingClient.Debug("Resending transmission is: " + t.ID + " for: " + t.Notification.Slug)
	resend(t)
}
