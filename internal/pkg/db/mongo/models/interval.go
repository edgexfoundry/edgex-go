/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

type Interval struct {
	Created   int64         `bson:"created"`
	Modified  int64         `bson:"modified"`
	Origin    int64         `bson:"origin"`
	Id        bson.ObjectId `bson:"_id,omitempty"`
	Uuid      string        `bson:"uuid,omitempty"`
	Name      string        `bson:"name"`
	Start     string        `bson:"start"`
	End       string        `bson:"end"`
	Frequency string        `bson:"frequency"`
	Cron      string        `bson:"cron,omitempty"`
	RunOnce   bool          `bson:"runonce"`
}

func (in *Interval) ToContract() (c contract.Interval) {
	// Always hand back the UUID as the contract event ID unless it'in blank (an old event, for example blackbox test scripts
	id := in.Uuid
	if id == "" {
		id = in.Id.Hex()
	}

	c.Created = in.Created
	c.Modified = in.Modified
	c.Origin = in.Origin
	c.ID = id
	c.Name = in.Name
	c.Start = in.Start
	c.End = in.End
	c.Frequency = in.Frequency
	c.Cron = in.Cron
	c.RunOnce = in.RunOnce
	return
}

func (in *Interval) FromContract(from contract.Interval) (id string, err error) {
	in.Id, in.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return
	}

	in.Created = from.Created
	in.Modified = from.Modified
	in.Origin = from.Origin
	in.Name = from.Name
	in.Start = from.Start
	in.End = from.End
	in.Frequency = from.Frequency
	in.RunOnce = from.RunOnce
	in.Cron = from.Cron

	id = toContractId(in.Id, in.Uuid)
	return
}

func (in *Interval) TimestampForUpdate() {
	in.Modified = db.MakeTimestamp()
}

func (in *Interval) TimestampForAdd() {
	in.TimestampForUpdate()
	in.Created = in.Modified
}
