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

type addressableTransform interface {
	DBRefToAddressable(dbRef mgo.DBRef) (model Addressable, err error)
	AddressableToDBRef(model Addressable) (dbRef mgo.DBRef, err error)
}

type Addressable struct {
	Created    int64         `bson:"created"`
	Modified   int64         `bson:"modified"`
	Origin     int64         `bson:"origin"`
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Uuid       string        `bson:"uuid,omitempty"`
	Name       string        `bson:"name"`
	Protocol   string        `bson:"protocol"`  // Protocol for the address (HTTP/TCP)
	HTTPMethod string        `bson:"method"`    // Method for connecting (i.e. POST)
	Address    string        `bson:"address"`   // Address of the addressable
	Port       int           `bson:"port"`      // Port for the address
	Path       string        `bson:"path"`      // Path for callbacks
	Publisher  string        `bson:"publisher"` // For message bus protocols
	User       string        `bson:"user"`      // User id for authentication
	Password   string        `bson:"password"`  // Password of the user for authentication for the addressable
	Topic      string        `bson:"topic"`     // Topic for message bus addressables
}

func (a *Addressable) ToContract() (c contract.Addressable) {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts
	id := a.Uuid
	if id == "" {
		id = a.Id.Hex()
	}

	c.Created = a.Created
	c.Modified = a.Modified
	c.Origin = a.Origin
	c.Id = id
	c.Name = a.Name
	c.Protocol = a.Protocol
	c.HTTPMethod = a.HTTPMethod
	c.Address = a.Address
	c.Port = a.Port
	c.Path = a.Path
	c.Publisher = a.Publisher
	c.User = a.User
	c.Password = a.Password
	c.Topic = a.Topic

	return
}

func (a *Addressable) FromContract(from contract.Addressable) (id string, err error) {
	if a.Id, a.Uuid, err = fromContractId(from.Id); err != nil {
		return
	}

	a.Created = from.Created
	a.Modified = from.Modified
	a.Origin = from.Origin
	a.Name = from.Name
	a.Protocol = from.Protocol
	a.HTTPMethod = from.HTTPMethod
	a.Address = from.Address
	a.Port = from.Port
	a.Path = from.Path
	a.Publisher = from.Publisher
	a.User = from.User
	a.Password = from.Password
	a.Topic = from.Topic

	id = toContractId(a.Id, a.Uuid)
	return
}

func (a *Addressable) TimestampForUpdate() {
	a.Modified = db.MakeTimestamp()
}

func (a *Addressable) TimestampForAdd() {
	a.TimestampForUpdate()
	a.Created = a.Modified
}
