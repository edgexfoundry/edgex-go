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
 *
 * @microservice: core-metadata-go service
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package metadata

import (
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Internal version of the schedule event struct
// Use this to handle DBRef
type MongoScheduleEvent struct {
	models.ScheduleEvent
}

// Custom marshaling into mongo
func (mse MongoScheduleEvent) GetBSON() (interface{}, error) {
	return struct {
		models.BaseObject `bson:",inline"`
		Id                bson.ObjectId `bson:"_id,omitempty"`
		Name              string        `bson:"name"`        // non-database unique identifier for a schedule event
		Schedule          string        `bson:"schedule"`    // Name to associated owning schedule
		Addressable       mgo.DBRef     `bson:"addressable"` // address {MQTT topic, HTTP address, serial bus, etc.} for the action (can be empty)
		Parameters        string        `bson:"parameters"`  // json body for parameters
		Service           string        `bson:"service"`     // json body for parameters
	}{
		BaseObject:  mse.BaseObject,
		Id:          mse.Id,
		Name:        mse.Name,
		Schedule:    mse.Schedule,
		Parameters:  mse.Parameters,
		Service:     mse.Service,
		Addressable: mgo.DBRef{Collection: ADDCOL, Id: mse.Addressable.Id},
	}, nil
}

// Custom unmarshaling out of mongo
func (mse *MongoScheduleEvent) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		models.BaseObject `bson:",inline"`
		Id                bson.ObjectId `bson:"_id,omitempty"`
		Name              string        `bson:"name"`        // non-database unique identifier for a schedule event
		Schedule          string        `bson:"schedule"`    // Name to associated owning schedule
		Addressable       mgo.DBRef     `bson:"addressable"` // address {MQTT topic, HTTP address, serial bus, etc.} for the action (can be empty)
		Parameters        string        `bson:"parameters"`  // json body for parameters
		Service           string        `bson:"service"`     // json body for parameters
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	mse.BaseObject = decoded.BaseObject
	mse.Id = decoded.Id
	mse.Name = decoded.Name
	mse.Schedule = decoded.Schedule
	mse.Parameters = decoded.Parameters
	mse.Service = decoded.Service

	// De-reference the DBRef fields
	ds := DS.dataStore()
	defer ds.s.Close()

	addCol := ds.s.DB(DB).C(ADDCOL)

	var a models.Addressable

	if err := addCol.FindId(decoded.Addressable.Id).One(&a); err != nil {
		return err
	}

	mse.Addressable = a

	return nil
}
