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

type IdExecutor interface {
	Execute() (contract.Notification, error)
}

type notificationLoadById struct {
	database NotificationLoader
	id       string
}

func (op notificationLoadById) Execute() (contract.Notification, error) {
	res, err := op.database.GetNotificationById(op.id)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrNotificationNotFound(op.id)
		}
		return res, err
	}
	return res, nil
}

func NewIdExecutor(db NotificationLoader, id string) IdExecutor {
	return notificationLoadById{
		database: db,
		id:       id,
	}
}
