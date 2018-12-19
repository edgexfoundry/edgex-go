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

type Addressable struct {
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
	Created    int64         `bson:"created"`
	Modified   int64         `bson:"modified"`
	Origin     int64         `bson:"origin"`
}

func (a Addressable) ToContract() contract.Addressable {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts
	id := a.Uuid
	if id == "" {
		id = a.Id.Hex()
	}
	to := contract.Addressable{
		Id:         id,
		Name:       a.Name,
		Protocol:   a.Protocol,
		HTTPMethod: a.HTTPMethod,
		Address:    a.Address,
		Port:       a.Port,
		Path:       a.Path,
		Publisher:  a.Publisher,
		User:       a.User,
		Password:   a.Password,
		Topic:      a.Topic,
	}
	to.Created = a.Created
	to.Modified = a.Modified
	to.Origin = a.Origin

	return to
}

func (a *Addressable) FromContract(from contract.Addressable) error {
	// In this first case, ID is empty so this must be an add.
	// Generate new BSON/UUIDs
	if from.Id == "" {
		a.Id = bson.NewObjectId()
		a.Uuid = uuid.New().String()
	} else {
		// In this case, we're dealing with an existing event
		if !bson.IsObjectIdHex(from.Id) {
			// EventID is not a BSON ID. Is it a UUID?
			_, err := uuid.Parse(from.Id)
			if err != nil { // It is some unsupported type of string
				return db.ErrInvalidObjectId
			}
			// Leave model's ID blank for now. We will be querying based on the UUID.
			a.Uuid = from.Id
		} else {
			// ID of pre-existing event is a BSON ID. We will query using the BSON ID.
			a.Id = bson.ObjectIdHex(from.Id)
		}
	}
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
	a.Origin = from.Origin

	if a.Created == 0 {
		ts := db.MakeTimestamp()
		a.Created = ts
		a.Modified = ts
	} else {
		a.Modified = from.Modified
	}

	return nil
}
