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

func distributeAndMark(n models.Notification) error {
	err := distribute(n)
	if err != nil {
		LoggingClient.Error("Trouble on distribution of notification: " + n.Slug)
		return err
	}
	err = dbClient.MarkNotificationProcessed(n)
	if err != nil {
		LoggingClient.Error("Trouble updating notification to Processed for: " + n.Slug)
		return err
	}
	return nil
}
