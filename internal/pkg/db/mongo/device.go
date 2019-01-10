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
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo/models"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Internal version of the device struct
// Use this to handle DBRef
type mongoDevice struct {
	contract.Device `bson:",inline"`
}

// Struct to hold the result of GetBSON
// This struct is used by MongoDeviceManager so that it can call GetBSON explicitly on MongoDevice
type mongoDeviceBSON struct {
	contract.DescribedObject `bson:",inline"`
	Id                       bson.ObjectId           `bson:"_id,omitempty"`
	Name                     string                  `bson:"name"`           // Unique name for identifying a device
	AdminState               contract.AdminState     `bson:"adminState"`     // Admin state (locked/unlocked)
	OperatingState           contract.OperatingState `bson:"operatingState"` // Operating state (enabled/disabled)
	Addressable              mgo.DBRef               `bson:"addressable"`    // Addressable for the device - stores information about it's address
	LastConnected            int64                   `bson:"lastConnected"`  // Time (milliseconds) that the device last provided any feedback or responded to any request
	LastReported             int64                   `bson:"lastReported"`   // Time (milliseconds) that the device reported data to the core microservice
	Labels                   []string                `bson:"labels"`         // Other labels applied to the device to help with searching
	Location                 interface{}             `bson:"location"`       // Device service specific location (interface{} is an empty interface so it can be anything)
	Service                  mgo.DBRef               `bson:"service"`        // Associated Device Service - One per device
	Profile                  mgo.DBRef               `bson:"profile"`        // Associated Device Profile - Describes the device
}

// Custom marshaling into mongo
func (md mongoDevice) GetBSON() (interface{}, error) {
	return mongoDeviceBSON{
		DescribedObject: md.DescribedObject,
		Id:              md.Id,
		Name:            md.Name,
		AdminState:      md.AdminState,
		OperatingState:  md.OperatingState,
		Addressable:     mgo.DBRef{Collection: db.Addressable, Id: md.Addressable.Id},
		LastConnected:   md.LastConnected,
		LastReported:    md.LastReported,
		Labels:          md.Labels,
		Location:        md.Location,
		Service:         mgo.DBRef{Collection: db.DeviceService, Id: md.Service.Service.Id},
		Profile:         mgo.DBRef{Collection: db.DeviceProfile, Id: md.Profile.Id},
	}, nil
}

// Custom unmarshaling out of mongo
func (md *mongoDevice) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		contract.DescribedObject `bson:",inline"`
		Id                       bson.ObjectId           `bson:"_id,omitempty"`
		Name                     string                  `bson:"name"`           // Unique name for identifying a device
		AdminState               contract.AdminState     `bson:"adminState"`     // Admin state (locked/unlocked)
		OperatingState           contract.OperatingState `bson:"operatingState"` // Operating state (enabled/disabled)
		Addressable              mgo.DBRef               `bson:"addressable"`    // Addressable for the device - stores information about it's address
		LastConnected            int64                   `bson:"lastConnected"`  // Time (milliseconds) that the device last provided any feedback or responded to any request
		LastReported             int64                   `bson:"lastReported"`   // Time (milliseconds) that the device reported data to the core microservice
		Labels                   []string                `bson:"labels"`         // Other labels applied to the device to help with searching
		Location                 interface{}             `bson:"location"`       // Device service specific location (interface{} is an empty interface so it can be anything)
		Service                  mgo.DBRef               `bson:"service"`        // Associated Device Service - One per device
		Profile                  mgo.DBRef               `bson:"profile"`        // Associated Device Profile - Describes the device
	})
	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	md.DescribedObject = decoded.DescribedObject
	md.Id = decoded.Id
	md.Name = decoded.Name
	md.AdminState = decoded.AdminState
	md.OperatingState = decoded.OperatingState
	md.LastConnected = decoded.LastConnected
	md.LastReported = decoded.LastReported
	md.Labels = decoded.Labels
	md.Location = decoded.Location

	// De-reference the DBRef fields

	m, err := getCurrentMongoClient()
	if err != nil {
		return err
	}
	s := m.session.Copy()
	defer s.Close()

	addCol := s.DB(m.database.Name).C(db.Addressable)
	dsCol := s.DB(m.database.Name).C(db.DeviceService)
	dpCol := s.DB(m.database.Name).C(db.DeviceProfile)

	var a models.Addressable
	var mdp mongoDeviceProfile
	var ds models.DeviceService

	err = addCol.Find(bson.M{"_id": decoded.Addressable.Id}).One(&a)
	if err == mgo.ErrNotFound {
		err = addCol.Find(bson.M{"uuid": decoded.Addressable.Id}).One(&a)
	}
	if err != nil {
		return err
	}

	err = dsCol.Find(bson.M{"_id": decoded.Service.Id}).One(&ds)
	if err != nil {
		return err
	}
	err = dpCol.Find(bson.M{"_id": decoded.Profile.Id}).One(&mdp)
	if err != nil {
		return err
	}

	md.Addressable = a.ToContract()
	md.Profile = mdp.DeviceProfile
	md.Service, err = ds.ToContract(m)
	return err
}
