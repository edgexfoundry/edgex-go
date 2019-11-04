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
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func escalate(
	t models.Transmission,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient) {

	loggingClient.Warn("Escalating transmission: " + t.ID + ", for: " + t.Notification.Slug)

	var err error
	s, err := dbClient.GetSubscriptionBySlug(ESCALATIONSUBSCRIPTIONSLUG)
	if err != nil {
		loggingClient.Error("Unable to find Escalation subscriber to send escalation notice for " + t.ID)
		return
	}

	n, err := createEscalatedNotification(t, dbClient)
	if err != nil {
		loggingClient.Error("Unable to create new escalating notice to send escalation notice for " + t.ID)
		return
	}

	send(n, s, loggingClient, dbClient)
}

func createEscalatedNotification(
	t models.Transmission,
	dbClient interfaces.DBClient) (models.Notification, error) {

	old := t.Notification
	n := models.Notification{Category: old.Category, Severity: old.Severity, Description: old.Description, Labels: old.Labels, ContentType: "text/plain"}
	n.Slug = ESCALATIONPREFIX + old.Slug
	n.Sender = ESCALATIONPREFIX + old.Sender
	n.Content = ESCALATEDCONTENTNOTICE + " " + t.String() + " " + old.Content
	n.Status = models.Escalated
	_, err := dbClient.AddNotification(n)
	return n, err
}
