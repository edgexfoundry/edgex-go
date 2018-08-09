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
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

/*
Core data client
Has functions for interacting with the core data mongo database
*/

// ******************************* EVENTS **********************************

// Return all the events
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (mc *MongoClient) Events() ([]models.Event, error) {
	return mc.getEvents(bson.M{})
}

// Return events up to the max number specified
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (mc *MongoClient) EventsWithLimit(limit int) ([]models.Event, error) {
	return mc.getEventsLimit(bson.M{}, limit)
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (mc *MongoClient) AddEvent(e *models.Event) (bson.ObjectId, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	e.Created = db.MakeTimestamp()
	e.ID = bson.NewObjectId()

	// Insert readings
	var ui []interface{}
	if len(e.Readings) != 0 {
		for i := range e.Readings {
			e.Readings[i].Id = bson.NewObjectId()
			e.Readings[i].Created = e.Created
			e.Readings[i].Device = e.Device
			ui = append(ui, e.Readings[i])
		}
		err := s.DB(mc.database.Name).C(db.ReadingsCollection).Insert(ui...)
		if err != nil {
			return e.ID, err
		}
	}

	// Handle DBRefs
	me := mongoEvent{Event: *e}

	// Add the event
	err := s.DB(mc.database.Name).C(db.EventsCollection).Insert(me)
	if err != nil {
		return e.ID, err
	}

	return e.ID, err
}

// Update an event - do NOT update readings
// UnexpectedError - problem updating in database
// NotFound - no event with the ID was found
func (mc *MongoClient) UpdateEvent(e models.Event) error {
	s := mc.getSessionCopy()
	defer s.Close()

	e.Modified = db.MakeTimestamp()

	// Handle DBRef
	me := mongoEvent{Event: e}

	err := s.DB(mc.database.Name).C(db.EventsCollection).UpdateId(me.ID, me)
	return errorMap(err)
}

// Get an event by id
func (mc *MongoClient) EventById(id string) (models.Event, error) {
	if !bson.IsObjectIdHex(id) {
		return models.Event{}, db.ErrInvalidObjectId
	}
	return mc.getEvent(bson.M{"_id": bson.ObjectIdHex(id)})
}

// Get the number of events in Mongo
func (mc *MongoClient) EventCount() (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	return s.DB(mc.database.Name).C(db.EventsCollection).Find(nil).Count()
}

// Get the number of events in Mongo for the device
func (mc *MongoClient) EventCountByDeviceId(id string) (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	query := bson.M{"device": id}
	return s.DB(mc.database.Name).C(db.EventsCollection).Find(query).Count()
}

// Delete an event by ID and all of its readings
// 404 - Event not found
// 503 - Unexpected problems
func (mc *MongoClient) DeleteEventById(id string) error {
	return mc.deleteById(db.EventsCollection, id)
}

// Get a list of events based on the device id and limit
func (mc *MongoClient) EventsForDeviceLimit(id string, limit int) ([]models.Event, error) {
	return mc.getEventsLimit(bson.M{"device": id}, limit)
}

// Get a list of events based on the device id
func (mc *MongoClient) EventsForDevice(id string) ([]models.Event, error) {
	return mc.getEvents(bson.M{"device": id})
}

// Return a list of events whos creation time is between startTime and endTime
// Limit the number of results by limit
func (mc *MongoClient) EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error) {
	query := bson.M{"created": bson.M{
		"$gte": startTime,
		"$lte": endTime,
	}}
	return mc.getEventsLimit(query, limit)
}

// Get Events that are older than the given age (defined by age = now - created)
func (mc *MongoClient) EventsOlderThanAge(age int64) ([]models.Event, error) {
	expireDate := (db.MakeTimestamp()) - age
	return mc.getEvents(bson.M{"created": bson.M{"$lt": expireDate}})
}

// Get all of the events that have been pushed
func (mc *MongoClient) EventsPushed() ([]models.Event, error) {
	return mc.getEvents(bson.M{"pushed": bson.M{"$gt": int64(0)}})
}

// Delete all of the readings and all of the events
func (mc *MongoClient) ScrubAllEvents() error {
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
func (mc *MongoClient) getEvents(q bson.M) ([]models.Event, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var me []mongoEvent
	events := []models.Event{}
	err := s.DB(mc.database.Name).C(db.EventsCollection).Find(q).All(&me)
	if err != nil {
		return events, err
	}

	// Append all the events
	for _, e := range me {
		events = append(events, e.Event)
	}

	return events, nil
}

// Get events with a limit
func (mc *MongoClient) getEventsLimit(q bson.M, limit int) ([]models.Event, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var me []mongoEvent
	events := []models.Event{}

	// Check if limit is 0
	if limit == 0 {
		return events, nil
	}

	err := s.DB(mc.database.Name).C(db.EventsCollection).Find(q).Limit(limit).All(&me)
	if err != nil {
		return events, err
	}

	// Append all the events
	for _, e := range me {
		events = append(events, e.Event)
	}

	return events, nil
}

// Get a single event for the passed query
func (mc *MongoClient) getEvent(q bson.M) (models.Event, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRef
	var me mongoEvent
	err := s.DB(mc.database.Name).C(db.EventsCollection).Find(q).One(&me)
	err = errorMap(err)
	return me.Event, err
}

// ************************ READINGS ************************************8

// Return a list of readings sorted by reading id
func (mc *MongoClient) Readings() ([]models.Reading, error) {
	return mc.getReadings(nil)
}

// Post a new reading
func (mc *MongoClient) AddReading(r models.Reading) (bson.ObjectId, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Get the reading ready
	r.Id = bson.NewObjectId()
	r.Created = db.MakeTimestamp()

	err := s.DB(mc.database.Name).C(db.ReadingsCollection).Insert(&r)
	return r.Id, err
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (mc *MongoClient) UpdateReading(r models.Reading) error {
	s := mc.getSessionCopy()
	defer s.Close()

	r.Modified = db.MakeTimestamp()

	// Update the reading
	err := s.DB(mc.database.Name).C(db.ReadingsCollection).UpdateId(r.Id, r)
	return errorMap(err)
}

// Get a reading by ID
func (mc *MongoClient) ReadingById(id string) (models.Reading, error) {
	// Check if the id is a id hex
	if !bson.IsObjectIdHex(id) {
		return models.Reading{}, db.ErrInvalidObjectId
	}

	query := bson.M{"_id": bson.ObjectIdHex(id)}

	return mc.getReading(query)
}

// Get the count of readings in Mongo
func (mc *MongoClient) ReadingCount() (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	return s.DB(mc.database.Name).C(db.ReadingsCollection).Find(bson.M{}).Count()
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (mc *MongoClient) DeleteReadingById(id string) error {
	// Check if the id is a bson id
	if !bson.IsObjectIdHex(id) {
		return db.ErrInvalidObjectId
	}

	return mc.deleteById(db.ReadingsCollection, id)
}

// Return a list of readings for the given device (id or name)
// Sort the list of readings on creation date
func (mc *MongoClient) ReadingsByDevice(id string, limit int) ([]models.Reading, error) {
	query := bson.M{"device": id}
	return mc.getReadingsLimit(query, limit)
}

// Return a list of readings for the given value descriptor
// Limit by the given limit
func (mc *MongoClient) ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error) {
	query := bson.M{"name": name}
	return mc.getReadingsLimit(query, limit)
}

// Return a list of readings whose name is in the list of value descriptor names
func (mc *MongoClient) ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error) {
	query := bson.M{"name": bson.M{"$in": names}}
	return mc.getReadingsLimit(query, limit)
}

// Return a list of readings whos creation time is in-between start and end
// Limit by the limit parameter
func (mc *MongoClient) ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error) {
	query := bson.M{"created": bson.M{
		"$gte": start,
		"$lte": end,
	}}
	return mc.getReadingsLimit(query, limit)
}

// Return a list of readings for a device filtered by the value descriptor and limited by the limit
// The readings are linked to the device through an event
func (mc *MongoClient) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error) {
	query := bson.M{"device": deviceId, "name": valueDescriptor}
	return mc.getReadingsLimit(query, limit)
}

func (mc *MongoClient) getReadingsLimit(q bson.M, limit int) ([]models.Reading, error) {
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
func (mc *MongoClient) getReadings(q bson.M) ([]models.Reading, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	readings := []models.Reading{}
	err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(q).All(&readings)
	return readings, err
}

// Get a reading from the database with the passed query
func (mc *MongoClient) getReading(q bson.M) (models.Reading, error) {
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
func (mc *MongoClient) AddValueDescriptor(v models.ValueDescriptor) (bson.ObjectId, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Created/Modified now
	v.Created = db.MakeTimestamp()

	// See if the name is unique and add the value descriptors
	info, err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Upsert(bson.M{"name": v.Name}, v)
	if err != nil {
		return v.Id, err
	}

	// Duplicate name
	if info.UpsertedId == nil {
		return v.Id, db.ErrNotUnique
	}

	// Set ID
	v.Id = info.UpsertedId.(bson.ObjectId)

	return v.Id, err
}

// Return a list of all the value descriptors
// 513 Service Unavailable - database problems
func (mc *MongoClient) ValueDescriptors() ([]models.ValueDescriptor, error) {
	return mc.getValueDescriptors(nil)
}

// Update a value descriptor
// First use the ID for identification, then the name
// TODO: Check for the valid printf formatting
// 404 not found if the value descriptor cannot be found by the identifiers
func (mc *MongoClient) UpdateValueDescriptor(v models.ValueDescriptor) error {
	s := mc.getSessionCopy()
	defer s.Close()

	// See if the name is unique if it changed
	vd, err := mc.getValueDescriptor(bson.M{"name": v.Name})
	if err != db.ErrNotFound {
		if err != nil {
			return err
		}

		// IDs are different -> name not unique
		if vd.Id != v.Id {
			return db.ErrNotUnique
		}
	}

	v.Modified = db.MakeTimestamp()

	err = s.DB(mc.database.Name).C(db.ValueDescriptorCollection).UpdateId(v.Id, v)
	return errorMap(err)
}

// Delete the value descriptor based on the id
// Not found error if there isn't a value descriptor for the ID
// ValueDescriptorStillInUse if the value descriptor is still referenced by readings
func (mc *MongoClient) DeleteValueDescriptorById(id string) error {
	if !bson.IsObjectIdHex(id) {
		return db.ErrInvalidObjectId
	}
	return mc.deleteById(db.ValueDescriptorCollection, id)
}

// Return a value descriptor based on the name
// Can return null if no value descriptor is found
func (mc *MongoClient) ValueDescriptorByName(name string) (models.ValueDescriptor, error) {
	query := bson.M{"name": name}
	return mc.getValueDescriptor(query)
}

// Return all of the value descriptors based on the names
func (mc *MongoClient) ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error) {
	vList := []models.ValueDescriptor{}

	for _, name := range names {
		v, err := mc.ValueDescriptorByName(name)
		if err != nil && err != db.ErrNotFound {
			return []models.ValueDescriptor{}, err
		}
		if err == nil {
			vList = append(vList, v)
		}
	}

	return vList, nil
}

// Return a value descriptor based on the id
// Return NotFoundError if there is no value descriptor for the id
func (mc *MongoClient) ValueDescriptorById(id string) (models.ValueDescriptor, error) {
	if !bson.IsObjectIdHex(id) {
		return models.ValueDescriptor{}, db.ErrInvalidObjectId
	}

	query := bson.M{"_id": bson.ObjectIdHex(id)}
	return mc.getValueDescriptor(query)
}

// Return all the value descriptors that match the UOM label
func (mc *MongoClient) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	query := bson.M{"uomLabel": uomLabel}
	return mc.getValueDescriptors(query)
}

// Return value descriptors based on if it has the label
func (mc *MongoClient) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	query := bson.M{"labels": label}
	return mc.getValueDescriptors(query)
}

// Return value descriptors based on the type
func (mc *MongoClient) ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error) {
	query := bson.M{"type": t}
	return mc.getValueDescriptors(query)
}

// Delete all of the value descriptors
func (mc *MongoClient) ScrubAllValueDescriptors() error {
	s := mc.getSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).RemoveAll(nil)
	if err != nil {
		return err
	}

	return nil
}

// Get value descriptors based on the query
func (mc *MongoClient) getValueDescriptors(q bson.M) ([]models.ValueDescriptor, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	v := []models.ValueDescriptor{}
	err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Find(q).All(&v)

	return v, err
}

// Get value descriptors with a limit based on the query
func (mc *MongoClient) getValueDescriptorsLimit(q bson.M, limit int) ([]models.ValueDescriptor, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	v := []models.ValueDescriptor{}
	err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Find(q).Limit(limit).All(&v)

	return v, err
}

// Get a value descriptor based on the query
func (mc *MongoClient) getValueDescriptor(q bson.M) (models.ValueDescriptor, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var v models.ValueDescriptor
	err := s.DB(mc.database.Name).C(db.ValueDescriptorCollection).Find(q).One(&v)
	err = errorMap(err)
	return v, err
}
