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

type AddExecutor interface {
	Execute() (string, error)
}

type intervalActionAdd struct {
	database       IntervalActionWriter
	intervalAction contract.IntervalAction
	scClient       SchedulerQueueWriter
}

// This method adds the provided Addressable to the database.
func (op intervalActionAdd) Execute() (id string, err error) {
	newId, err := addNewIntervalAction(op)
	if err != nil {
		return newId, err
	}
	return newId, nil
}

// This factory method returns an executor used to add an addressable.
func NewAddExecutor(db IntervalActionWriter, scClient SchedulerQueueWriter, intervalAction contract.IntervalAction) AddExecutor {
	return intervalActionAdd{
		database:       db,
		scClient:       scClient,
		intervalAction: intervalAction,
	}
}

func addNewIntervalAction(iaa intervalActionAdd) (string, error) {
	name := iaa.intervalAction.Name

	// Validate the IntervalAction is not in use
	ret, err := iaa.database.IntervalActionByName(name)
	if err == nil && ret.Name == name {
		return "", errors.NewErrIntervalActionNameInUse(name)
	}

	// Validate the Target
	target := iaa.intervalAction.Target
	if target == "" {
		return "", errors.NewErrIntervalActionTargetNameRequired(iaa.intervalAction.ID)
	}

	// Validate the Interval
	interval := iaa.intervalAction.Interval
	if interval != "" {
		_, err := iaa.database.IntervalByName(interval)
		if err != nil {
			return "", errors.NewErrIntervalNotFound(interval)
		}
	} else {
		return "", errors.NewErrIntervalNotFound(iaa.intervalAction.ID)
	}

	// Validate the IntervalAction does not exist in the scheduler queue
	retQ, err := iaa.scClient.QueryIntervalActionByName(name)
	if err == nil && retQ.Name == name {
		return "", errors.NewErrIntervalActionNameInUse(name)
	}

	// Add the new Interval Action to the database
	ID, err := iaa.database.AddIntervalAction(iaa.intervalAction)
	if err != nil {
		return "", err
	}

	iaa.intervalAction.ID = ID

	// Add the new IntervalAction into scheduler queue
	err = iaa.scClient.AddIntervalActionToQueue(iaa.intervalAction)
	if err != nil {
		return "", err
	}

	return ID, nil
}
