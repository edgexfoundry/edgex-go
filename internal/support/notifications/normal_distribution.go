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

func distributeAndMark(
	n models.Notification,
	loggingClient logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) error {

	go distribute(n, loggingClient, dbClient, config)

	err := dbClient.MarkNotificationProcessed(n)
	if err != nil {
		loggingClient.Error("Trouble updating notification to Processed for: " + n.Slug)
		return err
	}
	return nil
}
