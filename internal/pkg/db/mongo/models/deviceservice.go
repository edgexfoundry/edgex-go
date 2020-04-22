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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type deviceServiceTransform interface {
	DBRefToDeviceService(dbRef mgo.DBRef) (model DeviceService, err error)
	DeviceServiceToDBRef(model DeviceService) (dbRef mgo.DBRef, err error)
}

// DeviceService
//
// Deprecated: Mongo functionality is deprecated as of the Geneva release.
type DeviceService struct {
	Created        int64                   `bson:"created"`
	Modified       int64                   `bson:"modified"`
	Origin         int64                   `bson:"origin"`
	Description    string                  `bson:"description"`
	Id             bson.ObjectId           `bson:"_id,omitempty"`
	Uuid           string                  `bson:"uuid,omitempty"`
	Name           string                  `bson:"name"`           // time in milliseconds that the device last provided any feedback or responded to any request
	LastConnected  int64                   `bson:"lastConnected"`  // time in milliseconds that the device last reported data to the core
	LastReported   int64                   `bson:"lastReported"`   // operational state - either enabled or disabled
	OperatingState contract.OperatingState `bson:"operatingState"` // operational state - ether enabled or disableddc
	Labels         []string                `bson:"labels"`         // tags or other labels applied to the device service for search or other identification needs
	Addressable    mgo.DBRef               `bson:"addressable"`    // address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
	AdminState     contract.AdminState     `bson:"adminState"`     // Device Service Admin State
}

func (ds *DeviceService) ToContract(transform addressableTransform) (c contract.DeviceService, err error) {
	// Always hand back the UUID as the contract command ID unless it's blank (an old command, for example blackbox test scripts)
	id := ds.Uuid
	if id == "" {
		id = ds.Id.Hex()
	}

	c.Created = ds.Created
	c.Modified = ds.Modified
	c.Origin = ds.Origin
	c.Description = ds.Description
	c.Id = id
	c.Name = ds.Name
	c.LastConnected = ds.LastConnected
	c.LastReported = ds.LastReported
	c.OperatingState = ds.OperatingState
	c.Labels = ds.Labels

	aModel, err := transform.DBRefToAddressable(ds.Addressable)
	if err != nil {
		return contract.DeviceService{}, err
	}
	c.Addressable = aModel.ToContract()
	c.AdminState = ds.AdminState
	return
}

func (ds *DeviceService) FromContract(from contract.DeviceService, transform addressableTransform) (id string, err error) {
	ds.AdminState = from.AdminState
	ds.Id, ds.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return
	}

	ds.Created = from.Created
	ds.Modified = from.Modified
	ds.Origin = from.Origin
	ds.Description = from.Description
	ds.Name = from.Name
	ds.LastConnected = from.LastConnected
	ds.LastReported = from.LastReported
	ds.OperatingState = from.OperatingState
	ds.Labels = from.Labels

	if from.Addressable.Id == "" {
		byName, err := transform.GetAddressableByName(from.Addressable.Name)
		if err != nil {
			return "", err
		}

		from.Addressable = byName
	}

	var aModel Addressable
	if _, err = aModel.FromContract(from.Addressable); err != nil {
		return
	}
	if ds.Addressable, err = transform.AddressableToDBRef(aModel); err != nil {
		return
	}

	id = toContractId(ds.Id, ds.Uuid)
	return
}

func (s *DeviceService) TimestampForUpdate() {
	s.Modified = db.MakeTimestamp()
}

func (s *DeviceService) TimestampForAdd() {
	s.TimestampForUpdate()
	s.Created = s.Modified
}
