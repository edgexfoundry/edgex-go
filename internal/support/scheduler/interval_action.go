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
)

func addNewIntervalAction(intervalAction contract.IntervalAction) (string, error) {
	name := intervalAction.Name

	// Validate the IntervalAction is not in use
	ret, err := dbClient.IntervalActionByName(name)
	if err == nil && ret.Name == name {
		return "", errors.NewErrIntervalActionNameInUse(name)
	}

	// Validate the Target
	target := intervalAction.Target
	if target == "" {
		return "", errors.NewErrIntervalActionTargetNameRequired(intervalAction.ID)
	}

	// Validate the Interval
	interval := intervalAction.Interval
	if interval != "" {
		_, err := dbClient.IntervalByName(interval)
		if err != nil {
			return "", errors.NewErrIntervalNotFound(interval)
		}
	} else {
		return "", errors.NewErrIntervalNotFound(intervalAction.ID)
	}

	// Validate the IntervalAction does not exist in the scheduler queue
	retQ, err := scClient.QueryIntervalActionByName(name)
	if err == nil && retQ.Name == name {
		return "", errors.NewErrIntervalActionNameInUse(name)
	}

	// Add the new Interval Action to the database
	ID, err := dbClient.AddIntervalAction(intervalAction)
	if err != nil {
		return "", err
	}

	intervalAction.ID = ID

	// Add the new IntervalAction into scheduler queue
	err = scClient.AddIntervalActionToQueue(intervalAction)
	if err != nil {
		return "", err
	}

	return ID, nil
}

func updateIntervalAction(from contract.IntervalAction) error {
	to, err := dbClient.IntervalActionById(from.ID)
	if err != nil {
		// check by name
		_, err := dbClient.IntervalActionByName(from.Name)
		if err != nil {
			return errors.NewErrIntervalNotFound(from.ID)
		}
	}
	// Validate interval
	interval := from.Interval
	if interval != "" {
		_, err := dbClient.IntervalByName(interval)
		if err != nil {
			return errors.NewErrIntervalNotFound(interval)
		}
	}
	if interval != to.Interval {
		to.Interval = interval
	}

	// Name
	name := from.Name
	if name == "" {
		return errors.NewErrIntervalActionTargetNameRequired("")
	}
	// Ensure name is unique
	if name != to.Name {
		ret, err := dbClient.IntervalActionByName(name)
		if err == nil && ret.Name == name {
			return errors.NewErrIntervalActionNameInUse(name)
		}
		to.Name = name
	}

	// Validate target
	target := from.Target
	if target == "" {
		return errors.NewErrIntervalActionTargetNameRequired(from.ID)
	}
	if target != to.Target {
		to.Target = target
	}
	// Topic
	topic := from.Topic
	if topic != "" {
		to.Topic = from.Topic
	}
	// User
	user := from.User
	if user != to.User {
		to.User = user
	}
	// Publisher
	pub := from.Publisher
	if pub != to.Publisher {
		to.Publisher = pub
	}
	// Password
	pass := from.Password
	if pass != to.Password {
		to.Password = pass
	}
	// Port
	// TODO: Do we need a reasonable port restriction here?
	port := from.Port
	if port != to.Port {
		to.Port = port
	}
	// Address
	//TODO: Do we need a regex on a valid path sequence?
	address := from.Address
	if address != to.Address {
		to.Address = address
	}
	// HTTPMethod
	//TODO: Valid set of HTTP Verbs
	method := from.HTTPMethod
	if method != to.HTTPMethod {
		to.HTTPMethod = method
	}
	// Protocol
	//TODO: Valid protocol constraint?
	protocol := from.Protocol
	if protocol != to.Protocol {
		to.Protocol = protocol
	}
	// Parameters
	params := from.Parameters
	if params != to.Parameters {
		to.Parameters = params
	}

	// Validate the IntervalAction does not exist in the scheduler queue
	_, err = scClient.QueryIntervalActionByName(to.Name)
	if err == nil {
		// it's found we need to really update it
		err = scClient.UpdateIntervalActionQueue(to)
		if err != nil {
			return errors.NewErrIntervalActionNotFound(to.Name)
		}
	}
	return dbClient.UpdateIntervalAction(to)
}

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
