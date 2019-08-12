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

type CollectionExecutor interface {
	Execute() ([]contract.Notification, error)
}

type notificationLoadById struct {
	database NotificationLoader
	id       string
}

type notificationLoadBySlug struct {
	database NotificationLoader
	slug     string
}

type notificationsLoadBySender struct {
	database NotificationLoader
	limit    int
	sender   string
}

type notificationsLoadByStartEnd struct {
	database NotificationLoader
	limit    int
	start    int64
	end      int64
}

type notificationsLoadByStart struct {
	database NotificationLoader
	limit    int
	start    int64
}

type notificationsLoadByEnd struct {
	database NotificationLoader
	limit    int
	end      int64
}

type notificationsLoadByLabels struct {
	database NotificationLoader
	limit    int
	labels   []string
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

func (op notificationLoadBySlug) Execute() (contract.Notification, error) {
	res, err := op.database.GetNotificationBySlug(op.slug)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrNotificationNotFound(op.slug)
		}
		return res, err
	}
	return res, nil
}

func (op notificationsLoadBySender) Execute() ([]contract.Notification, error) {
	res, err := op.database.GetNotificationBySender(op.sender, op.limit)
	if err != nil {
		return res, err
	}
	if len(res) == 0 {
		return res, db.ErrNotFound
	}
	return res, nil
}

func (op notificationsLoadByStartEnd) Execute() ([]contract.Notification, error) {
	res, err := op.database.GetNotificationsByStartEnd(op.start, op.end, op.limit)
	if err != nil {
		return res, err
	}
	if len(res) == 0 {
		return res, db.ErrNotFound
	}
	return res, nil
}

func (op notificationsLoadByStart) Execute() ([]contract.Notification, error) {
	res, err := op.database.GetNotificationsByStart(op.start, op.limit)
	if err != nil {
		return res, err
	}
	if len(res) == 0 {
		return res, db.ErrNotFound
	}
	return res, nil
}

func (op notificationsLoadByEnd) Execute() ([]contract.Notification, error) {
	res, err := op.database.GetNotificationsByEnd(op.end, op.limit)
	if err != nil {
		return res, err
	}
	if len(res) == 0 {
		return res, db.ErrNotFound
	}
	return res, nil
}

func (op notificationsLoadByLabels) Execute() ([]contract.Notification, error) {
	n, err := op.database.GetNotificationsByLabels(op.labels, op.limit)
	if err != nil {
		return n, err
	}
	if len(n) == 0 {
		return n, db.ErrNotFound
	}
	return n, nil
}

func NewIdExecutor(db NotificationLoader, id string) IdExecutor {
	return notificationLoadById{
		database: db,
		id:       id,
	}
}

func NewSlugExecutor(db NotificationLoader, slug string) IdExecutor {
	return notificationLoadBySlug{
		database: db,
		slug:     slug,
	}
}

func NewSenderExecutor(db NotificationLoader, sender string, limit int) CollectionExecutor {
	return notificationsLoadBySender{
		database: db,
		limit:    limit,
		sender:   sender,
	}
}

func NewStartExecutor(db NotificationLoader, start int64, limit int) CollectionExecutor {
	return notificationsLoadByStart{
		database: db,
		limit:    limit,
		start:    start,
	}
}

func NewEndExecutor(db NotificationLoader, end int64, limit int) CollectionExecutor {
	return notificationsLoadByEnd{
		database: db,
		limit:    limit,
		end:      end,
	}
}

func NewStartEndExecutor(db NotificationLoader, start int64, end int64, limit int) CollectionExecutor {
	return notificationsLoadByStartEnd{
		database: db,
		limit:    limit,
		start:    start,
		end:      end,
	}
}

func NewLabelsExecutor(db NotificationLoader, labels []string, limit int) CollectionExecutor {
	return notificationsLoadByLabels{
		database: db,
		limit:    limit,
		labels:   labels,
	}
}
