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

package influx

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	client "github.com/influxdata/influxdb/client/v2"
	"gopkg.in/mgo.v2/bson"
)

// ******************************* EVENTS **********************************

// Return all the events
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (ic *InfluxClient) Events() ([]models.Event, error) {
	return ic.getEvents("")
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (ic *InfluxClient) AddEvent(e *models.Event) (bson.ObjectId, error) {
	e.Created = db.MakeTimestamp()
	e.ID = bson.NewObjectId()

	// Add the event
	err := ic.eventToDB(ic.Database, db.EventsCollection, e, true)
	if err != nil {
		return e.ID, err
	}

	return e.ID, err
}

// Update an event - do NOT update readings
// UnexpectedError - problem updating in database
// NotFound - no event with the ID was found
func (ic *InfluxClient) UpdateEvent(e models.Event) error {
	e.Modified = db.MakeTimestamp()

	// Delete event
	if err := ic.deleteById(db.EventsCollection, e.ID.Hex()); err != nil {
		return err
	}

	// Add the event
	return ic.eventToDB(ic.Database, db.EventsCollection, &e, false)
}

// Get an event by id
func (ic *InfluxClient) EventById(id string) (models.Event, error) {
	if !bson.IsObjectIdHex(id) {
		return models.Event{}, db.ErrInvalidObjectId
	}
	q := fmt.Sprintf("WHERE id = '%s'", id)
	events, err := ic.getEvents(q)
	if len(events) < 1 {
		return models.Event{}, nil
	}
	return events[0], err
}

// Get the number of events in Influx
func (ic *InfluxClient) EventCount() (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", db.EventsCollection)
	return ic.getCount(query)
}

// Get the number of events in Influx for the device
func (ic *InfluxClient) EventCountByDeviceId(id string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE device = '%s'", db.EventsCollection, id)
	return ic.getCount(query)
}

// Delete an event by ID and all of its readings
// 404 - Event not found
// 503 - Unexpected problems
func (ic *InfluxClient) DeleteEventById(id string) error {
	return ic.deleteById(db.EventsCollection, id)
}

// Get a list of events based on the device id and limit
func (ic *InfluxClient) EventsForDeviceLimit(id string, limit int) ([]models.Event, error) {
	query := fmt.Sprintf("WHERE device = '%s' LIMIT %d", id, limit)
	return ic.getEvents(query)
}

// Get a list of events based on the device id
func (ic *InfluxClient) EventsForDevice(id string) ([]models.Event, error) {
	query := fmt.Sprintf("WHERE device = '%s'", id)
	return ic.getEvents(query)
}

// Return a list of events whos creation time is between startTime and endTime
// Limit the number of results by limit
func (ic *InfluxClient) EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error) {
	query := fmt.Sprintf("WHERE created >= %d AND created <= %d LIMIT %d", startTime, endTime, limit)
	return ic.getEvents(query)
}

// Get Events that are older than the given age (defined by age = now - created)
func (ic *InfluxClient) EventsOlderThanAge(age int64) ([]models.Event, error) {
	expireDate := db.MakeTimestamp() - age
	query := fmt.Sprintf("WHERE created < %d", expireDate)
	return ic.getEvents(query)
}

// Get all of the events that have been pushed
func (ic *InfluxClient) EventsPushed() ([]models.Event, error) {
	query := fmt.Sprintf("WHERE pushed > 0")
	return ic.getEvents(query)
}

// Delete all of the readings and all of the events
func (ic *InfluxClient) ScrubAllEvents() error {
	err := ic.deleteAll(db.ReadingsCollection)
	if err != nil {
		return err
	}

	return ic.deleteAll(db.EventsCollection)
}

// Get events for the passed query
func (ic *InfluxClient) getEvents(q string) ([]models.Event, error) {
	events := []models.Event{}
	query := fmt.Sprintf("SELECT * FROM %s %s", db.EventsCollection, q)
	res, err := ic.queryDB(query)
	if err != nil {
		return events, err
	}

	if len(res) == 1 {
		if len(res[0].Series) == 1 {
			events, err = parseEvents(res[0])
			if err != nil {
				return events, err
			}
		}
	}

	return events, nil
}

func (ic *InfluxClient) eventToDB(dbStr string, collection string, e *models.Event, addReadings bool) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  dbStr,
		Precision: "us",
	})
	if err != nil {
		return err
	}

	var stringArray []string
	for i := range e.Readings {
		if addReadings == true {
			e.Readings[i].Id = bson.NewObjectId()
			e.Readings[i].Created = e.Created
			e.Readings[i].Device = e.Device

			fields := map[string]interface{}{
				"pushed":   e.Readings[i].Pushed,
				"created":  e.Readings[i].Created,
				"origin":   e.Readings[i].Origin,
				"modified": e.Readings[i].Modified,
			}

			tags := map[string]string{
				"id":     e.Readings[i].Id.Hex(),
				"device": e.Readings[i].Device,
				"name":   e.Readings[i].Name,
				"value":  e.Readings[i].Value,
			}

			pt, err := client.NewPoint(
				db.ReadingsCollection,
				tags,
				fields,
				time.Now(),
			)
			if err != nil {
				return err
			}
			bp.AddPoint(pt)
		}
		stringArray = append(stringArray, e.Readings[i].Id.Hex())
	}
	readings := strings.Join(stringArray, " ")

	fields := map[string]interface{}{
		"pushed":   e.Pushed,
		"created":  e.Created,
		"origin":   e.Origin,
		"modified": e.Modified,
		"readings": readings,
	}

	tags := map[string]string{
		"id":     e.ID.Hex(),
		"device": e.Device,
		"event":  e.Event,
	}

	pt, err := client.NewPoint(
		collection,
		tags,
		fields,
		time.Now(),
	)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)
	return ic.Client.Write(bp)
}

func parseEvents(res client.Result) ([]models.Event, error) {
	var events []models.Event
	for i := range res.Series[0].Values {
		var event models.Event
		var readings []string
		for j, col := range res.Series[0].Columns {
			switch col {
			case "id":
				event.ID = bson.ObjectIdHex(res.Series[0].Values[i][j].(string))
			case "pushed":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return events, err
				}
				event.Pushed = n
			case "created":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return events, err
				}
				event.Created = n
			case "origin":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return events, err
				}
				event.Origin = n
			case "modified":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return events, err
				}
				event.Modified = n
			case "device":
				if res.Series[0].Values[i][j] != nil {
					event.Device = res.Series[0].Values[i][j].(string)
				}
			case "event":
				if res.Series[0].Values[i][j] != nil {
					event.Event = res.Series[0].Values[i][j].(string)
				}
			case "readings":
				if res.Series[0].Values[i][j] != nil {
					s := res.Series[0].Values[i][j].(string)
					readings = strings.Split(s, " ")
				}
			}
		}
		for _, id := range readings {
			reading, err := currentInfluxClient.ReadingById(id)
			if err != nil {
				return events, nil
			}
			event.Readings = append(event.Readings, reading)
		}

		events = append(events, event)
	}
	return events, nil
}

// ************************ READINGS ************************************

// Return a list of readings sorted by reading id
func (ic *InfluxClient) Readings() ([]models.Reading, error) {
	return ic.getReadings("")
}

// Post a new reading
func (ic *InfluxClient) AddReading(r models.Reading) (bson.ObjectId, error) {
	// Get the reading ready
	r.Id = bson.NewObjectId()
	r.Created = db.MakeTimestamp()
	err := ic.addReadingToDB(ic.Database, db.ReadingsCollection, &r)
	return r.Id, err
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (ic *InfluxClient) UpdateReading(r models.Reading) error {
	r.Modified = db.MakeTimestamp()

	// Delete reading
	if err := ic.deleteById(db.ReadingsCollection, r.Id.Hex()); err != nil {
		return err
	}

	// Add the reading
	return ic.addReadingToDB(ic.Database, db.ReadingsCollection, &r)
}

// Get a reading by ID
func (ic *InfluxClient) ReadingById(id string) (models.Reading, error) {
	// Check if the id is a id hex
	if !bson.IsObjectIdHex(id) {
		return models.Reading{}, db.ErrInvalidObjectId
	}

	query := fmt.Sprintf("WHERE id = '%s'", id)
	readings, err := ic.getReadings(query)
	return readings[0], err
}

// Get the count of readings in Influx
func (ic *InfluxClient) ReadingCount() (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", db.ReadingsCollection)
	return ic.getCount(query)
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (ic *InfluxClient) DeleteReadingById(id string) error {
	return ic.deleteById(db.ReadingsCollection, id)
}

// Return a list of readings for the given device (id or name)
// Sort the list of readings on creation date
func (ic *InfluxClient) ReadingsByDevice(id string, limit int) ([]models.Reading, error) {
	query := fmt.Sprintf("WHERE device = '%s' LIMIT %d", id, limit)
	return ic.getReadings(query)
}

// Return a list of readings for the given value descriptor
// Limit by the given limit
func (ic *InfluxClient) ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error) {
	query := fmt.Sprintf("WHERE name = '%s' LIMIT %d", name, limit)
	return ic.getReadings(query)
}

// Return a list of readings whose name is in the list of value descriptor names
func (ic *InfluxClient) ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error) {
	var readings []models.Reading
	for _, name := range names {
		query := fmt.Sprintf("WHERE name = '%s' LIMIT %d", name, limit)
		rlist, err := ic.getReadings(query)
		if err != nil {
			return readings, err
		}
		readings = append(readings, rlist...)
	}
	return readings, nil
}

// Return a list of readings whos creation time is in-between start and end
// Limit by the limit parameter
func (ic *InfluxClient) ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error) {
	query := fmt.Sprintf("WHERE created >= %d AND created <= %d LIMIT %d", start, end, limit)
	return ic.getReadings(query)
}

// Return a list of readings for a device filtered by the value descriptor and limited by the limit
// The readings are linked to the device through an event
func (ic *InfluxClient) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error) {
	query := fmt.Sprintf("WHERE device = '%s' AND value = '%s' LIMIT %d", deviceId, valueDescriptor, limit)
	return ic.getReadings(query)
}

// Get readings for the passed query
func (ic *InfluxClient) getReadings(q string) ([]models.Reading, error) {
	readings := []models.Reading{}
	query := fmt.Sprintf("SELECT * FROM %s %s", db.ReadingsCollection, q)
	res, err := ic.queryDB(query)
	if err != nil {
		return readings, err
	}

	if len(res) == 1 {
		if len(res[0].Series) == 1 {
			readings, err = parseReadings(res[0])
			if err != nil {
				return readings, err
			}
		}
	}

	return readings, nil
}

func (ic *InfluxClient) addReadingToDB(db string, collection string, r *models.Reading) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db,
		Precision: "us",
	})
	if err != nil {
		return err
	}
	fields := map[string]interface{}{
		"pushed":   r.Pushed,
		"created":  r.Created,
		"origin":   r.Origin,
		"modified": r.Modified,
	}

	tags := map[string]string{
		"id":     r.Id.Hex(),
		"device": r.Device,
		"name":   r.Name,
		"value":  r.Value,
	}

	pt, err := client.NewPoint(
		collection,
		tags,
		fields,
		time.Now(),
	)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)
	return ic.Client.Write(bp)
}

func parseReadings(res client.Result) ([]models.Reading, error) {
	var readings []models.Reading
	for i := range res.Series[0].Values {
		var reading models.Reading
		for j, col := range res.Series[0].Columns {
			switch col {
			case "id":
				reading.Id = bson.ObjectIdHex(res.Series[0].Values[i][j].(string))
			case "pushed":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return readings, err
				}
				reading.Pushed = n
			case "created":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return readings, err
				}
				reading.Created = n
			case "origin":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return readings, err
				}
				reading.Origin = n
			case "modified":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return readings, err
				}
				reading.Modified = n
			case "device":
				if res.Series[0].Values[i][j] != nil {
					reading.Device = res.Series[0].Values[i][j].(string)
				}
			case "name":
				if res.Series[0].Values[i][j] != nil {
					reading.Name = res.Series[0].Values[i][j].(string)
				}
			case "value":
				if res.Series[0].Values[i][j] != nil {
					reading.Value = res.Series[0].Values[i][j].(string)
				}
			}
		}

		readings = append(readings, reading)
	}
	return readings, nil
}

// ************************* VALUE DESCRIPTORS *****************************

// Add a value descriptor
// 409 - Formatting is bad or it is not unique
// 503 - Unexpected
// TODO: Check for valid printf formatting
func (ic *InfluxClient) AddValueDescriptor(v models.ValueDescriptor) (bson.ObjectId, error) {
	// Created/Modified now
	v.Created = db.MakeTimestamp()

	// See if the name is unique and add the value descriptors
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE 'name' = '%s'", db.ValueDescriptorCollection, v.Name)
	num, err := ic.getCount(query)
	if err != nil {
		return v.Id, err
	}

	// Duplicate name
	if num != 0 {
		return v.Id, db.ErrNotUnique
	}

	// Set id
	v.Id = bson.NewObjectId()

	// Add Value Descriptor
	err = ic.addValueDescriptorToDB(ic.Database, db.ValueDescriptorCollection, &v)
	if err != nil {
		return v.Id, err
	}

	return v.Id, err
}

// Return a list of all the value descriptors
// 513 Service Unavailable - database problems
func (ic *InfluxClient) ValueDescriptors() ([]models.ValueDescriptor, error) {
	return ic.getValueDescriptors("")
}

// Update a value descriptor
// First use the ID for identification, then the name
// TODO: Check for the valid printf formatting
// 404 not found if the value descriptor cannot be found by the identifiers
func (ic *InfluxClient) UpdateValueDescriptor(v models.ValueDescriptor) error {
	// See if the name is unique if it changed
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = '%s'", db.ValueDescriptorCollection, v.Id)
	num, err := ic.getCount(query)
	if err != nil {
		return err
	}
	if num != 0 {
		return db.ErrNotUnique
	}
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE name = '%s'", db.ValueDescriptorCollection, v.Name)
	num, err = ic.getCount(query)
	if err != nil {
		return err
	}
	if num != 0 {
		query = fmt.Sprintf("WHERE name = '%s'", v.Name)
		err := ic.deleteValueDescriptorBy(query)
		if err != nil {
			return err
		}
	}
	v.Modified = db.MakeTimestamp()
	// Delete Value Descriptor
	// Add Value Descriptor
	return ic.addValueDescriptorToDB(ic.Database, db.ValueDescriptorCollection, &v)
}

// Delete the value descriptor based on the id
// Not found error if there isn't a value descriptor for the ID
// ValueDescriptorStillInUse if the value descriptor is still referenced by readings
func (ic *InfluxClient) DeleteValueDescriptorById(id string) error {
	query := fmt.Sprintf("WHERE id = '%s'", id)
	return ic.deleteValueDescriptorBy(query)
}

// Return a value descriptor based on the name
// Can return null if no value descriptor is found
func (ic *InfluxClient) ValueDescriptorByName(name string) (models.ValueDescriptor, error) {
	query := fmt.Sprintf("WHERE \"name\" = '%s'", name)
	v, err := ic.getValueDescriptors(query)
	if err != nil || len(v) < 1 {
		var vret models.ValueDescriptor
		return vret, err
	}
	return v[0], err
}

// Return all of the value descriptors based on the names
func (ic *InfluxClient) ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error) {
	vList := []models.ValueDescriptor{}

	for _, name := range names {
		query := fmt.Sprintf("WHERE name = '%s'", name)
		v, err := ic.getValueDescriptors(query)
		if err != nil || len(v) < 1 {
			return []models.ValueDescriptor{}, err
		}
		vList = append(vList, v[0])
	}

	return vList, nil
}

// Return a value descriptor based on the id
// Return NotFoundError if there is no value descriptor for the id
func (ic *InfluxClient) ValueDescriptorById(id string) (models.ValueDescriptor, error) {
	if !bson.IsObjectIdHex(id) {
		return models.ValueDescriptor{}, db.ErrInvalidObjectId
	}

	query := fmt.Sprintf("WHERE id = '%s'", id)
	v, err := ic.getValueDescriptors(query)
	if err != nil || len(v) < 1 {
		return models.ValueDescriptor{}, err
	}
	return v[0], err
}

// Return all the value descriptors that match the UOM label
func (ic *InfluxClient) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	query := fmt.Sprintf("WHERE uomLabel = '%s'", uomLabel)
	return ic.getValueDescriptors(query)
}

// Return value descriptors based on if it has the label
func (ic *InfluxClient) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	query := fmt.Sprintf("WHERE label = '%s'", label)
	return ic.getValueDescriptors(query)
}

// Return value descriptors based on if it has the label
func (ic *InfluxClient) ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error) {
	query := fmt.Sprintf("WHERE type = '%s'", t)
	return ic.getValueDescriptors(query)
}

// Delete all of the value descriptors
func (ic *InfluxClient) ScrubAllValueDescriptors() error {
	return ic.deleteAll(db.ValueDescriptorCollection)
}

func (ic *InfluxClient) deleteValueDescriptorBy(query string) error {
	q := fmt.Sprintf("DELETE  FROM %s %s", db.ValueDescriptorCollection, query)
	_, err := ic.queryDB(q)
	if err != nil {
		return err
	}
	return nil
}

// Get value descriptors for the passed query
func (ic *InfluxClient) getValueDescriptors(q string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	query := fmt.Sprintf("SELECT * FROM %s %s", db.ValueDescriptorCollection, q)
	res, err := ic.queryDB(query)
	if err != nil {
		return vds, err
	}

	if len(res) == 1 {
		if len(res[0].Series) == 1 {
			vds, err = parseValueDescriptors(res[0])
			if err != nil {
				return vds, err
			}
		}
	}

	return vds, nil
}

func (ic *InfluxClient) addValueDescriptorToDB(db string, collection string, v *models.ValueDescriptor) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db,
		Precision: "us",
	})
	if err != nil {
		return err
	}
	fields := map[string]interface{}{
		"description":  v.Description,
		"created":      v.Created,
		"origin":       v.Origin,
		"modified":     v.Modified,
		"min":          v.Min,
		"max":          v.Max,
		"defaultvalue": v.DefaultValue,
		"Labels":       v.Labels,
	}

	tags := map[string]string{
		"id":         v.Id.Hex(),
		"name":       v.Name,
		"UomLabel":   v.UomLabel,
		"type":       v.Type,
		"formatting": v.Formatting,
	}

	pt, err := client.NewPoint(
		collection,
		tags,
		fields,
		time.Now(),
	)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)
	return ic.Client.Write(bp)
}

func parseValueDescriptors(res client.Result) ([]models.ValueDescriptor, error) {
	var vds []models.ValueDescriptor
	for i := range res.Series[0].Values {
		var vd models.ValueDescriptor
		for j, col := range res.Series[0].Columns {
			switch col {
			case "id":
				vd.Id = bson.ObjectIdHex(res.Series[0].Values[i][j].(string))
			case "created":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return vds, err
				}
				vd.Created = n
			case "origin":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return vds, err
				}
				vd.Origin = n
			case "modified":
				n, err := res.Series[0].Values[i][j].(json.Number).Int64()
				if err != nil {
					return vds, err
				}
				vd.Modified = n
			case "name":
				if res.Series[0].Values[i][j] != nil {
					vd.Name = res.Series[0].Values[i][j].(string)
				}
			case "description":
				if res.Series[0].Values[i][j] != nil {
					vd.Description = res.Series[0].Values[i][j].(string)
				}
			case "min":
				if res.Series[0].Values[i][j] != nil {
					vd.Min = res.Series[0].Values[i][j].(string)
				}
			case "max":
				if res.Series[0].Values[i][j] != nil {
					vd.Max = res.Series[0].Values[i][j].(string)
				}
			case "type":
				if res.Series[0].Values[i][j] != nil {
					vd.Type = res.Series[0].Values[i][j].(string)
				}
			case "uomLabel":
				if res.Series[0].Values[i][j] != nil {
					vd.UomLabel = res.Series[0].Values[i][j].(string)
				}
			case "labels":
				// ToDo set labels
				strings := []string{"dummy"}
				vd.Labels = strings
			case "defalutvalue":
				if res.Series[0].Values[i][j] != nil {
					vd.DefaultValue = res.Series[0].Values[i][j].(string)
				}
			}
		}

		vds = append(vds, vd)
	}
	return vds, nil
}
