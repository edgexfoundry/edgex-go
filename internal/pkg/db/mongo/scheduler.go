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
package mongo

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo/models"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
)

// ******************************* INTERVALS **********************************

// Return all the Interval(s)
// UnexpectedError - failed to retrieve intervals from the database
// Sort the events in descending order by ID

func (mc MongoClient) Intervals() ([]contract.Interval, error) {
	return mapIntervals(mc.getIntervals(bson.M{}))
}

// Return Interval(s) up to the max number specified
// UnexpectedError - failed to retrieve intervals from the database
// Sort the intervals in descending order by ID
func (mc MongoClient) IntervalsWithLimit(limit int) ([]contract.Interval, error) {
	return mapIntervals(mc.getIntervalsLimit(bson.M{}, limit))
}

// Return an Interval by name
// UnexpectedError - failed to retrieve interval from the database
func (mc MongoClient) IntervalByName(name string) (contract.Interval, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	query := bson.M{"name": name}

	mi := models.Interval{}
	err := s.DB(mc.database.Name).C(db.Interval).Find(query).One(&mi)
	if err != nil {
		err = errorMap(err)
		return contract.Interval{}, err
	}
	return mi.ToContract(), nil
}

// Return an Interval by ID
// UnexpectedError - failed to retrieve interval from the database
func (mc MongoClient) IntervalById(id string) (contract.Interval, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// Id is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return contract.Interval{}, db.ErrInvalidObjectId
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}

	interval := models.Interval{}
	err := s.DB(mc.database.Name).C(db.Interval).Find(query).One(&interval)
	if err != nil {
		err = errorMap(err)
		return contract.Interval{}, err
	}

	return interval.ToContract(), err
}

// Add an Interval
// UnexpectedError - failed to add interval into  the database
func (mc MongoClient) AddInterval(interval contract.Interval) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	mapped := &models.Interval{}
	err := mapped.FromContract(interval)
	if err != nil {
		return interval.ID, err
	}

	// See if the name is unique and add the value descriptors
	found, err := s.DB(mc.database.Name).C(db.Interval).Find(bson.M{"name": mapped.Name}).Count()
	// Duplicate name
	if found > 0 {
		return interval.ID, db.ErrNotUnique
	}

	err = s.DB(mc.database.Name).C(db.Interval).Insert(mapped)
	if err != nil {
		return interval.ID, err
	}

	to := mapped.ToContract()
	return to.ID, err
}


// Update an Interval
// UnexpectedError - failed to update interval in the database
func (mc MongoClient) UpdateInterval(interval contract.Interval) error {
	s := mc.getSessionCopy()
	defer s.Close()

	mapped := &models.Interval{}
	err := mapped.FromContract(interval)

	if err != nil {
		return err
	}
	mapped.Modified = db.MakeTimestamp()

	if mapped.Id.Valid() {
		err := s.DB(mc.database.Name).C(db.Interval).UpdateId(mapped.Id, mapped)
		if err != nil {
			return errorMap(err)
		}
	} else {
		query := bson.M{"uuid": mapped.Uuid}
		err := s.DB(mc.database.Name).C(db.Interval).Update(query, mapped)
		if err != nil {
			return errorMap(err)
		}
	}

	return nil
}

// Remove an Interval by ID
// UnexpectedError - failed to remove interval from the database
func (mc MongoClient) DeleteIntervalById(id string) error {
	return mc.deleteById(db.Interval, id)
}


// ******************************* INTERVAL ACTIONS **********************************

// Return all the Interval Action(s)
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (mc MongoClient) IntervalActions() ([]contract.IntervalAction, error) {

	return mapIntervalActions(mc.getIntervalActions(bson.M{}))
}

// Return Interval Action(s) up to the max number specified
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (mc MongoClient) IntervalActionsWithLimit(limit int) ([]contract.IntervalAction, error){
	return mapIntervalActions(mc.getIntervalActionsLimit(bson.M{},limit))
}

// Return Interval Action(s) by interval name
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (mc MongoClient) IntervalActionsByIntervalName(name string) ([]contract.IntervalAction, error) {

	return mapIntervalActions(mc.getIntervalActions(bson.M{"interval": name}))
}

// Return Interval Action(s) by target name
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (mc MongoClient) IntervalActionsByTarget(name string) ([]contract.IntervalAction, error) {
	return mapIntervalActions(mc.getIntervalActions(bson.M{"target": name}))
}

// Return an Interval Action by ID
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (mc MongoClient) IntervalActionById(id string) (contract.IntervalAction, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// id is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return contract.IntervalAction{}, db.ErrInvalidObjectId
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}

	action := models.IntervalAction{}
	err := s.DB(mc.database.Name).C(db.IntervalAction).Find(query).One(&action)
	if err != nil {
		err = errorMap(err)
		return contract.IntervalAction{}, err
	}

	return action.ToContract(), err
}

// Return an Interval Action by name
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (mc MongoClient) IntervalActionByName(name string) (contract.IntervalAction, error) {

	s := mc.getSessionCopy()
	defer s.Close()

	query := bson.M{"name": name}

	mia := models.IntervalAction{}
	err := s.DB(mc.database.Name).C(db.IntervalAction).Find(query).One(&mia)
	if err != nil {
		err = errorMap(err)
		return contract.IntervalAction{}, err
	}
	return mia.ToContract(), nil
}

// Add a new Interval Action
// UnexpectedError - failed to add interval action into the database
func (mc MongoClient) AddIntervalAction(action contract.IntervalAction) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	mapped := &models.IntervalAction{}
	err := mapped.FromContract(action)
	if err != nil {
		return action.ID, err
	}

	// See if the name is unique and add the value descriptors
	found, err := s.DB(mc.database.Name).C(db.IntervalAction).Find(bson.M{"name": mapped.Name}).Count()
	// Duplicate name
	if found > 0 {
		return action.ID, db.ErrNotUnique
	}

	err = s.DB(mc.database.Name).C(db.IntervalAction).Insert(mapped)
	if err != nil {
		return action.ID, err
	}

	to := mapped.ToContract()
	return to.ID, err
}

// Update an Interval Action
// UnexpectedError - failed to update interval action in the database
func (mc MongoClient) UpdateIntervalAction(action contract.IntervalAction) error {
	s := mc.getSessionCopy()
	defer s.Close()

	mapped := &models.IntervalAction{}
	err := mapped.FromContract(action)
	if err != nil {
		return err
	}
	mapped.Modified = db.MakeTimestamp()

	if mapped.Id.Valid() {
		err := s.DB(mc.database.Name).C(db.IntervalAction).UpdateId(mapped.Id, mapped)
		if err != nil {
			return errorMap(err)
		}
	} else {
		query := bson.M{"uuid": mapped.Uuid}
		err := s.DB(mc.database.Name).C(db.IntervalAction).Update(query, mapped)
		if err != nil {
			return errorMap(err)
		}
	}

	return nil
}


// Remove an Interval Action by ID
// UnexpectedError - failed to remove interval action from the database
func (mc MongoClient) DeleteIntervalActionById(id string) error {
	return mc.deleteById(db.IntervalAction, id)
}

// ******************************* HELPER FUNCTIONS **********************************

// Get Interval Action(s)
func (mc MongoClient) getIntervalActions(q bson.M) ([]models.IntervalAction, error) {

	s := mc.getSessionCopy()
	defer s.Close()

	var mia []models.IntervalAction
	err := s.DB(mc.database.Name).C(db.IntervalAction).Find(q).All(&mia)
	if err != nil {
		return []models.IntervalAction{}, err
	}
	return mia, nil
}

// Get Interval Action(s) with a limit
func (mc MongoClient) getIntervalActionsLimit(q bson.M, limit int)([]models.IntervalAction, error){
	s := mc.getSessionCopy()
	defer s.Close()

	var mia []models.IntervalAction
	// Check if limit is 0
	if limit == 0 {
		return []models.IntervalAction{}, nil
	}

	err := s.DB(mc.database.Name).C(db.IntervalAction).Find(q).Limit(limit).All(&mia)
	if err != nil {
		return []models.IntervalAction{}, err
	}
	return mia, nil
}

// Get all Interval(s)
func (mc MongoClient) getIntervals(q bson.M) ([]models.Interval, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var mi []models.Interval
	err := s.DB(mc.database.Name).C(db.Interval).Find(q).All(&mi)
	if err != nil {
		return []models.Interval{}, err
	}
	return mi, nil
}

// Get Interval(s) with a limit
func (mc MongoClient) getIntervalsLimit(q bson.M, limit int) ([]models.Interval, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var mi []models.Interval
	// Check if limit is 0
	if limit == 0 {
		return []models.Interval{}, nil
	}

	err := s.DB(mc.database.Name).C(db.Interval).Find(q).Limit(limit).All(&mi)
	if err != nil {
		return []models.Interval{}, err
	}
	return mi, nil
}

// Map IntervalActions
func mapIntervalActions(actions []models.IntervalAction, err error) ([]contract.IntervalAction, error) {
	if err != nil {
		return []contract.IntervalAction{}, err
	}
	var mapped []contract.IntervalAction
	for _, action := range actions {
		mapped = append(mapped, action.ToContract())
	}
	return mapped, nil
}

// Map Intervals
func mapIntervals(intervals []models.Interval, err error) ([]contract.Interval, error) {
	if err != nil {
		return []contract.Interval{}, err
	}
	var mapped []contract.Interval
	for _, interval := range intervals {
		mapped = append(mapped, interval.ToContract())
	}
	return mapped, nil
}

// ******************************* UTILITY FUNCTIONS **********************************

// Removes all of the Interval Action(s)
// Returns number of Interval Action(s) removed
// UnexpectedError - failed to remove all of the Interval and IntervalActions from the database
func (m MongoClient) ScrubAllIntervalActions() (int, error) {
	s := m.session.Copy()
	defer s.Close()

	count, err := s.DB(m.database.Name).C(db.IntervalAction).Count()
	if err != nil {
		return 0, err
	}
	_, err = s.DB(m.database.Name).C(db.IntervalAction).RemoveAll(nil)
	if err != nil {
		return 0,err
	}

	return count, nil
}

// Removes all of the Intervals
// Removes any IntervalAction(s) previously not removed as well
// Returns number Interval(s) removed
// UnexpectedError - failed to remove all of the Interval and IntervalActions from the database
func (m MongoClient) ScrubAllIntervals()(int, error){
	s := m.session.Copy()
	defer s.Close()

	// Ensure we have removed interval actions first
	count, err := s.DB(m.database.Name).C(db.IntervalAction).Count()
	if count >0 {
		_, err = s.DB(m.database.Name).C(db.IntervalAction).RemoveAll(nil)
		if err != nil {
			return 0,err
		}
	}
	// count the number interval(s) were removing "overwrite interval actions count"
	count, err = s.DB(m.database.Name).C(db.Interval).Count()
	if err != nil {
		return 0, err
	}
	_, err = s.DB(m.database.Name).C(db.Interval).RemoveAll(nil)
	if err != nil {
		return 0, err
	}

	return count, nil
}

