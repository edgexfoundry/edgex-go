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
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

// Utility function for adding configured locally intervals and scheduled events
func LoadScheduler() error {

	// ensure maps are clean
	clearMaps()

	// ensure queue is empty
	clearQueue()

	LoggingClient.Info("loading intervals, interval actions ...")

	// load data from support-scheduler database
	err := loadSupportSchedulerDBInformation()
	if err != nil {
		LoggingClient.Error("failed to load information from support-scheduler:" + err.Error())
		return err
	}

	// load config intervals
	errLCI := loadConfigIntervals()
	if errLCI != nil {
		LoggingClient.Error("failed to load scheduler config data:" + errLCI.Error())
		return errLCI
	}

	// load config interval actions
	errLCA := loadConfigIntervalActions()
	if errLCA != nil {
		LoggingClient.Error("failed to load interval actions config data:" + errLCA.Error())
		return errLCA
	}

	LoggingClient.Info("finished loading intervals, interval actions")

	return nil
}

// Query support-scheduler scheduler client get intervals
func getSchedulerDBIntervals() ([]contract.Interval, error) {
	var err error
	var intervals []contract.Interval

	intervals, err = dbClient.Intervals()

	if err != nil {
		LoggingClient.Error("failed connecting to metadata and retrieving intervals:" + err.Error())
		return intervals, err
	}

	if intervals != nil {
		LoggingClient.Debug("successfully queried support-scheduler intervals...")
		for _, v := range intervals {
			LoggingClient.Debug("found interval", "name", v.Name, "id", v.ID, "start", v.Start)
		}
	}
	return intervals, nil
}

// Query support-scheduler schedulerEvent client get scheduledEvents
func getSchedulerDBIntervalActions() ([]contract.IntervalAction, error) {
	var err error
	var intervalActions []contract.IntervalAction

	intervalActions, err = dbClient.IntervalActions()
	if err != nil {
		LoggingClient.Error("error connecting to metadata and retrieving interval actions:" + err.Error())
		return intervalActions, err
	}

	// debug information only
	if intervalActions != nil {
		LoggingClient.Debug("successfully queried support-scheduler interval actions...")
		for _, v := range intervalActions {
			LoggingClient.Debug("found interval action", "name", v.Name, "id", v.ID, "interval", v.Interval, "target", v.Target)
		}
	}
	return intervalActions, nil
}

// Iterate over the received intervals add them to scheduler memory queue
func addReceivedIntervals(intervals []contract.Interval) error {
	for _, interval := range intervals {
		err := scClient.AddIntervalToQueue(interval)
		if err != nil {
			LoggingClient.Info("problem adding support-scheduler interval name: %s - %s", interval.Name, err.Error())
			return err
		}
		LoggingClient.Info("added interval", "name", interval.Name, "id", interval.ID)
	}
	return nil
}

// Iterate over the received interval action(s)
func addReceivedIntervalActions(intervalActions []contract.IntervalAction) error {
	for _, intervalAction := range intervalActions {
		err := scClient.AddIntervalActionToQueue(intervalAction)
		if err != nil {
			LoggingClient.Info("problem adding support-scheduler interval action", "name:", intervalAction.Name, "message", err.Error())
			return err
		}
		LoggingClient.Info("added interval action", "name", intervalAction.Name, "id", intervalAction.ID)
	}
	return nil
}

// Add interval to support-scheduler
func addIntervalToSchedulerDB(interval contract.Interval) (string, error) {

	var err error
	var id string

	id, err = dbClient.AddInterval(interval)
	if err != nil {
		LoggingClient.Error("problem trying to add interval to support-scheduler service:" + err.Error())
		return "", err
	}
	interval.ID = id

	LoggingClient.Info("added interval to the support-scheduler database", "name", interval.Name, "id", ID)

	return id, nil
}

// Add interval event to support-scheduler
func addIntervalActionToSchedulerDB(intervalAction contract.IntervalAction) (string, error) {
	var err error
	var id string

	id, err = dbClient.AddIntervalAction(intervalAction)
	if err != nil {
		LoggingClient.Error("problem trying to add interval action to support-scheduler service:" + err.Error())
		return "", err
	}
	LoggingClient.Info("added interval action to the support-scheduler", "name", intervalAction.Name, "id", id)

	return id, nil
}

// Load intervals
func loadConfigIntervals() error {

	intervals := Configuration.Intervals
	for i := range intervals {
		interval := contract.Interval{
			ID:        "",
			Created:   0,
			Modified:  0,
			Origin:    0,
			Name:      intervals[i].Name,
			Start:     intervals[i].Start,
			End:       intervals[i].End,
			Frequency: intervals[i].Frequency,
			Cron:      intervals[i].Cron,
			RunOnce:   intervals[i].RunOnce,
		}

		// query scheduler service for interval in memory queue
		_, errExistingSchedule := scClient.QueryIntervalByName(interval.Name)

		if errExistingSchedule != nil {
			// add the interval support-scheduler
			newIntervalID, errAddedInterval := addIntervalToSchedulerDB(interval)
			if errAddedInterval != nil {
				LoggingClient.Error("problem adding interval to the scheduler database:" + errAddedInterval.Error())
				return errAddedInterval
			}

			// add the support-scheduler scheduler.id
			interval.ID = newIntervalID

			// add the interval to the scheduler
			err := scClient.AddIntervalToQueue(interval)

			if err != nil {
				LoggingClient.Error("problem loading interval from the scheduler config: " + err.Error())
				return err
			}
		} else {
			LoggingClient.Debug("did not add interval as it already exists in the scheduler database", "name", interval.Name)
		}
	}

	return nil
}

// Load interval actions if required
func loadConfigIntervalActions() error {

	intervalActions := Configuration.IntervalActions

	for ia := range intervalActions {
		intervalAction := contract.IntervalAction{
			Name:       intervalActions[ia].Name,
			Interval:   intervalActions[ia].Interval,
			Parameters: intervalActions[ia].Parameters,
			Target:     intervalActions[ia].Target,
			Path:       intervalActions[ia].Path,
			Port:       intervalActions[ia].Port,
			Protocol:   intervalActions[ia].Protocol,
			HTTPMethod: intervalActions[ia].Method,
			Address:    intervalActions[ia].Host,
		}

		// query scheduler in memory queue and determine of intervalAction exists
		_, err := scClient.QueryIntervalActionByName(intervalAction.Name)

		if err != nil {

			// add the interval action to support-scheduler database
			newIntervalActionID, err := addIntervalActionToSchedulerDB(intervalAction)
			if err != nil {
				LoggingClient.Error("problem adding interval action into support-scheduler database:" + err.Error())
				return err
			}

			// add the support-scheduler version of the intervalAction.ID
			intervalAction.ID = newIntervalActionID
			//TODO: Do we care about the Created,Modified, or Origin fields?

			errAddIntervalAction := scClient.AddIntervalActionToQueue(intervalAction)
			if errAddIntervalAction != nil {
				LoggingClient.Error("problem loading interval action into support-scheduler:" + errAddIntervalAction.Error())
				return errAddIntervalAction

			}
		} else {
			LoggingClient.Debug("did not load interval action as it exists in the scheduler database:" + intervalAction.Name)
		}
	}
	return nil
}

// Query support-scheduler database information
func loadSupportSchedulerDBInformation() error {

	receivedIntervals, err := getSchedulerDBIntervals()
	if err != nil {
		LoggingClient.Error("failed to receive intervals from support-scheduler database:" + err.Error())
		return err
	}

	err = addReceivedIntervals(receivedIntervals)
	if err != nil {
		LoggingClient.Error("failed to add received intervals from support-scheduler database:" + err.Error())
		return err
	}

	intervalActions, err := getSchedulerDBIntervalActions()
	if err != nil {
		LoggingClient.Error("failed to receive interval actions from support-scheduler database:" + err.Error())
		return err
	}

	err = addReceivedIntervalActions(intervalActions)
	if err != nil {
		LoggingClient.Error("failed to add received interval actions from support-scheduler database:" + err.Error())
		return err
	}

	return nil
}
