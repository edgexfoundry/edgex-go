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

import (
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

type DBClient interface {
	CloseSession()

	// **************************** INTERVAL ************************************

	// Return all the Interval(s)
	Intervals() ([]contract.Interval, error)

	// Return Interval(s) up to the number specified
	IntervalsWithLimit(limit int) ([]contract.Interval, error)

	// Return Interval by name
	IntervalByName(name string) (contract.Interval, error)

	// Return interval by contract id
	IntervalById(id string) (contract.Interval,error)

	// Add a new Interval
	AddInterval(interval contract.Interval) (string, error)

	// Update an Interval
	UpdateInterval(interval contract.Interval) error

	// Remove Interval by id
	DeleteIntervalById(id string) error

	// ************************* INTERVAL ACTIONS *******************************

	// Get all IntervalAction(s)
	IntervalActions() ([]contract.IntervalAction, error)

	// Return IntervalAction(s) up to the number specified
	IntervalActionsWithLimit(limit int) ([]contract.IntervalAction, error)

	// Get all IntervalAction(s) by interval name
	IntervalActionsByIntervalName(name string) ([]contract.IntervalAction, error)

	// Get all IntervalAction(s) by target name
	IntervalActionsByTarget(name string) ([]contract.IntervalAction, error)

	// Get IntervalAction by id
	IntervalActionById(id string) (contract.IntervalAction, error)

	// Get IntervalAction by name
	IntervalActionByName(name string) (contract.IntervalAction, error)

	// Add IntervalAction
	AddIntervalAction(intervalAction contract.IntervalAction) (string, error)

	// Update IntervalAction
	UpdateIntervalAction(intervalAction contract.IntervalAction) error

	// Remove IntervalAction by id
	DeleteIntervalActionById(id string) error

	// ************************** UTILITY FUNCTION(S) ***************************

	// Scrub all scheduler interval actions from the database data (only used in test)
	ScrubAllIntervalActions() (int, error)

	// Scrub all scheduler intervals from the database (only used in test)
	ScrubAllIntervals()(int, error)
}
