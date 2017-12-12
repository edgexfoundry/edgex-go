package main

import (
	"bitbucket.org/clientcto/go-core-domain/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Internal version of the device service struct
// Use this to handle DBRef
type MongoDeviceService struct {
	models.DeviceService
}

// Custom marshaling into mongo
func (mds MongoDeviceService) GetBSON() (interface{}, error) {
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
		Addressable:     mgo.DBRef{Collection: ADDCOL, Id: mds.Service.Addressable.Id},
		LastConnected:   mds.Service.LastConnected,
		LastReported:    mds.Service.LastReported,
		Labels:          mds.Service.Labels,
	}, nil
}

// Custom unmarshaling out of mongo
func (mds *MongoDeviceService) SetBSON(raw bson.Raw) error {
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
	ds := DS.dataStore()
	defer ds.s.Close()

	addCol := ds.s.DB(DB).C(ADDCOL)

	var a models.Addressable

	err := addCol.Find(bson.M{"_id": decoded.Addressable.Id}).One(&a)
	if err != nil {
		return err
	}

	mds.Service.Addressable = a

	return nil
}
