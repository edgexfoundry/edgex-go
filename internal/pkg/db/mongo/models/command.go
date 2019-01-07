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
	"github.com/google/uuid"
)

type Command struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Uuid     string        `bson:"uuid,omitempty"`
	Name     string        `bson:"name"`
	Get      *Get          `bson:"get"`
	Put      *Put          `bson:"put"`
	Created  int64         `bson:"created"`
	Modified int64         `bson:"modified"`
	Origin   int64         `bson:"origin"`
}

func (c *Command) ToContract() contract.Command {
	// Always hand back the UUID as the contract command ID unless it's blank (an old command, for example blackbox test scripts)
	id := c.Uuid
	if id == "" {
		id = c.Id.Hex()
	}

	var get *contract.Get
	if c.Get == nil {
		get = nil
	} else {
		get = &[]contract.Get{c.Get.ToContract()}[0]
	}

	var put *contract.Put
	if c.Put == nil {
		put = nil
	} else {
		put = &[]contract.Put{c.Put.ToContract()}[0]
	}

	to := contract.Command{
		Id:   id,
		Name: c.Name,
		Get:  get,
		Put:  put,
	}
	to.Created = c.Created
	to.Modified = c.Modified
	to.Origin = c.Origin
	return to
}

func (c *Command) FromContract(from contract.Command) error {
	// In this first case, ID is empty so this must be an add.
	// Generate new BSON/UUIDs
	if from.Id == "" {
		c.Id = bson.NewObjectId()
		c.Uuid = uuid.New().String()
	} else {
		// In this case, we're dealing with an existing command
		if !bson.IsObjectIdHex(from.Id) {
			// Command Id is not a BSON ID. Is it a UUID?
			_, err := uuid.Parse(from.Id)
			if err != nil { // It is some unsupported type of string
				return db.ErrInvalidObjectId
			}
			// Leave model's ID blank for now. We will be querying based on the UUID.
			c.Uuid = from.Id
		} else {
			// ID of pre-existing event is a BSON ID. We will query using the BSON ID.
			c.Id = bson.ObjectIdHex(from.Id)
		}
	}

	c.Name = from.Name
	c.Get = &Get{}
	err := c.Get.FromContract(*from.Get)
	if err != nil {
		return err
	}
	c.Put = &Put{}
	err = c.Put.FromContract(*from.Put)
	if err != nil {
		return err
	}

	c.Created = from.Created
	c.Modified = from.Modified
	c.Origin = from.Origin

	if c.Created == 0 {
		c.Created = db.MakeTimestamp()
	}

	return nil
}
