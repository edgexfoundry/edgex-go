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
	"github.com/robfig/cron"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type UpdateExecutor interface {
	Execute() error
}

type intervalUpdate struct {
	database IntervalUpdater
	scClient SchedulerQueueUpdater
	interval contract.Interval
}

// This method updates the provided Addressable in the database.
func (op intervalUpdate) Execute() error {
	to, err := op.database.IntervalById(op.interval.ID)
	if err != nil {
		// Check by name
		to, err = op.database.IntervalByName(op.interval.Name)
		if err != nil {
			return errors.NewErrIntervalNotFound(op.interval.ID)
		}
	}
	// Update the fields
	if op.interval.Cron != "" {
		if _, err := cron.Parse(op.interval.Cron); err != nil {
			return errors.NewErrInvalidCronFormat(op.interval.Cron)
		}
		to.Cron = op.interval.Cron
	}
	if op.interval.Timestamps.Origin != 0 {
		to.Timestamps.Origin = op.interval.Timestamps.Origin
	}
	// Check if new name is unique
	if op.interval.Name != "" && op.interval.Name != to.Name {
		checkInterval, err := op.database.IntervalByName(op.interval.Name)
		// Check for error other than not found
		if err != nil && err != db.ErrNotFound {
			return err
		}
		// Check if interval with new name exists
		if checkInterval.ID != "" {
			return errors.NewErrIntervalNameInUse(op.interval.Name)
		}
		// Check if the interval still has attached interval actions
		stillInUse, err := op.isIntervalStillInUse(to)
		if err != nil {
			return err
		}
		if stillInUse {
			return errors.NewErrIntervalStillInUse(to.Name)
		}
	}
	op.interval.ID = to.ID
	err = op.scClient.UpdateIntervalInQueue(op.interval)
	if err != nil {
		return err
	}

	return op.database.UpdateInterval(op.interval)
}

// This factory method returns an executor used to update an addressable.
func NewUpdateExecutor(database IntervalUpdater, scClient SchedulerQueueUpdater, interval contract.Interval) UpdateExecutor {
	return intervalUpdate{
		database: database,
		scClient: scClient,
		interval: interval,
	}
}

// Helper function
func (op intervalUpdate) isIntervalStillInUse(s contract.Interval) (bool, error) {
	var intervalActions []contract.IntervalAction

	intervalActions, err := op.database.IntervalActionsByIntervalName(s.Name)
	if err != nil {
		return false, err
	}
	if len(intervalActions) > 0 {
		return true, nil
	}

	return false, nil
}
