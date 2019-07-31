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

package interval

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
)

// DeleteExecutor handles the deletion of a interval.
// Returns ErrIntervalNotFound if an interval could not be found with a matching ID
type DeleteExecutor interface {
	Execute() error
}

type deleteIntervalByID struct {
	db        IntervalDeleter
	scLoader  SchedulerQueueLoader
	scDeleter SchedulerQueueDeleter
	did       string
}

// Execute performs the deletion of the interval.
func (dibi deleteIntervalByID) Execute() error {
	// check in memory first
	inMemory, err := dibi.scLoader.QueryIntervalByID(dibi.did)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(dibi.did)
		}
		return err
	}

	if err = dibi.db.DeleteIntervalById(dibi.did); err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(dibi.did)
		}
		return err
	}

	// remove in memory
	err = dibi.scDeleter.RemoveIntervalInQueue(inMemory.ID)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(dibi.did)
		}
		return err
	}

	return nil
}

// NewDeleteByIDExecutor creates a new DeleteExecutor which deletes a interval based on id.
func NewDeleteByIDExecutor(db IntervalDeleter, scLoader SchedulerQueueLoader, scDeleter SchedulerQueueDeleter, did string) DeleteExecutor {
	return deleteIntervalByID{
		db:        db,
		scLoader:  scLoader,
		scDeleter: scDeleter,
		did:       did,
	}
}
