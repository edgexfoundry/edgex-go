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
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Uuid     string        `bson:"uuid,omitempty"`
	Pushed   int64         `bson:"pushed"`
	Device   string        `bson:"device"` // Device identifier (name or id)
	Created  int64         `bson:"created"`
	Modified int64         `bson:"modified"`
	Origin   int64         `bson:"origin"`
	Event    string        `bson:"event"`              // Schedule event identifier
	Readings []mgo.DBRef   `bson:"readings,omitempty"` // List of readings

}

func (e *Event) ToContract(transform readingTransform) (c contract.Event, err error) {
	id := e.Uuid
	if id == "" {
		id = e.Id.Hex()
	}

	c.ID = id
	c.Pushed = e.Pushed
	c.Device = e.Device
	c.Created = e.Created
	c.Modified = e.Modified
	c.Origin = e.Origin
	c.Event = e.Event

	c.Readings = []contract.Reading{}
	for _, dbRef := range e.Readings {
		var r Reading
		r, err = transform.DBRefToReading(dbRef)
		if err != nil {
			return contract.Event{}, err
		}
		c.Readings = append(c.Readings, r.ToContract())
	}
	return
}

func (e *Event) FromContract(from contract.Event, transform readingTransform) (err error) {
	e.Id, e.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return
	}

	e.Pushed = from.Pushed
	e.Device = from.Device
	e.Created = from.Created
	e.Modified = from.Modified
	e.Origin = from.Origin
	e.Event = from.Event

	e.Readings = []mgo.DBRef{}
	for _, reading := range from.Readings {
		var readingModel Reading
		err = readingModel.FromContract(reading)
		if err != nil {
			return errors.New(err.Error() + " id: " + reading.Id)
		}

		var dbRef mgo.DBRef
		dbRef, err = transform.ReadingToDBRef(readingModel)
		if err != nil {
			return err
		}
		e.Readings = append(e.Readings, dbRef)
	}

	if e.Created == 0 {
		e.Created = db.MakeTimestamp()
	}

	return
}
