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
)

type readingTransform interface {
	DBRefToReading(dbRef mgo.DBRef) (a Reading, err error)
	ReadingToDBRef(a Reading) (dbRef mgo.DBRef, err error)
}

type Reading struct {
	Created  int64         `bson:"created"`
	Modified int64         `bson:"modified"`
	Origin   int64         `bson:"origin"`
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Uuid     string        `bson:"uuid"`
	Pushed   int64         `bson:"pushed"` // When the data was pushed out of EdgeX (0 - not pushed yet)
	Device   string        `bson:"device"`
	Name     string        `bson:"name"`
	Value    string        `bson:"value"` // Device sensor data value
}

func (r *Reading) ToContract() (c contract.Reading) {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts)
	id := r.Uuid
	if id == "" {
		id = r.Id.Hex()
	}

	c.Id = id
	c.Created = r.Created
	c.Modified = r.Modified
	c.Origin = r.Origin
	c.Pushed = r.Pushed
	c.Device = r.Device
	c.Name = r.Name
	c.Value = r.Value

	return c
}

func (r *Reading) FromContract(from contract.Reading) (id string, err error) {
	r.Id, r.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return
	}

	r.Created = from.Created
	r.Modified = from.Modified
	r.Origin = from.Origin
	r.Pushed = from.Pushed
	r.Device = from.Device
	r.Name = from.Name
	r.Value = from.Value

	id = toContractId(r.Id, r.Uuid)
	return
}

func (r *Reading) TimestampForUpdate() {
	r.Modified = db.MakeTimestamp()
}

func (r *Reading) TimestampForAdd() {
	r.TimestampForUpdate()
	r.Created = r.Modified
}
