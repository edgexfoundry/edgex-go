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
)

// DeleteExecutor handles the deletion of a notification.
// Returns ErrNotificationNotFound if a notification could not be found with a matching ID
type DeleteExecutor interface {
	Execute() error
}

type deleteNotificationByID struct {
	db  NotificationDeleter
	did string
}

type deleteNotificationBySlug struct {
	db    NotificationDeleter
	dslug string
}

type deleteNotificationsByAge struct {
	db   NotificationDeleter
	dage int
}

// Execute performs the deletion of the notification.
func (dnbi deleteNotificationByID) Execute() error {
	err := dnbi.db.DeleteNotificationById(dnbi.did)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrNotificationNotFound(dnbi.did)
		}
		return err
	}
	return nil
}

func (dnbs deleteNotificationBySlug) Execute() error {
	err := dnbs.db.DeleteNotificationBySlug(dnbs.dslug)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrNotificationNotFound(dnbs.dslug)
		}
		return err
	}
	return nil
}

func (dnba deleteNotificationsByAge) Execute() error {
	err := dnba.db.DeleteNotificationsOld(dnba.dage)
	if err != nil {
		return err
	}
	return nil
}

// NewDeleteByIDExecutor creates a new DeleteExecutor which deletes a notifcation based on id.
func NewDeleteByIDExecutor(db NotificationDeleter, did string) DeleteExecutor {
	return deleteNotificationByID{
		db:  db,
		did: did,
	}
}

// NewDeleteBySlugExecutor creates a new DeleteExecutor which deletes a notifcation based on slug.
func NewDeleteBySlugExecutor(db NotificationDeleter, dslug string) DeleteExecutor {
	return deleteNotificationBySlug{
		db:    db,
		dslug: dslug,
	}
}

// NewDeleteBySlugExecutor creates a new DeleteExecutor which deletes a notifcation based on slug.
func NewDeleteByAgeExecutor(db NotificationDeleter, dage int) DeleteExecutor {
	return deleteNotificationsByAge{
		db:   db,
		dage: dage,
	}
}
