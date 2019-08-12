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
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type AddExecutor interface {
	Execute() (string, error)
}

type intervalAdd struct {
	database IntervalWriter
	scClient SchedulerQueueWriter
	interval contract.Interval
}

// This method adds the provided Addressable to the database.
func (op intervalAdd) Execute() (id string, err error) {
	name := op.interval.Name

	// Check if the name is unique
	ret, err := op.database.IntervalByName(name)
	if err == nil && ret.Name == name {
		return "", errors.NewErrIntervalNameInUse(name)
	}
	// Add the new interval to the database
	ID, err := op.database.AddInterval(op.interval)
	if err != nil {
		return "", err
	}

	// Push the new interval into scheduler queue
	op.interval.ID = ID
	err = op.scClient.AddIntervalToQueue(op.interval)
	if err != nil {
		return ID, err
	}
	return ID, nil
}

// This factory method returns an executor used to add an addressable.
func NewAddExecutor(db IntervalWriter, scClient SchedulerQueueWriter, interval contract.Interval) AddExecutor {
	return intervalAdd{
		database: db,
		scClient: scClient,
		interval: interval,
	}
}
