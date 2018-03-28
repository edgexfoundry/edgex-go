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

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2/bson"
)

type DatabaseType int8 // Database type enum
const (
	MONGO DatabaseType = iota
	MOCK
	INFLUX
)

type DBClient interface {
	// ********************** EVENT FUNCTIONS *******************************
	// Return all the events
	// UnexpectedError - failed to retrieve events from the database
	Events() ([]models.Event, error)

	// Add a new event
	// UnexpectedError - failed to add to database
	// NoValueDescriptor - no existing value descriptor for a reading in the event
	AddEvent(e *models.Event) (bson.ObjectId, error)

	// Update an event - do NOT update readings
	// UnexpectedError - problem updating in database
	// NotFound - no event with the ID was found
	UpdateEvent(e models.Event) error

	// Get an event by id
	EventById(id string) (models.Event, error)

	// Get the number of events in Core Data
	EventCount() (int, error)

	// Get the number of events in Core Data for the device specified by id
	EventCountByDeviceId(id string) (int, error)

	// Update an event by ID
	// Set the pushed variable to the current time
	// 404 - Event not found
	// 503 - Unexpected problems
	//UpdateEventById(id string) error

	// Delete an event by ID and all of its readings
	// 404 - Event not found
	// 503 - Unexpected problems
	DeleteEventById(id string) error

	// Get a list of events based on the device id and limit
	EventsForDeviceLimit(id string, limit int) ([]models.Event, error)

	// Get a list of events based on the device id
	EventsForDevice(id string) ([]models.Event, error)

	// Delete all of the events by the device id (and the readings)
	//DeleteEventsByDeviceId(id string) error

	// Return a list of events whos creation time is between startTime and endTime
	// Limit the number of results by limit
	EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error)

	// Return a list of readings for a device filtered by the value descriptor and limited by the limit
	// The readings are linked to the device through an event
	ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error)

	// Remove all the events that are older than the given age
	// Return the number of events removed
	//RemoveEventByAge(age int64) (int, error)

	// Get events that are older than a age
	EventsOlderThanAge(age int64) ([]models.Event, error)

	// Remove all the events that have been pushed
	//func (dbc *DBClient) ScrubEvents()(int, error)

	// Get events that have been pushed (pushed field is not 0)
	EventsPushed() ([]models.Event, error)

	ScrubAllEvents() error

	// ********************* READING FUNCTIONS *************************
	// Return a list of readings sorted by reading id
	Readings() ([]models.Reading, error)

	// Post a new reading
	// Check if valuedescriptor exists in the database
	AddReading(r models.Reading) (bson.ObjectId, error)

	// Update a reading
	// 404 - reading cannot be found
	// 409 - Value descriptor doesn't exist
	// 503 - unknown issues
	UpdateReading(r models.Reading) error

	// Get a reading by ID
	ReadingById(id string) (models.Reading, error)

	// Get the number of readings in core data
	ReadingCount() (int, error)

	// Delete a reading by ID
	// 404 - can't find the reading with the given id
	DeleteReadingById(id string) error

	// Return a list of readings for the given device (id or name)
	// 404 - meta data checking enabled and can't find the device
	// Sort the list of readings on creation date
	ReadingsByDevice(id string, limit int) ([]models.Reading, error)

	// Return a list of readings for the given value descriptor
	// 413 - the number exceeds the current max limit
	ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error)

	// Return a list of readings whose name is in the list of value descriptor names
	ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error)

	// Return a list of readings specified by the UOM label
	//ReadingsByUomLabel(uomLabel string, limit int)([]models.Reading, error)

	// Return a list of readings based on the label (value descriptor)
	// 413 - limit exceeded
	//ReadingsByLabel(label string, limit int) ([]models.Reading, error)

	// Return a list of readings who's value descriptor has the type
	//ReadingsByType(typeString string, limit int) ([]models.Reading, error)

	// Return a list of readings whos created time is between the start and end times
	ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error)

	// ************************** VALUE DESCRIPTOR FUNCTIONS ***************************
	// Add a value descriptor
	// 409 - Formatting is bad or it is not unique
	// 503 - Unexpected
	// TODO: Check for valid printf formatting
	AddValueDescriptor(v models.ValueDescriptor) (bson.ObjectId, error)

	// Return a list of all the value descriptors
	// 513 Service Unavailable - database problems
	ValueDescriptors() ([]models.ValueDescriptor, error)

	// Update a value descriptor
	// First use the ID for identification, then the name
	// TODO: Check for the valid printf formatting
	// 404 not found if the value descriptor cannot be found by the identifiers
	UpdateValueDescriptor(v models.ValueDescriptor) error

	// Delete a value descriptor based on the ID
	DeleteValueDescriptorById(id string) error

	// Return a value descriptor based on the name
	ValueDescriptorByName(name string) (models.ValueDescriptor, error)

	// Return value descriptors based on the names
	ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error)

	// Delete a valuedescriptor based on the name
	//DeleteValueDescriptorByName(name string) error

	// Return a value descriptor based on the id
	ValueDescriptorById(id string) (models.ValueDescriptor, error)

	// Return value descriptors based on the unit of measure label
	ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error)

	// Return value descriptors based on the label
	ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error)

	// Return a list of value descriptors based on their type
	ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error)
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

// Return the dbClient interface
func NewDBClient(config DBConfiguration) (DBClient, error) {
	switch config.DbType {
	case MONGO:
		// Create the mongo client
		mc, err := newMongoClient(config)
		if err != nil {
			fmt.Println("Error creating the mongo client: " + err.Error())
			return nil, err
		}
		return mc, nil
	case INFLUX:
		// Create the influx client
		ic, err := newInfluxClient(config)
		if err != nil {
			fmt.Println("Error creating the influx client: " + err.Error())
			return nil, err
		}
		return ic, nil
	case MOCK:
		//Create the mock client
		mock := &MockDb{}
		return mock, nil
	default:
		return nil, ErrUnsupportedDatabase
	}
}
