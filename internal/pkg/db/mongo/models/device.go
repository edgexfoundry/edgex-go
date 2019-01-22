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
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

/*
 * This file is the model for the Device object in EdgeX
 *
 *
 * Device struct
 */
type Device struct {
	Created        int64                   `bson:"created"`
	Modified       int64                   `bson:"modified"`
	Origin         int64                   `bson:"origin"`
	Description    string                  `bson:"description"`
	Id             bson.ObjectId           `bson:"_id,omitempty"`
	Uuid           string                  `bson:"uuid,omitempty"`
	Name           string                  `bson:"name"`           // Unique name for identifying a device
	AdminState     contract.AdminState     `bson:"adminState"`     // Admin state (locked/unlocked)
	OperatingState contract.OperatingState `bson:"operatingState"` // Operating state (enabled/disabled)
	Addressable    mgo.DBRef               `bson:"addressable"`    // Addressable for the device - stores information about it's address
	LastConnected  int64                   `bson:"lastConnected"`  // Time (milliseconds) that the device last provided any feedback or responded to any request
	LastReported   int64                   `bson:"lastReported"`   // Time (milliseconds) that the device reported data to the core microservice
	Labels         []string                `bson:"labels"`         // Other labels applied to the device to help with searching
	Location       interface{}             `bson:"location"`       // Device service specific location (interface{} is an empty interface so it can be anything)
	Service        mgo.DBRef               `bson:"service"`        // Associated Device Service - One per device
	Profile        mgo.DBRef               `bson:"profile"`        // Associated Device Profile - Describes the device
}

func (d *Device) ToContract(dsTransform deviceServiceTransform, dpTransform deviceProfileTransform, cTransform commandTransform, aTransform addressableTransform) (c contract.Device, err error) {
	// Always hand back the UUID as the contract command ID unless it's blank (an old command, for example blackbox test scripts)
	id := d.Uuid
	if id == "" {
		id = d.Id.Hex()
	}

	var result contract.Device

	c.Created = d.Created
	c.Modified = d.Modified
	c.Origin = d.Origin
	c.Description = d.Description
	result.Id = id
	result.Name = d.Name
	result.AdminState = d.AdminState
	result.OperatingState = d.OperatingState

	aModel, err := aTransform.DBRefToAddressable(d.Addressable)
	if err != nil {
		return
	}
	result.Addressable = aModel.ToContract()

	result.LastConnected = d.LastConnected
	result.LastReported = d.LastReported
	result.Labels = d.Labels
	result.Location = d.Location

	dsModel, err := dsTransform.DBRefToDeviceService(d.Service)
	if err != nil {
		return
	}
	result.Service, err = dsModel.ToContract(aTransform)
	if err != nil {
		return
	}

	dpModel, err := dpTransform.DBRefToDeviceProfile(d.Profile)
	if err != nil {
		return
	}
	result.Profile, err = dpModel.ToContract(cTransform)
	if err != nil {
		return
	}

	c = result
	return
}

func (d *Device) FromContract(from contract.Device, dsTransform deviceServiceTransform, dpTransform deviceProfileTransform, cTransform commandTransform, aTransform addressableTransform) (id string, err error) {
	if d.Id, d.Uuid, err = fromContractId(from.Id); err != nil {
		return
	}

	d.Created = from.Created
	d.Modified = from.Modified
	d.Origin = from.Origin
	d.Description = from.Description
	d.Name = from.Name
	d.AdminState = from.AdminState
	d.OperatingState = from.OperatingState

	var aModel Addressable
	if _, err = aModel.FromContract(from.Addressable); err != nil {
		return
	}
	if d.Addressable, err = aTransform.AddressableToDBRef(aModel); err != nil {
		return
	}

	d.LastConnected = from.LastConnected
	d.LastReported = from.LastReported
	d.Labels = from.Labels
	d.Location = from.Location

	var dsModel DeviceService
	if _, err = dsModel.FromContract(from.Service, aTransform); err != nil {
		return
	}
	if d.Service, err = dsTransform.DeviceServiceToDBRef(dsModel); err != nil {
		return
	}

	var dpModel DeviceProfile
	if _, err = dpModel.FromContract(from.Profile, cTransform); err != nil {
		return
	}
	if d.Profile, err = dpTransform.DeviceProfileToDBRef(dpModel); err != nil {
		return
	}
	id = toContractId(d.Id, d.Uuid)
	return
}

func (d *Device) TimestampForUpdate() {
	d.Modified = db.MakeTimestamp()
}

func (d *Device) TimestampForAdd() {
	d.TimestampForUpdate()
	d.Created = d.Modified
}
