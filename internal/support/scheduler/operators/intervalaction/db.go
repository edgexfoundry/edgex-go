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
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/operators/interval"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// IntervalLoader provides functionality for obtaining Interval.
type IntervalActionLoader interface {
	IntervalActions() ([]contract.IntervalAction, error)
	IntervalActionsWithLimit(limit int) ([]contract.IntervalAction, error)
	IntervalActionByName(name string) (contract.IntervalAction, error)
	IntervalActionById(id string) (contract.IntervalAction, error)
	interval.IntervalLoader
}

// IntervalDeleter deletes interval.
type IntervalActionDeleter interface {
	IntervalActionLoader
}

// IntervalWriter adds interval.
type IntervalActionWriter interface {
	AddIntervalAction(interval contract.IntervalAction) (string, error)
	IntervalActionLoader
}

type IntervalActionUpdater interface {
	UpdateIntervalAction(intervalAction contract.IntervalAction) error
	IntervalActionLoader
}

type SchedulerQueueLoader interface {
	QueryIntervalActionByID(intervalActionId string) (contract.IntervalAction, error)
	QueryIntervalActionByName(intervalActionName string) (contract.IntervalAction, error)
}

// SchedulerQueueDeleter deletes interval from SchedulerQueue
type SchedulerQueueDeleter interface {
	RemoveIntervalActionInQueue(intervalId string) error
	SchedulerQueueLoader
}

// SchedulerQueueWriter adds interval in SchedulerQueue
type SchedulerQueueWriter interface {
	AddIntervalActionToQueue(interval contract.IntervalAction) error
	SchedulerQueueLoader
}

// SchedulerQueueUpdater update interval in SchedulerQueue
type SchedulerQueueUpdater interface {
	UpdateIntervalActionQueue(intervalAction contract.IntervalAction) error
	SchedulerQueueLoader
}
