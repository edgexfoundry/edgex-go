/*******************************************************************************
 * Copyright 2019 VMware Inc.
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
 *******************************************************************************/

package notification

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type AddExecutor interface {
	Execute() (string, error)
}

type notificationAdd struct {
	database     NotificationWriter
	notification contract.Notification
}

// This method adds the provided Addressable to the database.
func (op notificationAdd) Execute() (id string, err error) {
	op.notification.ID, err = op.database.AddNotification(op.notification)
	if err != nil {
		if err == db.ErrNotUnique {
			notFoundErr := errors.NewErrNotificationInUse(op.notification.Slug)
			return op.notification.ID, notFoundErr
		}
		return op.notification.ID, err
	}
	return op.notification.ID, nil
}

// This factory method returns an executor used to add an addressable.
func NewAddExecutor(db NotificationWriter, notification contract.Notification) AddExecutor {
	return notificationAdd{
		database:     db,
		notification: notification,
	}
}
