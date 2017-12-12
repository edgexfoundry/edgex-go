package main

import (
	"bitbucket.org/clientcto/go-core-domain/models"
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
