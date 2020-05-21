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
	intervalLoader       IntervalLoader
	intervalActionLoader IntervalActionLoader
	intervalDeleter      IntervalDeleter
	sqDeleter            SchedulerQueueDeleter
	did                  string
}

type deleteIntervalByName struct {
	intervalLoader       IntervalLoader
	intervalActionLoader IntervalActionLoader
	intervalDeleter      IntervalDeleter
	sqDeleter            SchedulerQueueDeleter
	dname                string
}

type scrubIntervals struct {
	db IntervalDeleter
}

// Execute() deletes the interval by ID.
func (dibi deleteIntervalByID) Execute() error {
	// Check in memory.
	inMemory, err := dibi.sqDeleter.QueryIntervalByID(dibi.did)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(dibi.did)
		}
		return err
	}

	return deleteInterval(inMemory, dibi.intervalDeleter, dibi.sqDeleter)
}

// Execute() deletes the interval by Name.
func (dibn deleteIntervalByName) Execute() error {
	// Check in memory.
	inMemory, err := dibn.sqDeleter.QueryIntervalByName(dibn.dname)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(dibn.dname)
		}
		return err
	}

	return deleteInterval(inMemory, dibn.intervalDeleter, dibn.sqDeleter)
}

// deleteInterval first checks the Interval to determine that it is not in use before deleting
// from both memory and database. Note that a failure in this function may result in the system
// ending up in an undesirable state, and "rollbacks" are not handled here. For example, if we
// first remove the Interval from memory, then encounter failure while trying to delete from the
// database, we will end up with GET /interval API calls still responding with the "deleted" Interval.
func deleteInterval(
	interval contract.Interval,
	intervalDeleter IntervalDeleter,
	sqDeleter SchedulerQueueDeleter) error {

	// Check if interval is in use. Get all IntervalActions that are associated with this interval.
	allIntervalActions, err := intervalDeleter.IntervalActionsByIntervalName(interval.Name)
	if err != nil {
		return err
	}

	if len(allIntervalActions) != 0 {
		return errors.NewErrIntervalNameInUse(interval.Name)
	}

	// Remove interval in memory
	err = sqDeleter.RemoveIntervalInQueue(interval.ID)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(interval.ID)
		}
		return err
	}

	// Delete the interval
	if err = intervalDeleter.DeleteIntervalById(interval.ID); err != nil {
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

// NewDeleteByIDExecutor creates a new DeleteExecutor which deletes an interval based on id.
func NewDeleteByIDExecutor(
	intervalDeleter IntervalDeleter,
	sqDeleter SchedulerQueueDeleter,
	did string) DeleteExecutor {

	return deleteIntervalByID{
		intervalDeleter: intervalDeleter,
		sqDeleter:       sqDeleter,
		did:             did,
	}
}

// NewDeleteByNameExecutor creates a new DeleteExecutor which deletes an interval based on name.
func NewDeleteByNameExecutor(
	intervalDeleter IntervalDeleter,
	sqDeleter SchedulerQueueDeleter,
	dname string) DeleteExecutor {

	return deleteIntervalByName{
		intervalDeleter: intervalDeleter,
		sqDeleter:       sqDeleter,
		dname:           dname,
	}
}

// NewDeleteByScrubExecutor creates a new DeleteExecutor which scrubs intervals.
func NewScrubExecutor(db IntervalDeleter) ScrubExecutor {
	return scrubIntervals{
		db: db,
	}
}
