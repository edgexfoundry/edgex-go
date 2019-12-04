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

package scheduler

import (
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
)

func getIntervals(limit int, dbClient interfaces.DBClient) ([]contract.Interval, error) {
	var err error
	var intervals []contract.Interval

	if limit <= 0 {
		intervals, err = dbClient.Intervals()
	} else {
		intervals, err = dbClient.IntervalsWithLimit(limit)
	}

	if err != nil {
		return nil, err
	}

	return intervals, err
}

func addNewInterval(
	interval contract.Interval,
	dbClient interfaces.DBClient,
	scClient interfaces.SchedulerQueueClient) (string, error) {

	name := interval.Name

	// Check if the name is unique
	ret, err := dbClient.IntervalByName(name)
	if err == nil && ret.Name == name {
		return "", errors.NewErrIntervalNameInUse(name)
	}

	// Validate the Start time format
	start := interval.Start
	if start != "" {
		if _, err := msToTime(start); err != nil {
			return "", errors.NewErrInvalidTimeFormat(start)
		}
	}
	// Validate the End time format
	end := interval.End
	if end != "" {
		if _, err := msToTime(end); err != nil {
			return "", errors.NewErrInvalidTimeFormat(end)
		}
	}
	// Validate the Frequency
	freq := interval.Frequency
	if freq != "" {
		_, err := parseFrequency(freq)
		if err != nil {
			return "", errors.NewErrInvalidFrequencyFormat(freq)
		}
	}

	// Add the new interval to the database
	ID, err := dbClient.AddInterval(interval)
	if err != nil {
		return "", err
	}

	// Push the new interval into scheduler queue
	interval.ID = ID
	err = scClient.AddIntervalToQueue(interval)
	if err != nil {
		return "", err
	}

	return ID, nil
}

func getIntervalByName(name string, dbClient interfaces.DBClient) (interval contract.Interval, err error) {
	interval, err = dbClient.IntervalByName(name)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(name)
		}

		return contract.Interval{}, err
	}

	return interval, nil
}
