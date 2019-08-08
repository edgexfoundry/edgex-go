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

package subscription

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
)

// DeleteExecutor handles the deletion of a subscription.
// Returns ErrNSubscriptionNotFound if a subscription could not be found with a matching ID
type DeleteExecutor interface {
	Execute() error
}

type deleteSubscriptionByID struct {
	db  SubscriptionDeleter
	did string
}

// Execute performs the deletion of the notification.
func (dnbi deleteSubscriptionByID) Execute() error {
	err := dnbi.db.DeleteSubscriptionById(dnbi.did)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrSubscriptionNotFound(dnbi.did)
		}
		return err
	}
	return nil
}

// NewDeleteByIDExecutor creates a new DeleteExecutor which deletes a subscription based on id.
func NewDeleteByIDExecutor(db SubscriptionDeleter, did string) DeleteExecutor {
	return deleteSubscriptionByID{
		db:  db,
		did: did,
	}
}
