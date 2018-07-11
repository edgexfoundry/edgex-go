/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package mongo

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Internal version of the device service struct
// Use this to handle DBRef
type mongoDeviceService struct {
	models.DeviceService
}

// Custom marshaling into mongo
func (mds mongoDeviceService) GetBSON() (interface{}, error) {
	return struct {
		models.DescribedObject `bson:",inline"`
		Id                     bson.ObjectId         `bson:"_id,omitempty"`
		Name                   string                `bson:"name"`           // time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected          int64                 `bson:"lastConnected"`  // time in milliseconds that the device last reported data to the core
		LastReported           int64                 `bson:"lastReported"`   // operational state - either enabled or disabled
		OperatingState         models.OperatingState `bson:"operatingState"` // operational state - ether enabled or disableddc
		Labels                 []string              `bson:"labels"`         // tags or other labels applied to the device service for search or other identification needs
		Addressable            mgo.DBRef             `bson:"addressable"`    // address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
		AdminState             models.AdminState     `bson:"adminState"`     // Device Service Admin State
	}{
		DescribedObject: mds.Service.DescribedObject,
		Id:              mds.Service.Id,
		Name:            mds.Service.Name,
		AdminState:      mds.AdminState,
		OperatingState:  mds.Service.OperatingState,
		Addressable:     mgo.DBRef{Collection: db.Addressable, Id: mds.Service.Addressable.Id},
		LastConnected:   mds.Service.LastConnected,
		LastReported:    mds.Service.LastReported,
		Labels:          mds.Service.Labels,
	}, nil
}

// Custom unmarshaling out of mongo
func (mds *mongoDeviceService) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		models.DescribedObject `bson:",inline"`
		Id                     bson.ObjectId         `bson:"_id,omitempty"`
		Name                   string                `bson:"name"`           // time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected          int64                 `bson:"lastConnected"`  // time in milliseconds that the device last reported data to the core
		LastReported           int64                 `bson:"lastReported"`   // operational state - either enabled or disabled
		OperatingState         models.OperatingState `bson:"operatingState"` // operational state - ether enabled or disableddc
		Labels                 []string              `bson:"labels"`         // tags or other labels applied to the device service for search or other identification needs
		Addressable            mgo.DBRef             `bson:"addressable"`    // address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
		AdminState             models.AdminState     `bson:"adminState"`     // Device Service Admin State
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	mds.Service.DescribedObject = decoded.DescribedObject
	mds.Service.Id = decoded.Id
	mds.Service.Name = decoded.Name
	mds.AdminState = decoded.AdminState
	mds.Service.OperatingState = decoded.OperatingState
	mds.Service.LastConnected = decoded.LastConnected
	mds.Service.LastReported = decoded.LastReported
	mds.Service.Labels = decoded.Labels

	// De-reference the DBRef fields
	m, err := getCurrentMongoClient()
	if err != nil {
		return err
	}
	s := m.session.Copy()
	defer s.Close()

	addCol := s.DB(m.database.Name).C(db.Addressable)

	var a models.Addressable

	err = addCol.Find(bson.M{"_id": decoded.Addressable.Id}).One(&a)
	if err != nil {
		return err
	}

	mds.Service.Addressable = a

	return nil
}
