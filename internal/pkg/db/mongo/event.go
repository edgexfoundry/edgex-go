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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Struct that wraps an event to handle DBRefs
type mongoEvent struct {
	models.Event
}

// Custom marshaling into mongo
func (me mongoEvent) GetBSON() (interface{}, error) {
	// Turn the readings into DBRef objects
	var readings []mgo.DBRef
	for _, reading := range me.Readings {
		readings = append(readings, mgo.DBRef{Collection: db.ReadingsCollection, Id: reading.Id})
	}

	return struct {
		ID       bson.ObjectId `bson:"_id,omitempty"`
		Pushed   int64         `bson:"pushed"`
		Device   string        `bson:"device"` // Device identifier (name or id)
		Created  int64         `bson:"created"`
		Modified int64         `bson:"modified"`
		Origin   int64         `bson:"origin"`
		Schedule string        `bson:"schedule,omitempty"` // Schedule identifier
		Event    string        `bson:"event"`              // Schedule event identifier
		Readings []mgo.DBRef   `bson:"readings"`           // List of readings
	}{
		ID:       me.ID,
		Pushed:   me.Pushed,
		Device:   me.Device,
		Created:  me.Created,
		Modified: me.Modified,
		Origin:   me.Origin,
		Event:    me.Event.Event,
		Readings: readings,
	}, nil
}

// Custom unmarshaling out of mongo
func (me *mongoEvent) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID       bson.ObjectId `bson:"_id,omitempty"`
		Pushed   int64         `bson:"pushed"`
		Device   string        `bson:"device"` // Device identifier (name or id)
		Created  int64         `bson:"created"`
		Modified int64         `bson:"modified"`
		Origin   int64         `bson:"origin"`
		Schedule string        `bson:"schedule,omitempty"` // Schedule identifier
		Event    string        `bson:"event"`              // Schedule event identifier
		Readings []mgo.DBRef   `bson:"readings"`           // List of readings
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	me.ID = decoded.ID
	me.Pushed = decoded.Pushed
	me.Device = decoded.Device
	me.Created = decoded.Created
	me.Modified = decoded.Modified
	me.Origin = decoded.Origin
	me.Event.Event = decoded.Event

	// De-reference the DBRef fields
	mc, err := getCurrentMongoClient()
	if err != nil {
		return err
	}

	var readings []models.Reading

	// Get all of the reading objects
	for _, rRef := range decoded.Readings {
		var reading models.Reading
		err := mc.database.C(db.ReadingsCollection).FindId(rRef.Id).One(&reading)
		if err != nil {
			return err
		}

		readings = append(readings, reading)
	}

	me.Readings = readings

	return nil
}
