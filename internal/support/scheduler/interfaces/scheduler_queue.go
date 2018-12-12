/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package interfaces

import(contract "github.com/edgexfoundry/edgex-go/pkg/models")

type SchedulerQueueClient interface {

	// **************************** INTERVAL ************************************

	// Return Interval by ID from the Scheduler Interval Context
	QueryIntervalByID(intervalId string) (contract.Interval, error)

	// Return Interval by Name from the Scheduler Interval Context
	QueryIntervalByName(intervalName string) (contract.Interval, error)

	// Add Interval into the Scheduler Queue
	AddIntervalToQueue(interval contract.Interval) error

	// Update Interval in the Scheduler Queue
	UpdateIntervalInQueue(interval contract.Interval) error

	// Remote the Interval from the Scheduler Queue
	RemoveIntervalInQueue(intervalId string) error

	// ************************* INTERVAL ACTIONS *******************************

	// Return IntervalAction by ID from the Scheduler IntervalAction Context
	QueryIntervalActionByID(intervalActionId string) (contract.IntervalAction, error)

	// Return IntervalAction by Name from the Scheduler IntervalAction Context
	QueryIntervalActionByName(intervalActionName string) (contract.IntervalAction, error)

	// Add IntervalAction into Scheduler Queue
	AddIntervalActionToQueue(intervalAction contract.IntervalAction) error

	// Update IntervalAction in the Scheduler Queue
	UpdateIntervalActionQueue(intervalAction contract.IntervalAction) error

	// Remove IntervalAction from the Scheduler Queue
	RemoveIntervalActionQueue(intervalActionId string) error

	// Check if we can connect to Scheduler Queue
	Connect()(string,error)
}