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

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

// IntervalLoader provides functionality for obtaining Interval.
type IntervalLoader interface {
	IntervalById(id string) (contract.Interval, error)
	IntervalByName(name string) (contract.Interval, error)
}

// IntervalDeleter deletes interval.
type IntervalDeleter interface {
	DeleteIntervalById(id string) error
	IntervalLoader
	IntervalActionLoader
}

// SchedulerQueueLoader provides functionality for obtaining Interval from SchedulerQueue
type SchedulerQueueLoader interface {
	QueryIntervalByID(intervalId string) (contract.Interval, error)
	QueryIntervalByName(intervalName string) (contract.Interval, error)
}

// SchedulerQueueDeleter deletes interval from SchedulerQueue
type SchedulerQueueDeleter interface {
	RemoveIntervalInQueue(intervalId string) error
	SchedulerQueueLoader
}

type IntervalActionLoader interface {
	IntervalActionsByIntervalName(name string) ([]contract.IntervalAction, error)
}
