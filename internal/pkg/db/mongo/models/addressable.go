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
	DBRefToAddressable(dbRef mgo.DBRef) (a Addressable, err error)
	AddressableToDBRef(a Addressable) (dbRef mgo.DBRef, err error)
}

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

func (a *Addressable) ToContract() contract.Addressable {
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
	var err error
	a.Id, a.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return err
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
