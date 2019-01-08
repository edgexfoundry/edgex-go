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
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Uuid     string        `bson:"uuid,omitempty"`
	Name     string        `bson:"name"`     // non-database identifier for a device report - must be unique
	Device   string        `bson:"device"`   // associated device name - should be a valid and unique device name
	Event    string        `bson:"event"`    // associated schedule event name - should be a valid and unique schedule event name
	Expected []string      `bson:"expected"` // array of value descriptor names describing the types of data captured in the report
	Created  int64         `bson:"created"`
	Modified int64         `bson:"modified"`
	Origin   int64         `bson:"origin"`
}

func (dr *DeviceReport) ToContract() contract.DeviceReport {
	// Always hand back the UUID as the contract devicereport ID unless it's blank (an old devicereport, for example blackbox test scripts)
	id := dr.Uuid
	if id == "" {
		id = dr.Id.Hex()
	}

	to := contract.DeviceReport{
		Id:       id,
		Name:     dr.Name,
		Device:   dr.Device,
		Event:    dr.Event,
		Expected: dr.Expected,
	}
	to.Created = dr.Created
	to.Modified = dr.Modified
	to.Origin = dr.Origin
	return to
}

func (dr *DeviceReport) FromContract(from contract.DeviceReport) error {
	var err error
	dr.Id, dr.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return err
	}

	dr.Name = from.Name
	dr.Device = from.Device
	dr.Event = from.Event
	dr.Expected = from.Expected
	dr.Created = from.Created
	dr.Modified = from.Modified
	dr.Origin = from.Origin

	if dr.Created == 0 {
		dr.Created = db.MakeTimestamp()
	}

	return nil
}
