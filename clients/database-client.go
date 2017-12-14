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
 *
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package clients

import (
	"errors"
	"fmt"
	"github.com/edgexfoundry/core-domain-go/models"
	"gopkg.in/mgo.v2/bson"
)

type DatabaseType int8 // Database type enum
const (
	MONGO DatabaseType = iota
)

type DBClient struct {
	dbType      DatabaseType
	mongoClient *MongoClient
}

type DBConfiguration struct {
	DbType       DatabaseType
	Host         string
	Port         int
	Timeout      int
	DatabaseName string
	Username     string
	Password     string
}

var ErrNotFound error = errors.New("Item not found")
var ErrUnsupportedDatabase error = errors.New("Unsuppored database type")
var ErrInvalidObjectId error = errors.New("Invalid object ID")
var ErrNotUnique error = errors.New("Resource already exists")

// Return a pointer to the dbClient
func NewDBClient(config DBConfiguration) (*DBClient, error) {
	var dbClient *DBClient
	switch config.DbType {
	case MONGO:
		// Create the mongo client
		mc, err := newMongoClient(config)
		if err != nil {
			fmt.Println("Error creating the mongo client: " + err.Error())
			return nil, err
		}

		dbClient = &DBClient{dbType: config.DbType, mongoClient: mc}
	default:
		return nil, ErrUnsupportedDatabase
	}

	return dbClient, nil
}

// ********************** EVENT FUNCTIONS *******************************

// Return all the events
// UnexpectedError - failed to retrieve events from the database
func (dbc *DBClient) Events() ([]models.Event, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.Events()
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (dbc *DBClient) AddEvent(e *models.Event) (bson.ObjectId, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.AddEvent(e)
	default:
		return bson.NewObjectId(), ErrUnsupportedDatabase
	}
}

// Update an event - do NOT update readings
// UnexpectedError - problem updating in database
// NotFound - no event with the ID was found
func (dbc *DBClient) UpdateEvent(e models.Event) error {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.UpdateEvent(e)
	default:
		return ErrUnsupportedDatabase
	}
}

// Get an event by id
func (dbc *DBClient) EventById(id string) (models.Event, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.EventById(id)
	default:
		return models.Event{}, ErrUnsupportedDatabase
	}
}

// Get the number of events in Core Data
func (dbc *DBClient) EventCount() (int, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.EventCount()
	default:
		return 0, ErrUnsupportedDatabase
	}
}

// Get the number of events in Core Data for the device specified by id
func (dbc *DBClient) EventCountByDeviceId(id string) (int, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.EventCountByDeviceId(id)
	default:
		return 0, ErrUnsupportedDatabase
	}
}

// Update an event by ID
// Set the pushed variable to the current time
// 404 - Event not found
// 503 - Unexpected problems
//func (dbc *DBClient) UpdateEventById(id string) error{
//	switch dbc.dbType {
//	case MONGO:
//		return dbc.mongoClient.UpdateEventById(id)
//	default:
//		return ErrUnsupportedDatabase
//	}
//}

// Delete an event by ID and all of its readings
// 404 - Event not found
// 503 - Unexpected problems
func (dbc *DBClient) DeleteEventById(id string) error {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.DeleteEventById(id)
	default:
		return ErrUnsupportedDatabase
	}
}

// Get a list of events based on the device id and limit
func (dbc *DBClient) EventsForDeviceLimit(id string, limit int) ([]models.Event, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.EventsForDeviceLimit(id, limit)
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Get a list of events based on the device id
func (dbc *DBClient) EventsForDevice(id string) ([]models.Event, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.EventsForDevice(id)
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Delete all of the events by the device id (and the readings)
func (dbc *DBClient) DeleteEventsByDeviceId(id string) error {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.DeleteEventById(id)
	default:
		return ErrUnsupportedDatabase
	}
}

// Return a list of events whos creation time is between startTime and endTime
// Limit the number of results by limit
func (dbc *DBClient) EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.EventsByCreationTime(startTime, endTime, limit)
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Return a list of readings for a device filtered by the value descriptor and limited by the limit
// The readings are linked to the device through an event
func (dbc *DBClient) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor, limit)
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Remove all the events that are older than the given age
// Return the number of events removed
//func (dbc *DBClient) RemoveEventByAge(age int64) (int, error){
//	switch dbc.dbType {
//	case MONGO:
//		return dbc.mongoClient.RemoveEventByAge(age)
//	default:
//		return 0, ErrUnsupportedDatabase
//	}
//}

// Get events that are older than a age
func (dbc *DBClient) EventsOlderThanAge(age int64) ([]models.Event, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.EventsOlderThanAge(age)
	default:
		return []models.Event{}, ErrUnsupportedDatabase
	}
}

// Remove all the events that have been pushed
//func (dbc *DBClient) ScrubEvents()(int, error){
//	switch dbc.dbType {
//	case MONGO:
//		return dbc.mongoClient.ScrubEvents()
//	default:
//		return 0, ErrUnsupportedDatabase
//	}
//}

// Get events that have been pushed (pushed field is not 0)
func (dbc *DBClient) EventsPushed() ([]models.Event, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.EventsPushed()
	default:
		return []models.Event{}, ErrUnsupportedDatabase
	}
}

func (dbc *DBClient) ScrubAllEvents() error {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ScrubAllEvents()
	default:
		return ErrUnsupportedDatabase
	}
}

// ********************* READING FUNCTIONS *************************

// Return a list of readings sorted by reading id
func (dbc *DBClient) Readings() ([]models.Reading, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.Readings()
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Post a new reading
// Check if valuedescriptor exists in the database
func (dbc *DBClient) AddReading(r models.Reading) (bson.ObjectId, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.AddReading(r)
	default:
		return bson.NewObjectId(), ErrUnsupportedDatabase
	}
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (dbc *DBClient) UpdateReading(r models.Reading) error {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.UpdateReading(r)
	default:
		return ErrUnsupportedDatabase
	}
}

// Get a reading by ID
func (dbc *DBClient) ReadingById(id string) (models.Reading, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ReadingById(id)
	default:
		return models.Reading{}, ErrUnsupportedDatabase
	}
}

// Get the number of readings in core data
func (dbc *DBClient) ReadingCount() (int, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ReadingCount()
	default:
		return 0, ErrUnsupportedDatabase
	}
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (dbc *DBClient) DeleteReadingById(id string) error {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.DeleteReadingById(id)
	default:
		return ErrUnsupportedDatabase
	}
}

// Return a list of readings for the given device (id or name)
// 404 - meta data checking enabled and can't find the device
// Sort the list of readings on creation date
func (dbc *DBClient) ReadingsByDevice(id string, limit int) ([]models.Reading, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ReadingsByDevice(id, limit)
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Return a list of readings for the given value descriptor
// 413 - the number exceeds the current max limit
func (dbc *DBClient) ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ReadingsByValueDescriptor(name, limit)
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Return a list of readings whose name is in the list of value descriptor names
func (dbc *DBClient) ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ReadingsByValueDescriptorNames(names, limit)
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Return a list of readings specified by the UOM label
//func (dbc *DBClient) ReadingsByUomLabel(uomLabel string, limit int)([]models.Reading, error){
//	switch dbc.dbType {
//	case MONGO:
//		return dbc.mongoClient.ReadingsByUomLabel(uomLabel, limit)
//	default:
//		return nil, ErrUnsupportedDatabase
//	}
//}

// Return a list of readings based on the label (value descriptor)
// 413 - limit exceeded
//func (dbc *DBClient) ReadingsByLabel(label string, limit int) ([]models.Reading, error){
//	switch dbc.dbType{
//	case MONGO:
//		return dbc.mongoClient.ReadingsByLabel(label, limit)
//	default:
//		return nil, ErrUnsupportedDatabase
//	}
//}

// Return a list of readings who's value descriptor has the type
//func (dbc *DBClient) ReadingsByType(typeString string, limit int) ([]models.Reading, error){
//	switch dbc.dbType {
//	case MONGO:
//		return dbc.mongoClient.ReadingsByType(typeString, limit)
//	default:
//		return nil, ErrUnsupportedDatabase
//	}
//}

// Return a list of readings whos created time is between the start and end times
func (dbc *DBClient) ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ReadingsByCreationTime(start, end, limit)
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// ************************** VALUE DESCRIPTOR FUNCTIONS ***************************

// Add a value descriptor
// 409 - Formatting is bad or it is not unique
// 503 - Unexpected
// TODO: Check for valid printf formatting
func (dbc *DBClient) AddValueDescriptor(v models.ValueDescriptor) (bson.ObjectId, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.AddValueDescriptor(v)
	default:
		return bson.NewObjectId(), ErrUnsupportedDatabase
	}
}

// Return a list of all the value descriptors
// 513 Service Unavailable - database problems
func (dbc *DBClient) ValueDescriptors() ([]models.ValueDescriptor, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ValueDescriptors()
	default:
		return nil, ErrUnsupportedDatabase
	}
}

// Update a value descriptor
// First use the ID for identification, then the name
// TODO: Check for the valid printf formatting
// 404 not found if the value descriptor cannot be found by the identifiers
func (dbc *DBClient) UpdateValueDescriptor(v models.ValueDescriptor) error {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.UpdateValueDescriptor(v)
	default:
		return ErrUnsupportedDatabase
	}
}

// Delete a value descriptor based on the ID
func (dbc *DBClient) DeleteValueDescriptorById(id string) error {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.DeleteValueDescriptorById(id)
	default:
		return ErrUnsupportedDatabase
	}
}

// Return a value descriptor based on the name
func (dbc *DBClient) ValueDescriptorByName(name string) (models.ValueDescriptor, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ValueDescriptorByName(name)
	default:
		return models.ValueDescriptor{}, ErrUnsupportedDatabase
	}
}

// Return value descriptors based on the names
func (dbc *DBClient) ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ValueDescriptorsByName(names)
	default:
		return []models.ValueDescriptor{}, ErrUnsupportedDatabase
	}
}

// Delete a valuedescriptor based on the name
//func (dbc *DBClient) DeleteValueDescriptorByName(name string) error{
//	switch dbc.dbType {
//	case MONGO:
//		return dbc.mongoClient.DeleteValueDescriptorByName(name)
//	default:
//		return ErrUnsupportedDatabase
//	}
//}

// Return a value descriptor based on the id
func (dbc *DBClient) ValueDescriptorById(id string) (models.ValueDescriptor, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ValueDescriptorById(id)
	default:
		return models.ValueDescriptor{}, ErrUnsupportedDatabase
	}
}

// Return value descriptors based on the unit of measure label
func (dbc *DBClient) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ValueDescriptorsByUomLabel(uomLabel)
	default:
		return []models.ValueDescriptor{}, ErrUnsupportedDatabase
	}
}

// Return value descriptors based on the label
func (dbc *DBClient) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ValueDescriptorsByLabel(label)
	default:
		return []models.ValueDescriptor{}, ErrUnsupportedDatabase
	}
}

// Return a list of value descriptors based on their type
func (dbc *DBClient) ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error) {
	switch dbc.dbType {
	case MONGO:
		return dbc.mongoClient.ValueDescriptorsByType(t)
	default:
		return []models.ValueDescriptor{}, ErrUnsupportedDatabase
	}
}
