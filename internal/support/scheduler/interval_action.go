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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func getIntervalActionById(id string) (contract.IntervalAction, error) {
	intervalAction, err := dbClient.IntervalActionById(id)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalActionNotFound(id)
		}
		return contract.IntervalAction{}, err
	}
	return intervalAction, nil
}

func getIntervalActions(limit int) ([]contract.IntervalAction, error) {
	var err error
	var intervalActions []contract.IntervalAction

	if limit <= 0 {
		intervalActions, err = dbClient.IntervalActions()
	} else {
		intervalActions, err = dbClient.IntervalActionsWithLimit(limit)
	}

	if err != nil {
		return nil, err
	}

	return intervalActions, err
}

func getIntervalActionByName(name string) (contract.IntervalAction, error) {
	intervalAction, err := dbClient.IntervalActionByName(name)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrIntervalActionNotFound(name)
		}
		return contract.IntervalAction{}, err
	}
	return intervalAction, nil
}

func getIntervalActionsByTarget(target string) ([]contract.IntervalAction, error) {
	intervalActions, err := dbClient.IntervalActionsByTarget(target)
	if err != nil {
		return []contract.IntervalAction{}, err
	}
	return intervalActions, err
}

func getIntervalActionsByInterval(interval string) ([]contract.IntervalAction, error) {
	intervalActions, err := dbClient.IntervalActionsByIntervalName(interval)
	if err != nil {
		return []contract.IntervalAction{}, err
	}
	return intervalActions, err
}

func deleteIntervalActionById(id string) error {

	// check in memory first
	inMemory, err := scClient.QueryIntervalActionByID(id)
	if err != nil {
		return errors.NewErrIntervalNotFound(id)
	}
	// remove in memory
	err = scClient.RemoveIntervalActionQueue(inMemory.ID)
	if err != nil {
		return errors.NewErrDbNotFound()
	}

	// check in DB
	intervalAction, err := getIntervalActionById(id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrIntervalNotFound(intervalAction.Name)
		} else {
			return err
		}
	}

	// remove from DB
	if err = deleteIntervalAction(intervalAction); err != nil {
		return err
	}
	return nil
}

func deleteIntervalActionByName(name string) error {
	// check in memory first
	inMemory, err := scClient.QueryIntervalActionByName(name)
	if err != nil {
		return errors.NewErrIntervalNotFound(name)
	}
	// remove in memory
	err = scClient.RemoveIntervalActionQueue(inMemory.ID)
	if err != nil {
		return errors.NewErrDbNotFound()
	}

	intervalAction, err := getIntervalActionByName(name)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrIntervalNotFound(intervalAction.Name)
		} else {
			return err
		}
	}
	if err = deleteIntervalAction(intervalAction); err != nil {
		return err
	}
	return nil
}

func deleteIntervalAction(intervalAction contract.IntervalAction) error {
	if err := dbClient.DeleteIntervalActionById(intervalAction.ID); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	return nil
}

func scrubAllInteralActions() (int, error) {
	LoggingClient.Info("Scrubbing All IntervalAction(s).")

	count, err := dbClient.ScrubAllIntervalActions()
	if err != nil {
		LoggingClient.Error(err.Error())
		return 0, err
	}
	return count, nil
}
