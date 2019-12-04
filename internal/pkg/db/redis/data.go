/*******************************************************************************
 * Copyright 2018 Redis Labs Inc.
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
package redis

import (
	"fmt"
	"time"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	correlation "github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/imdario/mergo"
)

// deleteReadingsChannel channel used to delete readings asynchronously
var deleteReadingsChannel = make(chan string, 50)
var deleteEventsChannel = make(chan string, 50)

const (
	DeletedEventsCollection   = "gc:" + db.EventsCollection
	DeletedReadingsCollection = "gc:" + db.ReadingsCollection
)

// ******************************* EVENTS **********************************

// ********************** EVENT FUNCTIONS *******************************
// Return all the events
// Sort the events in descending order by ID
// UnexpectedError - failed to retrieve events from the database
func (c *Client) Events() (events []contract.Event, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.EventsCollection, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return events, err
		}
	}
	events = make([]contract.Event, len(objects))
	err = unmarshalEvents(objects, events)
	if err != nil {
		return events, err
	}

	return events, nil
}

// Return events up to the number specified
// UnexpectedError - failed to retrieve events from the database
func (c *Client) EventsWithLimit(limit int) (events []contract.Event, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.EventsCollection, 0, limit-1)
	if err != nil {
		if err != redis.ErrNil {
			return events, err
		}
	}
	events = make([]contract.Event, len(objects))
	err = unmarshalEvents(objects, events)
	if err != nil {
		return events, err
	}

	return events, nil
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (c *Client) AddEvent(e correlation.Event) (id string, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	if e.ID != "" {
		_, err = uuid.Parse(e.ID)
		if err != nil {
			return "", db.ErrInvalidObjectId
		}
	}
	return addEvent(conn, e)
}

// Update an event - do NOT update readings
// UnexpectedError - problem updating in database
// NotFound - no event with the ID was found
func (c *Client) UpdateEvent(e correlation.Event) (err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	event := e.Event

	id := event.ID

	o, err := eventByID(conn, id)
	if err != nil {
		if err == redis.ErrNil {
			return db.ErrNotFound
		}
		return err
	}

	e.Modified = db.MakeTimestamp()
	err = mergo.Merge(&event, o)
	if err != nil {
		return err
	}

	err = deleteEvent(conn, id)
	if err != nil {
		return err
	}

	_, err = addEvent(conn, e)
	return err
}

// Get an event by id
func (c *Client) EventById(id string) (event contract.Event, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	event, err = eventByID(conn, id)
	if err != nil {
		if err == redis.ErrNil {
			return event, db.ErrNotFound
		}
		return event, err
	}

	return event, nil
}

// EventsByChecksum Get an event by checksum
func (c *Client) EventsByChecksum(checksum string) (events []contract.Event, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.EventsCollection+":checksum:"+checksum, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return events, err
		}
	}
	events = make([]contract.Event, len(objects))
	err = unmarshalEvents(objects, events)
	if err != nil {
		return events, err
	}

	if len(events) == 0 {
		return events, db.ErrNotFound
	}

	return events, nil
}

// Get the number of events in Core Data
func (c *Client) EventCount() (count int, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	count, err = redis.Int(conn.Do("ZCARD", db.EventsCollection))
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Get the number of events in Core Data for the device specified by id
func (c *Client) EventCountByDeviceId(id string) (count int, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	count, err = redis.Int(conn.Do("ZCARD", db.EventsCollection+":device:"+id))
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Delete an event by ID. Readings are not deleted as this should be handled by the contract layer
// 404 - Event not found
// 503 - Unexpected problems
func (c *Client) DeleteEventById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteEvent(conn, id)
	if err != nil {
		if err == redis.ErrNil {
			return db.ErrNotFound
		}
		return err
	}

	return nil
}

// DeleteEventsByDevice Delete events and readings associated with the specified deviceID
func (c *Client) DeleteEventsByDevice(deviceId string) (int, error) {
	err := c.DeleteReadingsByDevice(deviceId)
	if err != nil {
		return 0, err
	}

	conn := c.Pool.Get()
	defer conn.Close()

	ids, err := redis.Strings(conn.Do("ZRANGE", db.EventsCollection+":device:"+deviceId, 0, -1))
	if err != nil {
		return 0, err
	}

	err = conn.Send("MULTI")
	if err != nil {
		return 0, err
	}

	for _, id := range ids {
		err = conn.Send("RENAME", id, DeletedEventsCollection+":"+id)
		if err != nil {
			return 0, err
		}
	}

	err = conn.Send("EXEC")
	deleteEventsChannel <- deviceId

	return len(ids), nil
}

// AsyncDeleteEvents Handles the deletion of device events asynchronously. This function is expected to be running in
// a go-routine and works with the "DeleteEventsByDevice" function for better performance.
func (c *CoreDataClient) AsyncDeleteEvents() {
	c.logger.Debug("Starting background event deletion process")
	for {
		select {
		case device, ok := <-deleteEventsChannel:
			if ok {
				c.logger.Debug("Deleting event data for device: " + device)
				startTime := time.Now()
				c.deleteRenamedEvents(device)
				c.logger.Debug(fmt.Sprintf("Deleted events for device: '%s', elapsed time: %s", device, time.Since(startTime)))
			}
		}
	}
}

// deleteRenamedEvents deletes all events associated with the specified device which have been marked for deletion.
// See the "DeleteEventsByDevice" function for details on the how events are marked for deletion(renamed)
func (c *Client) deleteRenamedEvents(device string) {
	conn := c.Pool.Get()
	defer conn.Close()

	ids, err := redis.Strings(conn.Do("ZRANGE", db.EventsCollection+":device:"+device, 0, -1))
	if err != nil {
		c.loggingClient.Error("Unable to delete event:" + err.Error())
		return
	}

	_, err = conn.Do("MULTI")
	if err != nil {
		c.loggingClient.Error("Unable to start transaction for deletion:" + err.Error())
	}

	for _, id := range ids {
		_, err = conn.Do("GET", DeletedEventsCollection+":"+id)
		if err != nil {
			c.loggingClient.Error("Unable to obtain events marked for deletion:" + err.Error())
		}
	}
	events, err := redis.Strings(conn.Do("EXEC"))

	queriesInQueue := 0
	var e correlation.Event
	_, err = conn.Do("MULTI")
	if err != nil {
		c.loggingClient.Error("Unable to start batch processing for event deletion:" + err.Error())
	}

	for _, event := range events {
		err = unmarshalObject([]byte(event), &e)
		if err != nil {
			c.loggingClient.Error("Unable to marshal event: " + err.Error())
		}
		_ = conn.Send("UNLINK", DeletedEventsCollection+":"+e.ID)
		_ = conn.Send("ZREM", db.EventsCollection, e.ID)
		_ = conn.Send("ZREM", db.EventsCollection+":created", e.ID)
		_ = conn.Send("ZREM", db.EventsCollection+":device:"+e.Device, e.ID)
		_ = conn.Send("ZREM", db.EventsCollection+":pushed", e.ID)
		if e.Checksum != "" {
			_ = conn.Send("ZREM", db.EventsCollection+":checksum:"+e.Checksum, 0)
		}

		queriesInQueue++
		if queriesInQueue >= c.BatchSize {
			_, err = conn.Do("EXEC")
			queriesInQueue = 0
			if err != nil {
				c.loggingClient.Error("Unable to execute batch deletion: " + err.Error())
				return
			}
		}
	}

	if queriesInQueue > 0 {
		_, err = conn.Do("EXEC")

		if err != nil {
			c.loggingClient.Error("Unable to execute batch deletion: " + err.Error())
		}
	}
}

// Get a list of events based on the device id and limit
func (c *Client) EventsForDeviceLimit(id string, limit int) (events []contract.Event, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.EventsCollection+":device:"+id, 0, limit-1)
	if err != nil {
		if err != redis.ErrNil {
			return events, err
		}
	}

	events = make([]contract.Event, len(objects))
	err = unmarshalEvents(objects, events)
	if err != nil {
		return events, err
	}

	return events, nil
}

// Get a list of events based on the device id
func (c *Client) EventsForDevice(id string) (events []contract.Event, err error) {
	events, err = c.EventsForDeviceLimit(id, 0)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// Delete all of the events by the device id (and the readings)
// DeleteEventsByDeviceId(id string) error

// Return a list of events whos creation time is between startTime and endTime
// Limit the number of results by limit
func (c *Client) EventsByCreationTime(startTime, endTime int64, limit int) (events []contract.Event, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByScore(conn, db.EventsCollection+":created", startTime, endTime, limit)
	if err != nil {
		if err != redis.ErrNil {
			return events, err
		}
	}

	events = make([]contract.Event, len(objects))
	err = unmarshalEvents(objects, events)
	if err != nil {
		return events, err
	}

	return events, nil
}

// Return a list of readings for a device filtered by the value descriptor and limited by the limit
// The readings are linked to the device through an event
func (c *Client) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) (readings []contract.Reading, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	if limit == 0 {
		return readings, nil
	}

	objects, err := getObjectsByRangeFilter(conn,
		db.ReadingsCollection+":device:"+deviceId,
		db.ReadingsCollection+":name:"+valueDescriptor,
		0, limit-1)
	if err != nil {
		return readings, err
	}

	readings = make([]contract.Reading, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &readings[i])
		if err != nil {
			return readings, err
		}
	}

	return readings, nil

}

// Remove all the events that are older than the given age
// Return the number of events removed
// RemoveEventByAge(age int64) (int, error)

// Get events that are older than a age
func (c *Client) EventsOlderThanAge(age int64) ([]contract.Event, error) {
	expireDate := db.MakeTimestamp() - age

	return c.EventsByCreationTime(0, expireDate, 0)
}

// Remove all the events that have been pushed
// func (dbc *DBClient) ScrubEvents()(int, error)

// Get events that have been pushed (pushed field is not 0)
func (c *Client) EventsPushed() (events []contract.Event, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByScore(conn, db.EventsCollection+":pushed", 1, -1, 0)
	if err != nil {
		return events, err
	}

	events = make([]contract.Event, len(objects))
	err = unmarshalEvents(objects, events)
	if err != nil {
		return events, err
	}

	return events, nil
}

// Delete all readings and events
func (c *Client) ScrubAllEvents() (err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = unlinkCollection(conn, db.EventsCollection)
	if err != nil {
		return err
	}

	err = unlinkCollection(conn, db.ReadingsCollection)
	if err != nil {
		return err
	}

	return nil
}

// ********************* READING FUNCTIONS *************************
// Return a list of readings sorted by reading id
func (c *Client) Readings() (readings []contract.Reading, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ReadingsCollection, 0, -1)
	if err != nil {
		return readings, err
	}

	readings = make([]contract.Reading, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &readings[i])
		if err != nil {
			return readings, err
		}
	}

	return readings, nil
}

// Post a new reading
// Check if valuedescriptor exists in the database
func (c *Client) AddReading(r contract.Reading) (id string, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	if r.Id != "" {
		_, err = uuid.Parse(r.Id)
		if err != nil {
			return "", db.ErrInvalidObjectId
		}
	}
	return addReading(conn, true, r)
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (c *Client) UpdateReading(r contract.Reading) error {
	conn := c.Pool.Get()
	defer conn.Close()

	id := r.Id
	o := contract.Reading{}
	err := getObjectById(conn, id, unmarshalObject, &o)
	if err != nil {
		if err == redis.ErrNil {
			return db.ErrNotFound
		}
		return err
	}

	r.Modified = db.MakeTimestamp()
	err = mergo.Merge(&r, o)
	if err != nil {
		return err
	}

	err = deleteReading(conn, id)
	if err != nil {
		return err
	}

	if r.Id != "" {
		_, err = uuid.Parse(r.Id)
		if err != nil {
			return db.ErrInvalidObjectId
		}
	}
	_, err = addReading(conn, true, r)
	return err
}

// Get a reading by ID
func (c *Client) ReadingById(id string) (reading contract.Reading, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectById(conn, id, unmarshalObject, &reading)
	if err != nil {
		if err == redis.ErrNil {
			return reading, db.ErrNotFound
		}
		return reading, err
	}

	return reading, nil
}

// Get the number of readings in core data
func (c *Client) ReadingCount() (int, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	count, err := redis.Int(conn.Do("ZCARD", db.ReadingsCollection))
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (c *Client) DeleteReadingById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteReading(conn, id)
	if err != nil {
		return err
	}

	return nil
}

// DeleteReadingsByDevice deletes readings associated with the specified device
func (c *Client) DeleteReadingsByDevice(device string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	ids, err := redis.Strings(conn.Do("ZRANGE", db.ReadingsCollection+":device:"+device, 0, -1))
	if err != nil {
		return err
	}

	err = conn.Send("MULTI")
	if err != nil {
		return err
	}

	for _, id := range ids {
		err = conn.Send("RENAME", id, DeletedReadingsCollection+":"+id)
		if err != nil {
			return err
		}
	}

	err = conn.Send("EXEC")
	deleteReadingsChannel <- device

	return nil
}

// AsyncDeleteReadings Handles the deletion of device readings asynchronously. This function is expected to be running
// in a go-routine and works with the "DeleteReadingsByDevice" function for better performance.
func (c *CoreDataClient) AsyncDeleteReadings() {
	c.logger.Debug("Starting background event deletion process")
	for {
		select {
		case device, ok := <-deleteReadingsChannel:
			if ok {
				c.logger.Debug("Deleting reading data for device: " + device)
				startTime := time.Now()
				c.deleteRenamedReadings(device)
				c.logger.Debug(fmt.Sprintf("Deleted readings for device: '%s', elapsed time: %s", device, time.Since(startTime)))
			}
		}
	}
}

// deleteRenamedReadings deletes all readings associated with the specified device which have been marked for deletion.
// See the "DeleteReadingsByDevice" function for details on the how readings are marked for deletion(renamed)
func (c *Client) deleteRenamedReadings(device string) {
	conn := c.Pool.Get()
	defer conn.Close()

	ids, err := redis.Strings(conn.Do("ZRANGE", db.ReadingsCollection+":device:"+device, 0, -1))
	if err != nil {
		c.loggingClient.Error("Unable to delete reading:" + err.Error())
		return
	}

	_, err = conn.Do("MULTI")
	if err != nil {
		c.loggingClient.Error("Unable to start transaction for deletion:" + err.Error())
	}

	for _, id := range ids {
		_, err = conn.Do("GET", DeletedReadingsCollection+":"+id)
		if err != nil {
			c.loggingClient.Error("Unable to obtain readings marked for deletion:" + err.Error())
		}
	}
	readings, err := redis.Strings(conn.Do("EXEC"))

	queriesInQueue := 0
	var r contract.Reading
	_, err = conn.Do("MULTI")
	if err != nil {
		c.loggingClient.Error("Unable to start batch processing for reading deletion:" + err.Error())
	}

	for _, reading := range readings {
		err = unmarshalObject([]byte(reading), &r)
		if err != nil {
			c.loggingClient.Error("Unable to marshal reading: " + err.Error())
		}
		_ = conn.Send("UNLINK", DeletedReadingsCollection+":"+r.Id)
		_ = conn.Send("ZREM", db.ReadingsCollection, r.Id)
		_ = conn.Send("ZREM", db.ReadingsCollection+":created", r.Id)
		_ = conn.Send("ZREM", db.ReadingsCollection+":device:"+r.Device, r.Id)
		_ = conn.Send("ZREM", db.ReadingsCollection+":name:"+r.Name, r.Id)
		queriesInQueue++

		if queriesInQueue >= c.BatchSize {
			_, err = conn.Do("EXEC")
			queriesInQueue = 0
			if err != nil {
				c.loggingClient.Error("Unable to execute batch deletion: " + err.Error())
				return
			}
		}
	}

	if queriesInQueue > 0 {
		_, err = conn.Do("EXEC")

		if err != nil {
			c.loggingClient.Error("Unable to execute batch deletion: " + err.Error())
		}
	}
}

// Return a list of readings for the given device (id or name)
// 404 - meta data checking enabled and can't find the device
// Sort the list of readings on creation date
func (c *Client) ReadingsByDevice(id string, limit int) (readings []contract.Reading, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRevRange(conn, db.ReadingsCollection+":device:"+id, 0, limit-1)
	if err != nil {
		if err != redis.ErrNil {
			return readings, err
		}
	}

	readings = make([]contract.Reading, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &readings[i])
		if err != nil {
			return readings, err
		}
	}

	return readings, nil
}

// Return a list of readings for the given value descriptor
// 413 - the number exceeds the current max limit
func (c *Client) ReadingsByValueDescriptor(name string, limit int) (readings []contract.Reading, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ReadingsCollection+":name:"+name, 0, limit-1)
	if err != nil {
		if err != redis.ErrNil {
			return readings, err
		}
	}

	readings = make([]contract.Reading, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &readings[i])
		if err != nil {
			return readings, err
		}
	}

	return readings, nil
}

// Return a list of readings whose name is in the list of value descriptor names
func (c *Client) ReadingsByValueDescriptorNames(names []string, limit int) (readings []contract.Reading, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	if limit == 0 {
		return readings, nil
	}
	limit--

	for _, name := range names {
		objects, err := getObjectsByRange(conn, db.ReadingsCollection+":name:"+name, 0, limit)
		if err != nil {
			if err != redis.ErrNil {
				return readings, err
			}
		}

		t := make([]contract.Reading, len(objects))
		for i, in := range objects {
			err = unmarshalObject(in, &t[i])
			if err != nil {
				return readings, err
			}
		}

		readings = append(readings, t...)

		limit -= len(objects)
		if limit < 0 {
			break
		}
	}

	return readings, nil
}

// Return a list of readings whos created time is between the start and end times
func (c *Client) ReadingsByCreationTime(start, end int64, limit int) (readings []contract.Reading, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	if limit == 0 {
		return readings, nil
	}

	objects, err := getObjectsByScore(conn, db.ReadingsCollection+":created", start, end, limit)
	if err != nil {
		return readings, err
	}

	readings = make([]contract.Reading, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &readings[i])
		if err != nil {
			return readings, err
		}
	}

	return readings, nil
}

// ************************** VALUE DESCRIPTOR FUNCTIONS ***************************
// Add a value descriptor
// 409 - Formatting is bad or it is not unique
// 503 - Unexpected
// TODO: Check for valid printf formatting
func (c *Client) AddValueDescriptor(v contract.ValueDescriptor) (id string, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	if v.Id != "" {
		_, err = uuid.Parse(v.Id)
		if err != nil {
			return "", db.ErrInvalidObjectId
		}
	}
	return addValue(conn, v)
}

// Return a list of all the value descriptors
// 513 Service Unavailable - database problems
func (c *Client) ValueDescriptors() (values []contract.ValueDescriptor, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ValueDescriptorCollection, 0, -1)
	if err != nil {
		return values, err
	}

	values = make([]contract.ValueDescriptor, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &values[i])
		if err != nil {
			return values, err
		}
	}

	return values, nil
}

// Update a value descriptor
// First use the ID for identification, then the name
// TODO: Check for the valid printf formatting
// 404 not found if the value descriptor cannot be found by the identifiers
func (c *Client) UpdateValueDescriptor(v contract.ValueDescriptor) error {
	conn := c.Pool.Get()
	defer conn.Close()

	id := v.Id
	o, err := valueByName(conn, v.Name)
	if err != nil && err != redis.ErrNil {
		return err
	}
	if err == nil && o.Id != v.Id {
		// IDs are different -> name not unique
		return db.ErrNotUnique
	}

	v.Modified = db.MakeTimestamp()
	err = mergo.Merge(&v, o)
	if err != nil {
		return err
	}

	err = deleteValue(conn, id)
	if err != nil {
		return err
	}

	_, err = addValue(conn, v)
	return err
}

// Delete a value descriptor based on the ID
func (c *Client) DeleteValueDescriptorById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteValue(conn, id)
	if err != nil {
		if err == redis.ErrNil {
			return db.ErrNotFound
		}
		return err
	}
	return nil
}

// Return a value descriptor based on the name
func (c *Client) ValueDescriptorByName(name string) (value contract.ValueDescriptor, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	value, err = valueByName(conn, name)
	if err != nil {
		if err == redis.ErrNil {
			return value, db.ErrNotFound
		}
		return value, err
	}

	return value, nil
}

// Return value descriptors based on the names
func (c *Client) ValueDescriptorsByName(names []string) (values []contract.ValueDescriptor, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	for _, name := range names {
		value, err := valueByName(conn, name)
		if err != nil && err != redis.ErrNil {
			return nil, err
		}

		if err == nil {
			values = append(values, value)
		}
	}

	return values, nil
}

// Delete a valuedescriptor based on the name
// DeleteValueDescriptorByName(name string) error

// Return a value descriptor based on the id
func (c *Client) ValueDescriptorById(id string) (value contract.ValueDescriptor, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectById(conn, id, unmarshalObject, &value)
	if err == redis.ErrNil {
		return value, db.ErrNotFound
	}
	if err != nil {
		return value, err
	}

	return value, nil
}

// Return value descriptors based on the unit of measure label
func (c *Client) ValueDescriptorsByUomLabel(uomLabel string) (values []contract.ValueDescriptor, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ValueDescriptorCollection+":uomlabel:"+uomLabel, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return values, err
		}
	}

	values = make([]contract.ValueDescriptor, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &values[i])
		if err != nil {
			return values, err
		}
	}

	return values, nil
}

// Return value descriptors based on the label
func (c *Client) ValueDescriptorsByLabel(label string) (values []contract.ValueDescriptor, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ValueDescriptorCollection+":label:"+label, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return values, err
		}
	}

	values = make([]contract.ValueDescriptor, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &values[i])
		if err != nil {
			return values, err
		}
	}

	return values, nil
}

// Return a list of value descriptors based on their type
func (c *Client) ValueDescriptorsByType(t string) (values []contract.ValueDescriptor, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ValueDescriptorCollection+":type:"+t, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return values, err
		}
	}

	values = make([]contract.ValueDescriptor, len(objects))
	for i, in := range objects {
		err = unmarshalObject(in, &values[i])
		if err != nil {
			return values, err
		}
	}

	return values, nil
}

// Delete all value descriptors
func (c *Client) ScrubAllValueDescriptors() error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := unlinkCollection(conn, db.ValueDescriptorCollection)
	if err != nil {
		return err
	}

	return nil
}

// ************************** HELPER FUNCTIONS ***************************
func addEvent(conn redis.Conn, e correlation.Event) (id string, err error) {
	if e.Created == 0 {
		e.Created = db.MakeTimestamp()
	}

	if e.ID == "" {
		e.ID = uuid.New().String()
	}

	m, err := marshalEvent(e)
	if err != nil {
		return "", err
	}

	for i, reading := range e.Readings {
		if reading.Device == "" {
			e.Readings[i].Device = e.Device
		}
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", e.ID, m)
	_ = conn.Send("ZADD", db.EventsCollection, 0, e.ID)
	_ = conn.Send("ZADD", db.EventsCollection+":created", e.Created, e.ID)
	_ = conn.Send("ZADD", db.EventsCollection+":pushed", e.Pushed, e.ID)
	_ = conn.Send("ZADD", db.EventsCollection+":device:"+e.Device, e.Created, e.ID)
	if e.Checksum != "" {
		_ = conn.Send("ZADD", db.EventsCollection+":checksum:"+e.Checksum, 0, e.ID)
	}

	rids := make([]interface{}, len(e.Readings)*2+1)
	rids[0] = db.EventsCollection + ":readings:" + e.ID
	for i, r := range e.Readings {
		r.Created = e.Created
		r.Device = e.Device

		if r.Id != "" {
			_, err = uuid.Parse(r.Id)
			if err != nil {
				return "", db.ErrInvalidObjectId
			}
		}
		id, err = addReading(conn, false, r)
		if err != nil {
			return id, err
		}
		rids[i*2+1] = 0
		rids[i*2+2] = id
	}
	if len(rids) > 1 {
		_ = conn.Send("ZADD", rids...)
	}

	_, err = conn.Do("EXEC")
	return e.ID, err
}

func deleteEvent(conn redis.Conn, id string) error {
	o, err := eventByID(conn, id)
	if err != nil {
		if err == redis.ErrNil {
			return db.ErrNotFound
		}
		return err
	}

	// Obtain the checksum associated with the event.
	// This is necessary since the helper functions result in an Event type which does not contain the checksum
	// information, so we need to do a separate lookup to get the information.
	checksum, err := checksumByEventID(conn, o.ID)
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("UNLINK", id)
	_ = conn.Send("UNLINK", db.EventsCollection+":readings:"+id)
	_ = conn.Send("ZREM", db.EventsCollection, id)
	_ = conn.Send("ZREM", db.EventsCollection+":created", id)
	_ = conn.Send("ZREM", db.EventsCollection+":pushed", id)
	_ = conn.Send("ZREM", db.EventsCollection+":device:"+o.Device, id)
	if checksum != "" {
		_ = conn.Send("ZREM", db.EventsCollection+":checksum:"+checksum, id)
	}

	res, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return err
	}
	exists, _ := redis.Bool(res[0], nil)
	if !exists {
		return redis.ErrNil
	}

	// The Contract is responsible for data coherency thus there is no cleanup of specific readings

	return nil
}

func eventByID(conn redis.Conn, id string) (event contract.Event, err error) {
	obj, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return event, db.ErrNotFound
	}
	if err != nil {
		return event, err
	}

	event, err = unmarshalEvent(obj)
	if err != nil {
		return event, err
	}

	return event, err
}

func eventByChecksum(conn redis.Conn, checksum string) (events []contract.Event, err error) {
	objects, err := getObjectsByRange(conn, db.EventsCollection, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return events, err
		}
	}
	events = make([]contract.Event, len(objects))
	err = unmarshalEvents(objects, events)
	if err != nil {
		return events, err
	}

	return events, nil
}

// checksumByEventID retrieves the checksum of the event associated with the provided Event ID.
// If there is no checksum associated with the Event then an empty string is returned.
func checksumByEventID(conn redis.Conn, id string) (string, error) {
	obj, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return "", db.ErrNotFound
	}
	if err != nil {
		return "", err
	}

	event, err := unmarshalRedisEvent(obj)
	if err != nil {
		return "", err
	}

	return event.Checksum, nil
}

// Add a reading to the database
func addReading(conn redis.Conn, tx bool, r contract.Reading) (id string, err error) {
	if r.Created == 0 {
		r.Created = db.MakeTimestamp()
	}

	if r.Id == "" {
		r.Id = uuid.New().String()
	}

	m, err := marshalObject(r)
	if err != nil {
		return r.Id, err
	}

	if tx {
		_ = conn.Send("MULTI")
	}
	_ = conn.Send("SET", r.Id, m)
	_ = conn.Send("ZADD", db.ReadingsCollection, 0, r.Id)
	_ = conn.Send("ZADD", db.ReadingsCollection+":created", r.Created, r.Id)
	_ = conn.Send("ZADD", db.ReadingsCollection+":device:"+r.Device, r.Created, r.Id)
	_ = conn.Send("ZADD", db.ReadingsCollection+":name:"+r.Name, r.Created, r.Id)
	if tx {
		_, err = conn.Do("EXEC")
	}

	return r.Id, err
}

func deleteReading(conn redis.Conn, id string) error {
	r := contract.Reading{}
	err := getObjectById(conn, id, unmarshalObject, &r)
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("UNLINK", id)
	_ = conn.Send("ZREM", db.ReadingsCollection, id)
	_ = conn.Send("ZREM", db.ReadingsCollection+":created", id)
	_ = conn.Send("ZREM", db.ReadingsCollection+":device:"+r.Device, id)
	_ = conn.Send("ZREM", db.ReadingsCollection+":name:"+r.Name, id)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

func addValue(conn redis.Conn, v contract.ValueDescriptor) (id string, err error) {
	if v.Created == 0 {
		v.Created = db.MakeTimestamp()
	}

	if v.Id == "" {
		v.Id = uuid.New().String()
	}

	exists, err := redis.Bool(conn.Do("HEXISTS", db.ValueDescriptorCollection+":name", v.Name))
	if err != nil {
		return "", err
	} else if exists {
		return "", db.ErrNotUnique
	}

	m, err := marshalObject(v)
	if err != nil {
		return "", err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", v.Id, m)
	_ = conn.Send("ZADD", db.ValueDescriptorCollection, 0, v.Id)
	_ = conn.Send("HSET", db.ValueDescriptorCollection+":name", v.Name, v.Id)
	_ = conn.Send("ZADD", db.ValueDescriptorCollection+":uomlabel:"+v.UomLabel, 0, v.Id)
	_ = conn.Send("ZADD", db.ValueDescriptorCollection+":type:"+v.Type, 0, v.Id)
	for _, label := range v.Labels {
		_ = conn.Send("ZADD", db.ValueDescriptorCollection+":label:"+label, 0, v.Id)
	}
	_, err = conn.Do("EXEC")
	return v.Id, err
}

func deleteValue(conn redis.Conn, id string) error {
	v := contract.ValueDescriptor{}
	err := getObjectById(conn, id, unmarshalObject, &v)
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("UNLINK", id)
	_ = conn.Send("ZREM", db.ValueDescriptorCollection, id)
	_ = conn.Send("HDEL", db.ValueDescriptorCollection+":name", v.Name)
	_ = conn.Send("ZREM", db.ValueDescriptorCollection+":uomlabel:"+v.UomLabel, id)
	_ = conn.Send("ZREM", db.ValueDescriptorCollection+":type:"+v.Type, id)
	for _, label := range v.Labels {
		_ = conn.Send("ZREM", db.ValueDescriptorCollection+":label:"+label, 0, id)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}
	return nil
}

func valueByName(conn redis.Conn, name string) (value contract.ValueDescriptor, err error) {
	id, err := redis.String(conn.Do("HGET", db.ValueDescriptorCollection+":name", name))
	if err != nil {
		return value, err
	}

	err = getObjectById(conn, id, unmarshalObject, &value)
	if err != nil {
		return value, err
	}

	return value, nil
}
