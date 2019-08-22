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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// DeleteExecutor handles the deletion of a interval.
// Returns ErrIntervalNotFound if an interval could not be found with a matching ID
type DeleteExecutor interface {
	Execute() error
}

type ScrubExecutor interface {
	Execute() (int, error)
}

type deleteIntervalByID struct {
	db        IntervalDeleter
	scDeleter SchedulerQueueDeleter
	did       string
}

type deleteIntervalByName struct {
	db        IntervalDeleter
	scDeleter SchedulerQueueDeleter
	dname     string
}

type scrubIntervals struct {
	db IntervalDeleter
}

// Execute performs the deletion of the interval.
func (dibi deleteIntervalByID) Execute() error {
	// check in memory first
	inMemory, err := dibi.scDeleter.QueryIntervalByID(dibi.did)
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

func (dibn deleteIntervalByName) Execute() error {
	// check in memory first
	inMemory, err := dibn.scDeleter.QueryIntervalByName(dibn.dname)
	if err != nil {
		return errors.NewErrIntervalNotFound(dibn.dname)
	}
	// remove in memory
	err = dibn.scDeleter.RemoveIntervalInQueue(inMemory.ID)
	if err != nil {
		return errors.NewErrIntervalNotFound(inMemory.ID)
	}
	// check if interval exist
	op := NewNameExecutor(dibn.db, dibn.dname)
	result, err := op.Execute()

	if err != nil {
		return err
	}
	if err = deleteInterval(result, dibn); err != nil {
		return err
	}
	return nil
}

func (si scrubIntervals) Execute() (int, error) {
	count, err := si.db.ScrubAllIntervals()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func deleteInterval(interval contract.Interval, dibn deleteIntervalByName) error {
	intervalActions, err := dibn.db.IntervalActionsByIntervalName(interval.Name)
	if err != nil {
		return err
	}
	if len(intervalActions) > 0 {
		return errors.NewErrIntervalStillInUse(interval.Name)
	}

	// Delete the interval
	if err = dibn.db.DeleteIntervalById(interval.ID); err != nil {
		return err
	}
	return nil

}

// NewDeleteByIDExecutor creates a new DeleteExecutor which deletes an interval based on id.
func NewDeleteByIDExecutor(db IntervalDeleter, scDeleter SchedulerQueueDeleter, did string) DeleteExecutor {
	return deleteIntervalByID{
		db:        db,
		scDeleter: scDeleter,
		did:       did,
	}
}

// NewDeleteByNameExecutor creates a new DeleteExecutor which deletes an interval based on name.
func NewDeleteByNameExecutor(db IntervalDeleter, scDeleter SchedulerQueueDeleter, dname string) DeleteExecutor {
	return deleteIntervalByName{
		db:        db,
		scDeleter: scDeleter,
		dname:     dname,
	}
}

// NewDeleteByScrubExecutor creates a new DeleteExecutor which scrubs intervals.
func NewScrubExecutor(db IntervalDeleter) ScrubExecutor {
	return scrubIntervals{
		db: db,
	}
}
