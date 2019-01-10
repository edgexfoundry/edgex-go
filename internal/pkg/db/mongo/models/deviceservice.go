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

type DeviceService struct {
	Service    `bson:",inline"`
	AdminState contract.AdminState `bson:"adminState"` // Device Service Admin State
}

func (ds *DeviceService) ToContract(transform addressableTransform) (c contract.DeviceService, err error) {
	s, err := ds.Service.ToContract(transform)
	if err != nil {
		return
	}
	c.Service = s
	c.AdminState = ds.AdminState
	return
}

func (ds *DeviceService) FromContract(from contract.DeviceService, transform addressableTransform) (err error) {
	ds.AdminState = from.AdminState
	err = ds.Service.FromContract(from.Service, transform)
	return
}

// Custom marshaling into mongo
func (ds *DeviceService) GetBSON() (interface{}, error) {
	return struct {
		DescribedObject `bson:",inline"`
		Id              bson.ObjectId           `bson:"_id,omitempty"`
		Name            string                  `bson:"name"`           // time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected   int64                   `bson:"lastConnected"`  // time in milliseconds that the device last reported data to the core
		LastReported    int64                   `bson:"lastReported"`   // operational state - either enabled or disabled
		OperatingState  contract.OperatingState `bson:"operatingState"` // operational state - ether enabled or disabled
		Labels          []string                `bson:"labels"`         // tags or other labels applied to the device service for search or other identification needs
		Addressable     mgo.DBRef               `bson:"addressable"`    // address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
		AdminState      contract.AdminState     `bson:"adminState"`     // Device Service Admin State
	}{
		DescribedObject: ds.Service.DescribedObject,
		Id:              ds.Service.Id,
		Name:            ds.Service.Name,
		AdminState:      ds.AdminState,
		OperatingState:  ds.Service.OperatingState,
		Addressable:     mgo.DBRef{Collection: db.Addressable, Id: ds.Service.Addressable.Id},
		LastConnected:   ds.Service.LastConnected,
		LastReported:    ds.Service.LastReported,
		Labels:          ds.Service.Labels,
	}, nil
}

// Custom marshalling out of mongo
func (ds *DeviceService) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		DescribedObject `bson:",inline"`
		Id              bson.ObjectId           `bson:"_id,omitempty"`
		Name            string                  `bson:"name"`           // time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected   int64                   `bson:"lastConnected"`  // time in milliseconds that the device last reported data to the core
		LastReported    int64                   `bson:"lastReported"`   // operational state - either enabled or disabled
		OperatingState  contract.OperatingState `bson:"operatingState"` // operational state - ether enabled or disabled
		Labels          []string                `bson:"labels"`         // tags or other labels applied to the device service for search or other identification needs
		Addressable     mgo.DBRef               `bson:"addressable"`    // address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
		AdminState      contract.AdminState     `bson:"adminState"`     // Device Service Admin State
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	ds.Service.DescribedObject = decoded.DescribedObject
	ds.Service.Id = decoded.Id
	ds.Service.Name = decoded.Name
	ds.AdminState = decoded.AdminState
	ds.Service.OperatingState = decoded.OperatingState
	ds.Service.LastConnected = decoded.LastConnected
	ds.Service.LastReported = decoded.LastReported
	ds.Service.Labels = decoded.Labels
	ds.Service.Addressable = decoded.Addressable

	return nil
}
