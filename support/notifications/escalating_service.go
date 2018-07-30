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
	"github.com/edgexfoundry/edgex-go/support/notifications/models"
)

func escalate(t models.Transmission) {
	loggingClient.Warn("Escalating transmission: " + t.ID.String() + ", for: " + t.Notification.Slug)

	var err error
	s, err := dbc.SubscriptionBySlug(ESCALATIONSUBSCRIPTIONSLUG)
	if err == nil {
		n, err := createEscalatedNotification(t)
		if err == nil {
			send(n, s)
		} else {
			loggingClient.Error("Unable to create new escalating notice to send escalation notice for " + t.ID.String())
		}
	} else {
		loggingClient.Error("Unable to find Escalation subcriber to send escalation notice for " + t.ID.String())
	}

}

func createEscalatedNotification(t models.Transmission) (models.Notification, error) {
	old := t.Notification
	n := models.Notification{Category: old.Category, Severity: old.Severity, Description: old.Description, Labels: old.Labels, ContentType: "text/plain"}
	n.Slug = ESCALATIONPREFIX + old.Slug
	n.Sender = ESCALATIONPREFIX + old.Sender
	n.Content = ESCALATEDCONTENTNOTICE + " " + t.String() + " " + old.Content
	n.Status = models.Escalated
	_, err := dbc.AddNotification(&n)
	return n, err
}
