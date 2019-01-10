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
package models

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
)

type Event struct {
	Id       bson.ObjectId
	Uuid     string
	Pushed   int64
	Device   string
	Created  int64
	Modified int64
	Origin   int64
	Event    string
	Readings []Reading
	dbRefs   []mgo.DBRef
}

func (e *Event) ToContract() contract.Event {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts
	id := e.Uuid
	if id == "" {
		id = e.Id.Hex()
	}
	to := contract.Event{
		ID:       id,
		Pushed:   e.Pushed,
		Device:   e.Device,
		Created:  e.Created,
		Modified: e.Modified,
		Origin:   e.Origin,
		Event:    e.Event,
		Readings: []contract.Reading{},
	}
	for _, r := range e.Readings {
		to.Readings = append(to.Readings, r.ToContract())
	}
	return to
}

func (e *Event) FromContract(from contract.Event) error {
	var err error
	e.Id, e.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return err
	}

	e.Pushed = from.Pushed
	e.Device = from.Device
	e.Created = from.Created
	e.Modified = from.Modified
	e.Origin = from.Origin
	e.Event = from.Event
	e.Readings = []Reading{}
	for _, val := range from.Readings {
		r := &Reading{}
		err := r.FromContract(val)
		if err != nil {
			return errors.New(err.Error() + " id: " + val.Id)
		}
		e.Readings = append(e.Readings, *r)
	}

	if e.Created == 0 {
		e.Created = db.MakeTimestamp()
	}

	return nil
}

// Custom marshaling into mongo
func (e *Event) GetBSON() (interface{}, error) {
	// Turn the readings into DBRef objects
	var readings []mgo.DBRef
	for _, reading := range e.Readings {
		if reading.Id.Valid() {
			readings = append(readings, mgo.DBRef{Collection: db.ReadingsCollection, Id: reading.Id})
		}
	}

	return struct {
		ID       bson.ObjectId `bson:"_id,omitempty"`
		Uuid     string        `bson:"uuid,omitempty"`
		Pushed   int64         `bson:"pushed"`
		Device   string        `bson:"device"` // Device identifier (name or id)
		Created  int64         `bson:"created"`
		Modified int64         `bson:"modified"`
		Origin   int64         `bson:"origin"`
		Schedule string        `bson:"schedule,omitempty"` // Schedule identifier
		Event    string        `bson:"event"`              // Schedule event identifier
		Readings []mgo.DBRef   `bson:"readings,omitempty"` // List of readings
	}{
		ID:       e.Id,
		Uuid:     e.Uuid,
		Pushed:   e.Pushed,
		Device:   e.Device,
		Created:  e.Created,
		Modified: e.Modified,
		Origin:   e.Origin,
		Event:    e.Event,
		Readings: readings,
	}, nil
}

// Custom unmarshaling out of Mongo
func (e *Event) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID       bson.ObjectId `bson:"_id,omitempty"`
		Uuid     string        `bson:"uuid,omitempty"`
		Pushed   int64         `bson:"pushed"`
		Device   string        `bson:"device"` // Device identifier (name or id)
		Created  int64         `bson:"created"`
		Modified int64         `bson:"modified"`
		Origin   int64         `bson:"origin"`
		Schedule string        `bson:"schedule,omitempty"` // Schedule identifier
		Event    string        `bson:"event"`              // Schedule event identifier
		DBRefs   []mgo.DBRef   `bson:"readings"`           // List of readings
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	e.Id = decoded.ID
	e.Uuid = decoded.Uuid
	e.Pushed = decoded.Pushed
	e.Device = decoded.Device
	e.Created = decoded.Created
	e.Modified = decoded.Modified
	e.Origin = decoded.Origin
	e.Event = decoded.Event
	e.Readings = []Reading{}
	e.dbRefs = decoded.DBRefs
	return nil
}

// The purpose of this function is to expose the DBRefs used in our pseudo-FK relationship
// whereby an event is linked to its readings. As I write this, readings are associated to an
// event as a series of DBRefs, each item an Object ID of a reading like so:
// "readings" : [ DBRef("reading", ObjectId("5beb825bdeafc2bc618d4d8b")) ]
// The previous MongoEvent model type included a loop that iterated through the DBRefs and
// called to the database to obtain the reading details. This is poor separation of concerns
// as a model type should not call the database. Further, this was deemed necessary because of
// poor design in the underlying gopkg.in/mgo2 driver.
//
// In order to fully separate the database access and state concerns, I have to provide a getter
// for the internal list of DBRefs. I do not want to pollute the property-based signature of the
// mongo/model/event type.
func (e *Event) GetDBRefs() []mgo.DBRef {
	if e.dbRefs == nil {
		return []mgo.DBRef{}
	}
	return e.dbRefs
}
