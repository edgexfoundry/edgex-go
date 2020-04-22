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
	"encoding/json"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Device
//
// Deprecated: Mongo functionality is deprecated as of the Geneva release.
type Device struct {
	Created        int64                   `bson:"created"`
	Modified       int64                   `bson:"modified"`
	Origin         int64                   `bson:"origin"`
	Description    string                  `bson:"description"`
	Id             bson.ObjectId           `bson:"_id,omitempty"`
	Uuid           string                  `bson:"uuid,omitempty"`
	Protocols      string                  `bson:"protocols,omitempty"`  //Contains a JSON representation of the supported protocols for the device
	AutoEvents     string                  `bson:"autoEvents,omitempty"` //Contains a JSON representation of the device's auto-generated events
	Name           string                  `bson:"name"`                 // Unique name for identifying a device
	AdminState     contract.AdminState     `bson:"adminState"`           // Admin state (locked/unlocked)
	OperatingState contract.OperatingState `bson:"operatingState"`       // Operating state (enabled/disabled)
	LastConnected  int64                   `bson:"lastConnected"`        // Time (milliseconds) that the device last provided any feedback or responded to any request
	LastReported   int64                   `bson:"lastReported"`         // Time (milliseconds) that the device reported data to the core microservice
	Labels         []string                `bson:"labels"`               // Other labels applied to the device to help with searching
	Location       interface{}             `bson:"location"`             // Device service specific location (interface{} is an empty interface so it can be anything)
	Service        mgo.DBRef               `bson:"service"`              // Associated Device Service - One per device
	Profile        mgo.DBRef               `bson:"profile"`              // Associated Device Profile - Describes the device
	ProfileName    string                  `bson:"profileName"`          // Associated Device Profile Name
}

func (d *Device) ToContract(dsTransform deviceServiceTransform, dpTransform deviceProfileTransform, aTransform addressableTransform) (contract.Device, error) {
	// Always hand back the UUID as the contract command ID unless it's blank (an old command, for example blackbox test scripts)
	id := d.Uuid
	if id == "" {
		id = d.Id.Hex()
	}

	var err error
	var result contract.Device

	result.Created = d.Created
	result.Modified = d.Modified
	result.Origin = d.Origin
	result.Description = d.Description
	result.Id = id
	result.Name = d.Name
	result.AdminState = d.AdminState
	result.OperatingState = d.OperatingState

	p := make(map[string]contract.ProtocolProperties)
	err = json.Unmarshal([]byte(d.Protocols), &p)
	if err != nil {
		return contract.Device{}, err
	}
	result.Protocols = p

	ae := make([]contract.AutoEvent, 0)
	err = json.Unmarshal([]byte(d.AutoEvents), &ae)
	if err != nil {
		return contract.Device{}, err
	}
	result.AutoEvents = ae
	result.LastConnected = d.LastConnected
	result.LastReported = d.LastReported
	result.Labels = d.Labels
	result.Location = d.Location

	dsModel, err := dsTransform.DBRefToDeviceService(d.Service)
	if err != nil {
		return contract.Device{}, err
	}
	result.Service, err = dsModel.ToContract(aTransform)
	if err != nil {
		return contract.Device{}, err
	}

	dpModel, err := dpTransform.DBRefToDeviceProfile(d.Profile)
	if err != nil {
		return contract.Device{}, err
	}
	result.Profile, err = dpModel.ToContract()
	if err != nil {
		return contract.Device{}, err
	}

	return result, nil
}

func (d *Device) FromContract(from contract.Device, dsTransform deviceServiceTransform, dpTransform deviceProfileTransform, aTransform addressableTransform) (string, error) {
	var err error
	if d.Id, d.Uuid, err = fromContractId(from.Id); err != nil {
		return "", err
	}

	d.Created = from.Created
	d.Modified = from.Modified
	d.Origin = from.Origin
	d.Description = from.Description
	d.Name = from.Name
	d.AdminState = from.AdminState
	d.OperatingState = from.OperatingState

	p, err := json.Marshal(from.Protocols)
	if err != nil {
		return "", err
	}
	d.Protocols = string(p)

	ae, err := json.Marshal(from.AutoEvents)
	if err != nil {
		return "", err
	}
	d.AutoEvents = string(ae)
	d.LastConnected = from.LastConnected
	d.LastReported = from.LastReported
	d.Labels = from.Labels
	d.Location = from.Location

	var dsModel DeviceService
	if _, err = dsModel.FromContract(from.Service, aTransform); err != nil {
		return "", err
	}
	if d.Service, err = dsTransform.DeviceServiceToDBRef(dsModel); err != nil {
		return "", err
	}

	var dpModel DeviceProfile
	if _, err = dpModel.FromContract(from.Profile); err != nil {
		return "", err
	}
	d.ProfileName = dpModel.Name
	if d.Profile, err = dpTransform.DeviceProfileToDBRef(dpModel); err != nil {
		return "", err
	}

	return toContractId(d.Id, d.Uuid), nil
}

func (d *Device) TimestampForUpdate() {
	d.Modified = db.MakeTimestamp()
}

func (d *Device) TimestampForAdd() {
	d.TimestampForUpdate()
	d.Created = d.Modified
}
