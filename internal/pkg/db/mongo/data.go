/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
)

/*
Core data client
Has functions for interacting with the core data mongo database
*/

// ******************************* EVENTS **********************************

// Return all the events
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (mc MongoClient) Events() ([]contract.Event, error) {
	return mc.mapEvents(mc.getEvents(bson.M{}))

}

// Return events up to the max number specified
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (mc MongoClient) EventsWithLimit(limit int) ([]contract.Event, error) {
	return mc.mapEvents(mc.getEventsLimit(bson.M{}, limit))
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (mc MongoClient) AddEvent(e contract.Event) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	//Add the readings
	if len(e.Readings) > 0 {
		var ui []interface{}
		for i, reading := range e.Readings {
			var r models.Reading
			if err := r.FromContract(reading); err != nil {
				return "", err
			}
			ui = append(ui, r)

			e.Readings[i].Id = r.Id.Hex()
		}
		err := s.DB(mc.database.Name).C(db.ReadingsCollection).Insert(ui...)
		if err != nil {
			return "", err
		}
	}

	evt := &models.Event{}
	err := evt.FromContract(e, mc)
	if err != nil {
		return e.ID, err
	}

	// Add the event
	err = s.DB(mc.database.Name).C(db.EventsCollection).Insert(evt)
	mapped, err := evt.ToContract(mc)
	return mapped.ID, err
}

// Update an event - do NOT update readings
// UnexpectedError - problem updating in database
// NotFound - no event with the ID was found
func (mc MongoClient) UpdateEvent(e contract.Event) error {
	s := mc.getSessionCopy()
	defer s.Close()

	mapped := &models.Event{}
	mapped.FromContract(e, mc)
	mapped.Modified = db.MakeTimestamp()

	var err error
	if mapped.Id.Valid() {
		err = s.DB(mc.database.Name).C(db.EventsCollection).UpdateId(mapped.Id, mapped)
	} else {
		query := bson.M{"uuid": mapped.Uuid}
		err = s.DB(mc.database.Name).C(db.EventsCollection).Update(query, mapped)
	}

	return errorMap(err)
}

// Get an event by id
func (mc MongoClient) EventById(id string) (contract.Event, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// EventID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return contract.Event{}, db.ErrInvalidObjectId
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}

	evt := models.Event{}
	err := s.DB(mc.database.Name).C(db.EventsCollection).Find(query).One(&evt)
	if err != nil {
		err = errorMap(err)
		return contract.Event{}, err
	}

	return evt.ToContract(mc)
}

// Get the number of events in Mongo
func (mc MongoClient) EventCount() (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	return s.DB(mc.database.Name).C(db.EventsCollection).Find(nil).Count()
}

// Get the number of events in Mongo for the device
func (mc MongoClient) EventCountByDeviceId(id string) (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	query := bson.M{"device": id}
	return s.DB(mc.database.Name).C(db.EventsCollection).Find(query).Count()
}

// Delete an event by ID and all of its readings
// 404 - Event not found
// 503 - Unexpected problems
func (mc MongoClient) DeleteEventById(id string) error {
	return mc.deleteById(db.EventsCollection, id)
}

// Get a list of events based on the device id and limit
func (mc MongoClient) EventsForDeviceLimit(id string, limit int) ([]contract.Event, error) {
	return mc.mapEvents(mc.getEventsLimit(bson.M{"device": id}, limit))
}

// Get a list of events based on the device id
func (mc MongoClient) EventsForDevice(id string) ([]contract.Event, error) {
	return mc.mapEvents(mc.getEvents(bson.M{"device": id}))
}

// Return a list of events whos creation time is between startTime and endTime
// Limit the number of results by limit
func (mc MongoClient) EventsByCreationTime(startTime, endTime int64, limit int) ([]contract.Event, error) {
	query := bson.M{"created": bson.M{
		"$gte": startTime,
		"$lte": endTime,
	}}
	return mc.mapEvents(mc.getEventsLimit(query, limit))
}

// Get Events that are older than the given age (defined by age = now - created)
func (mc MongoClient) EventsOlderThanAge(age int64) ([]contract.Event, error) {
	expireDate := (db.MakeTimestamp()) - age
	return mc.mapEvents(mc.getEvents(bson.M{"created": bson.M{"$lt": expireDate}}))
}

// Get all of the events that have been pushed
func (mc MongoClient) EventsPushed() ([]contract.Event, error) {
	return mc.mapEvents(mc.getEvents(bson.M{"pushed": bson.M{"$gt": int64(0)}}))
}

// Delete all of the readings and all of the events
func (mc MongoClient) ScrubAllEvents() error {
	s := mc.getSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(db.ReadingsCollection).RemoveAll(nil)
	if err != nil {
		return err
	}

	_, err = s.DB(mc.database.Name).C(db.EventsCollection).RemoveAll(nil)
	if err != nil {
		return err
	}

	return nil
}

// Get events for the passed query
func (mc MongoClient) getEvents(q bson.M) (me []models.Event, err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	err = s.DB(mc.database.Name).C(db.EventsCollection).Find(q).All(&me)
	if err != nil {
		return []models.Event{}, err
	}
	return
}

// Get events with a limit
func (mc MongoClient) getEventsLimit(q bson.M, limit int) (me []models.Event, err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Check if limit is 0
	if limit == 0 {
		return []models.Event{}, nil
	}

	err = s.DB(mc.database.Name).C(db.EventsCollection).Find(q).Limit(limit).All(&me)
	if err != nil {
		return []models.Event{}, err
	}
	return
}

// ************************ READINGS ************************************8

func (mc MongoClient) DBRefToReading(dbRef mgo.DBRef) (a models.Reading, err error) {
	if err = mc.database.C(db.ReadingsCollection).Find(bson.M{"_id": dbRef.Id}).One(&a); err != nil {
		return models.Reading{}, err
	}
	return
}

func (mc MongoClient) ReadingToDBRef(r models.Reading) (dbRef mgo.DBRef, err error) {
	s := mc.session.Copy()
	defer s.Close()

	// validate identity provided in contract actually exists and populate missing Id, Uuid field
	var reading models.Reading
	if r.Id.Valid() {
		reading, err = mc.readingById(r.Id.Hex())
	} else {
		reading, err = mc.readingById(r.Uuid)
	}
	if err != nil {
		return
	}

	dbRef = mgo.DBRef{Collection: db.Addressable, Id: reading.Id}
	return
}

// Return a list of readings sorted by reading id
func (mc MongoClient) Readings() ([]contract.Reading, error) {
	readings, err := mc.getReadings(nil)
	if err != nil {
		return []contract.Reading{}, err
	}

	mapped := []contract.Reading{}
	for _, r := range readings {
		mapped = append(mapped, r.ToContract())
	}
	return mapped, nil
}

// Post a new reading
func (mc MongoClient) AddReading(r contract.Reading) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	mr := models.Reading{}
	mr.FromContract(r)

	err := s.DB(mc.database.Name).C(db.ReadingsCollection).Insert(&mr)

	r = mr.ToContract()
	return r.Id, err
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (mc MongoClient) UpdateReading(r contract.Reading) error {
	s := mc.getSessionCopy()
	defer s.Close()

	mr := models.Reading{}
	mr.FromContract(r)
	mr.Modified = db.MakeTimestamp()

	// Update the reading
	var err error
	if mr.Id.Valid() {
		err = s.DB(mc.database.Name).C(db.ReadingsCollection).UpdateId(mr.Id, mr)
	} else {
		query := bson.M{"uuid": mr.Uuid}
		err = s.DB(mc.database.Name).C(db.ReadingsCollection).Update(query, mr)
	}

	return errorMap(err)
}

// Get a reading by ID
func (mc MongoClient) ReadingById(id string) (contract.Reading, error) {
	res, err := mc.readingById(id)
	if err != nil {
		return contract.Reading{}, err
	}
	return res.ToContract(), nil
}

func (mc MongoClient) readingById(id string) (models.Reading, error) {
	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// ReadingID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return models.Reading{}, db.ErrInvalidObjectId
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}

	s := mc.getSessionCopy()
	defer s.Close()

	var res models.Reading
	err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(query).One(&res)
	return res, errorMap(err)
}

// Get the count of readings in Mongo
func (mc MongoClient) ReadingCount() (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	return s.DB(mc.database.Name).C(db.ReadingsCollection).Find(bson.M{}).Count()
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (mc MongoClient) DeleteReadingById(id string) error {
	return mc.deleteById(db.ReadingsCollection, id)
}

// Return a list of readings for the given device (id or name)
// Sort the list of readings on creation date
func (mc MongoClient) ReadingsByDevice(id string, limit int) ([]contract.Reading, error) {
	query := bson.M{"device": id}
	return mapReadings(mc.getReadingsLimit(query, limit))
}

// Return a list of readings for the given value descriptor
// Limit by the given limit
func (mc MongoClient) ReadingsByValueDescriptor(name string, limit int) ([]contract.Reading, error) {
	query := bson.M{"name": name}
	return mapReadings(mc.getReadingsLimit(query, limit))
}

// Return a list of readings whose name is in the list of value descriptor names
func (mc MongoClient) ReadingsByValueDescriptorNames(names []string, limit int) ([]contract.Reading, error) {
	query := bson.M{"name": bson.M{"$in": names}}
	return mapReadings(mc.getReadingsLimit(query, limit))
}

// Return a list of readings whos creation time is in-between start and end
// Limit by the limit parameter
func (mc MongoClient) ReadingsByCreationTime(start, end int64, limit int) ([]contract.Reading, error) {
	query := bson.M{"created": bson.M{
		"$gte": start,
		"$lte": end,
	}}
	return mapReadings(mc.getReadingsLimit(query, limit))
}

// Return a list of readings for a device filtered by the value descriptor and limited by the limit
// The readings are linked to the device through an event
func (mc MongoClient) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]contract.Reading, error) {
	query := bson.M{"device": deviceId, "name": valueDescriptor}
	return mapReadings(mc.getReadingsLimit(query, limit))
}

func (mc MongoClient) getReadingsLimit(q bson.M, limit int) ([]models.Reading, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	readings := []models.Reading{}

	// Check if limit is 0
	if limit == 0 {
		return readings, nil
	}

	err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(q).Limit(limit).All(&readings)
	return readings, err
}

// Get readings from the database
func (mc MongoClient) getReadings(q bson.M) ([]models.Reading, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	readings := []models.Reading{}
	err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(q).All(&readings)
	return readings, err
}

// Get a reading from the database with the passed query
func (mc MongoClient) getReading(q bson.M) (models.Reading, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var res models.Reading
	err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(q).One(&res)
	err = errorMap(err)
	return res, err
}

// ************************* VALUE DESCRIPTORS *****************************

// Add a value descriptor
// 409 - Formatting is bad or it is not unique
// 503 - Unexpected
// TODO: Check for valid printf formatting
func (mc MongoClient) AddValueDescriptor(v contract.ValueDescriptor) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	mapped := &models.ValueDescriptor{}
	err := mapped.FromContract(v)
	if err != nil {
		return v.Id, err
	}

	// See if the name is unique and add the value descriptors
	found, err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Find(bson.M{"name": mapped.Name}).Count()
	// Duplicate name
	if found > 0 {
		return v.Id, db.ErrNotUnique
	}

	err = s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Insert(mapped)
	if err != nil {
		return v.Id, err
	}

	to := mapped.ToContract()
	return to.Id, err
}

// Return a list of all the value descriptors
// 513 Service Unavailable - database problems
func (mc MongoClient) ValueDescriptors() ([]contract.ValueDescriptor, error) {
	return mapValueDescriptors(mc.getValueDescriptors(nil))
}

// Update a value descriptor
// First use the ID for identification, then the name
// TODO: Check for the valid printf formatting
// 404 not found if the value descriptor cannot be found by the identifiers
func (mc MongoClient) UpdateValueDescriptor(cvd contract.ValueDescriptor) error {
	s := mc.getSessionCopy()
	defer s.Close()

	// See if the name is unique if it changed
	chk, err := mc.getValueDescriptor(bson.M{"name": cvd.Name})
	if err != db.ErrNotFound {
		if err != nil {
			return err
		}

		// IDs are different -> name not unique
		if chk.Id.Hex() != cvd.Id && cvd.Id != chk.Uuid {
			return db.ErrNotUnique
		}
	}

	mvd := models.ValueDescriptor{}
	err = mvd.FromContract(cvd)
	if err != nil {
		return err
	}
	mvd.Modified = db.MakeTimestamp()

	if mvd.Id.Valid() {
		err = s.DB(mc.database.Name).C(db.ValueDescriptorCollection).UpdateId(mvd.Id, mvd)
	} else {
		query := bson.M{"uuid": mvd.Uuid}
		err = s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Update(query, mvd)
	}
	return errorMap(err)
}

// Delete the value descriptor based on the id
// Not found error if there isn't a value descriptor for the ID
// ValueDescriptorStillInUse if the value descriptor is still referenced by readings
func (mc MongoClient) DeleteValueDescriptorById(id string) error {
	return mc.deleteById(db.ValueDescriptorCollection, id)
}

// Return a value descriptor based on the name
// Can return null if no value descriptor is found
func (mc MongoClient) ValueDescriptorByName(name string) (contract.ValueDescriptor, error) {
	query := bson.M{"name": name}
	mvd, err := mc.getValueDescriptor(query)
	if err != nil {
		return contract.ValueDescriptor{}, errorMap(err)
	}
	return mvd.ToContract(), nil
}

// Return all of the value descriptors based on the names
func (mc MongoClient) ValueDescriptorsByName(names []string) ([]contract.ValueDescriptor, error) {
	vList := []contract.ValueDescriptor{}

	for _, name := range names {
		v, err := mc.ValueDescriptorByName(name)
		if err != nil && err != db.ErrNotFound {
			return []contract.ValueDescriptor{}, err
		}
		if err == nil {
			vList = append(vList, v)
		}
	}

	return vList, nil
}

// Return a value descriptor based on the id
// Return NotFoundError if there is no value descriptor for the id
func (mc MongoClient) ValueDescriptorById(id string) (contract.ValueDescriptor, error) {
	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// EventID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return contract.ValueDescriptor{}, db.ErrInvalidObjectId
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}

	mvd, err := mc.getValueDescriptor(query)
	if err != nil {
		return contract.ValueDescriptor{}, err
	}
	return mvd.ToContract(), nil
}

// Return all the value descriptors that match the UOM label
func (mc MongoClient) ValueDescriptorsByUomLabel(uomLabel string) ([]contract.ValueDescriptor, error) {
	query := bson.M{"uomLabel": uomLabel}
	return mapValueDescriptors(mc.getValueDescriptors(query))
}

// Return value descriptors based on if it has the label
func (mc MongoClient) ValueDescriptorsByLabel(label string) ([]contract.ValueDescriptor, error) {
	query := bson.M{"labels": label}
	return mapValueDescriptors(mc.getValueDescriptors(query))
}

// Return value descriptors based on the type
func (mc MongoClient) ValueDescriptorsByType(t string) ([]contract.ValueDescriptor, error) {
	query := bson.M{"type": t}
	return mapValueDescriptors(mc.getValueDescriptors(query))
}

// Delete all of the value descriptors
func (mc MongoClient) ScrubAllValueDescriptors() error {
	s := mc.getSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).RemoveAll(nil)
	if err != nil {
		return err
	}

	return nil
}

// Get value descriptors based on the query
func (mc MongoClient) getValueDescriptors(q bson.M) ([]models.ValueDescriptor, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var v []models.ValueDescriptor
	err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Find(q).All(&v)

	return v, err
}

// Get value descriptors with a limit based on the query
func (mc MongoClient) getValueDescriptorsLimit(q bson.M, limit int) ([]models.ValueDescriptor, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var v []models.ValueDescriptor
	err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Find(q).Limit(limit).All(&v)

	return v, err
}

// Get a value descriptor based on the query
func (mc MongoClient) getValueDescriptor(q bson.M) (models.ValueDescriptor, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var m models.ValueDescriptor
	err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Find(q).One(&m)
	if err != nil {
		err = errorMap(err)
		return models.ValueDescriptor{}, err
	}

	return m, err
}

func (mc MongoClient) mapEvents(events []models.Event, errIn error) (ce []contract.Event, err error) {
	if errIn != nil {
		return []contract.Event{}, errIn
	}

	ce = []contract.Event{}
	for _, event := range events {
		var contractEvent contract.Event
		contractEvent, err = event.ToContract(mc)
		if err != nil {
			return []contract.Event{}, err
		}
		ce = append(ce, contractEvent)
	}

	return
}

func mapReadings(readings []models.Reading, err error) ([]contract.Reading, error) {
	if err != nil {
		return []contract.Reading{}, err
	}

	mapped := []contract.Reading{}
	for _, r := range readings {
		mapped = append(mapped, r.ToContract())
	}
	return mapped, nil
}

func mapValueDescriptors(descriptors []models.ValueDescriptor, err error) ([]contract.ValueDescriptor, error) {
	if err != nil {
		return []contract.ValueDescriptor{}, err
	}

	var mapped []contract.ValueDescriptor
	for _, v := range descriptors {
		mapped = append(mapped, v.ToContract())
	}
	return mapped, nil
}
