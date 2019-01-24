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
	"github.com/globalsign/mgo/bson"
)

type DeviceReport struct {
	Created  int64         `bson:"created"`
	Modified int64         `bson:"modified"`
	Origin   int64         `bson:"origin"`
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Uuid     string        `bson:"uuid,omitempty"`
	Name     string        `bson:"name"`     // non-database identifier for a device report - must be unique
	Device   string        `bson:"device"`   // associated device name - should be a valid and unique device name
	Event    string        `bson:"event"`    // associated schedule event name - should be a valid and unique schedule event name
	Expected []string      `bson:"expected"` // array of value descriptor names describing the types of data captured in the report
}

func (dr *DeviceReport) ToContract() (c contract.DeviceReport) {
	// Always hand back the UUID as the contract devicereport ID unless it's blank (an old devicereport, for example blackbox test scripts)
	id := dr.Uuid
	if id == "" {
		id = dr.Id.Hex()
	}

	c.Created = dr.Created
	c.Modified = dr.Modified
	c.Origin = dr.Origin
	c.Id = id
	c.Name = dr.Name
	c.Device = dr.Device
	c.Event = dr.Event
	c.Expected = dr.Expected

	return
}

func (dr *DeviceReport) FromContract(from contract.DeviceReport) (id string, err error) {
	if dr.Id, dr.Uuid, err = fromContractId(from.Id); err != nil {
		return
	}

	dr.Created = from.Created
	dr.Modified = from.Modified
	dr.Origin = from.Origin
	dr.Name = from.Name
	dr.Device = from.Device
	dr.Event = from.Event
	dr.Expected = from.Expected

	id = toContractId(dr.Id, dr.Uuid)
	return
}

func (dr *DeviceReport) TimestampForUpdate() {
	dr.Modified = db.MakeTimestamp()
}

func (dr *DeviceReport) TimestampForAdd() {
	dr.TimestampForUpdate()
	dr.Created = dr.Modified
}
