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

type Reading struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Uuid     string        `bson:"uuid"`
	Pushed   int64         `bson:"pushed"`  // When the data was pushed out of EdgeX (0 - not pushed yet)
	Created  int64         `bson:"created"` // When the reading was created
	Origin   int64         `bson:"origin"`
	Modified int64         `bson:"modified"`
	Device   string        `bson:"device"`
	Name     string        `bson:"name"`
	Value    string        `bson:"value"` // Device sensor data value
}

func (r *Reading) ToContract() contract.Reading {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts)
	id := r.Uuid
	if id == "" {
		id = r.Id.Hex()
	}
	to := contract.Reading{
		Id:       id,
		Pushed:   r.Pushed,
		Created:  r.Created,
		Origin:   r.Origin,
		Modified: r.Modified,
		Device:   r.Device,
		Name:     r.Name,
		Value:    r.Value,
	}
	return to
}

func (r *Reading) FromContract(from contract.Reading) error {
	var err error
	r.Id, r.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return err
	}

	r.Pushed = from.Pushed
	r.Created = from.Created
	r.Origin = from.Origin
	r.Modified = from.Modified
	r.Device = from.Device
	r.Name = from.Name
	r.Value = from.Value

	if r.Created == 0 {
		r.Created = db.MakeTimestamp()
	}

	return nil
}
