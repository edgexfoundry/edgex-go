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
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/errors"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/robfig/cron"
)

func getIntervals(limit int) ([]contract.Interval, error) {
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

func addNewInterval(interval contract.Interval) (string, error) {
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
		if !isFrequencyValid(interval.Frequency) {
			return "", errors.NewErrInvalidFrequencyFormat(freq)
		}
	}

	// Validate that interval is not in queue
	ret, err = scClient.QueryIntervalByName(name)
	if err == nil && ret.Name == name {
		return "", errors.NewErrIntervalNameInUse(name)
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
		//failed to add to scheduler queue
		LoggingClient.Error(err.Error())
	}
	return ID, nil
}

func updateInterval(from contract.Interval) error {
	to, err := dbClient.IntervalById(from.ID)
	if err != nil {
		// Check by name
		_, err := dbClient.IntervalByName(from.Name)
		if err != nil {
			return errors.NewErrIntervalNotFound(from.ID)
		}
	}
	// Update the fields
	if from.Cron != "" {
		if _, err := cron.Parse(from.Cron); err != nil {
			return errors.NewErrInvalidCronFormat(from.Cron)
		}
		to.Cron = from.Cron
	}
	if from.End != "" {
		if _, err := msToTime(from.End); err != nil {
			return errors.NewErrInvalidTimeFormat(from.End)
		}
		to.End = from.End
	}
	if from.Frequency != "" {
		if !isFrequencyValid(from.Frequency) {
			return errors.NewErrInvalidFrequencyFormat(from.Frequency)
		}
		to.Frequency = from.Frequency
	}
	if from.Start != "" {
		if _, err := msToTime(from.Start); err != nil {
			return errors.NewErrInvalidTimeFormat(from.Start)
		}
		to.Start = from.Start
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}
	// Check if new name is unique
	if from.Name != "" && from.Name != to.Name {
		checkInterval, err := dbClient.IntervalByName(from.Name)
		if err != nil {
			if checkInterval.ID != "" {
				return errors.NewErrIntervalNameInUse(from.Name)
			}
		}

		// Check if the interval still has attached interval actions
		stillInUse, err := isIntervalStillInUse(to)
		if err != nil {
			return err
		}
		if stillInUse {
			return errors.NewErrIntervalStillInUse(to.Name)
		}
		to.Name = from.Name
	}

	err = scClient.UpdateIntervalInQueue(to)
	if err != nil {
		//failed to update the scheduler queue
		LoggingClient.Error(err.Error())
		return err
	}

	return dbClient.UpdateInterval(to)
}

func getIntervalById(id string) (contract.Interval, error) {
	interval, err := dbClient.IntervalById(id)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(id)
		}
		return contract.Interval{}, err
	}
	return interval, nil
}
func getIntervalByName(name string) (interval contract.Interval, err error) {
	interval, err = dbClient.IntervalByName(name)
	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalNotFound(name)
		}
		return contract.Interval{}, err
	}
	return interval, nil
}
func deleteIntervalByName(name string) error {

	// check in memory first
	inMemory, err := scClient.QueryIntervalByName(name)
	if err != nil {
		return errors.NewErrIntervalNotFound(name)
	}
	// remove in memory
	err = scClient.RemoveIntervalInQueue(inMemory.ID)
	if err != nil {
		return errors.NewErrDbNotFound()
	}
	// check if interval exists
	interval, err := getIntervalByName(name)
	if err != nil {
		return err
	}

	if err = deleteInterval(interval); err != nil {
		return err
	}
	return nil
}

func deleteIntervalById(id string) error {

	// check in memory first
	inMemory, err := scClient.QueryIntervalByID(id)
	if err != nil {
		return errors.NewErrIntervalNotFound(id)
	}
	// remove in memory
	err = scClient.RemoveIntervalInQueue(inMemory.ID)
	if err != nil {
		return errors.NewErrDbNotFound()
	}
	// check if interval exists
	interval, err := getIntervalById(id)
	if err != nil {
		return err
	}

	if err = deleteInterval(interval); err != nil {
		return err
	}
	return nil
}

func deleteInterval(interval contract.Interval) error {
	intervalActions, err := dbClient.IntervalActionsByIntervalName(interval.Name)
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	if len(intervalActions) > 0 {
		LoggingClient.Error("Data integrity issue.  Interval is still referenced by existing Interval Actions.")
		return errors.NewErrIntervalStillInUse(interval.Name)
	}

	// Delete the interval
	if err = dbClient.DeleteIntervalById(interval.ID); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	return nil
}

// Helper function
func isIntervalStillInUse(s contract.Interval) (bool, error) {
	var intervalActions []contract.IntervalAction

	intervalActions, err := dbClient.IntervalActionsByIntervalName(s.Name)
	if err != nil {
		return false, err
	}
	if len(intervalActions) > 0 {
		return true, nil
	}

	return false, nil
}

func scrubAll() (int, error) {
	LoggingClient.Info("Scrubbing All Interval(s) and IntervalAction(s).")

	count, err := dbClient.ScrubAllIntervals()
	if err != nil {
		LoggingClient.Error(err.Error())
		return 0, err
	}
	return count, nil
}
