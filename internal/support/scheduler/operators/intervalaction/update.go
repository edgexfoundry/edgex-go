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
package intervalaction

import (
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type UpdateExecutor interface {
	Execute() error
}

type intervalActionUpdate struct {
	database       IntervalActionUpdater
	scClient       SchedulerQueueUpdater
	intervalAction contract.IntervalAction
}

// This method updates the provided Addressable in the database.
func (op intervalActionUpdate) Execute() error {
	err := updateIntervalAction(op)
	if err != nil {
		return err
	}
	return nil
}

// This factory method returns an executor used to update an addressable.
func NewUpdateExecutor(database IntervalActionUpdater, scClient SchedulerQueueUpdater, intervalAction contract.IntervalAction) UpdateExecutor {
	return intervalActionUpdate{
		database:       database,
		scClient:       scClient,
		intervalAction: intervalAction,
	}
}

func updateIntervalAction(iau intervalActionUpdate) error {
	to, err := iau.database.IntervalActionById(iau.intervalAction.ID)
	if err != nil {
		// check by name
		_, err := iau.database.IntervalActionByName(iau.intervalAction.Name)
		if err != nil {
			return errors.NewErrIntervalActionNotFound(iau.intervalAction.ID)
		}
	}
	// Validate interval
	interval := iau.intervalAction.Interval
	if interval != "" {
		_, err := iau.database.IntervalByName(interval)
		if err != nil {
			return errors.NewErrIntervalNotFound(interval)
		}
	}

	// Name
	name := iau.intervalAction.Name
	if name == "" {
		return errors.NewErrIntervalActionTargetNameRequired("")
	}
	// Ensure name is unique
	if name != to.Name {
		ret, err := iau.database.IntervalActionByName(name)
		if err == nil && ret.Name == name {
			return errors.NewErrIntervalActionNameInUse(name)
		}
	}

	// Validate target
	target := iau.intervalAction.Target
	if target == "" {
		return errors.NewErrIntervalActionTargetNameRequired(iau.intervalAction.ID)
	}
	to = iau.intervalAction

	err = iau.scClient.UpdateIntervalActionQueue(to)
	if err != nil {
		return errors.NewErrIntervalActionNotFound(to.Name)
	}

	return iau.database.UpdateIntervalAction(to)
}
