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
package mongo

import (
	"fmt"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	EXPORT_COLLECTION = "exportConfiguration"
)

/*
Export client client
Has functions for interacting with the export client mongo database
*/

type ExportMongoClient struct {
	Session  *mgo.Session  // Mongo database session
	Database *mgo.Database // Mongo database
}

// Return a pointer to the ExportMongoClient
func NewExportMongoClient(config db.Configuration) (*ExportMongoClient, error) {
	// Create the dial info for the Mongo session
	connectionString := config.Host + ":" + strconv.Itoa(config.Port)
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{connectionString},
		Timeout:  time.Duration(config.Timeout) * time.Millisecond,
		Database: config.DatabaseName,
		Username: config.Username,
		Password: config.Password,
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return nil, fmt.Errorf("Error dialing the mongo server: " + err.Error())
	}

	mongoClient := &ExportMongoClient{Session: session, Database: session.DB(config.DatabaseName)}
	return mongoClient, nil
}

// Get a copy of the session
func (mc *ExportMongoClient) GetSessionCopy() *mgo.Session {
	return mc.Session.Copy()
}

func (mc *ExportMongoClient) CloseSession() {
	mc.Session.Close()
}

// ****************************** REGISTRATIONS ********************************

// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (mc *ExportMongoClient) Registrations() ([]export.Registration, error) {
	return mc.getRegistrations(bson.M{})
}

// Add a new registration
// UnexpectedError - failed to add to database
func (mc *ExportMongoClient) AddRegistration(reg *export.Registration) (bson.ObjectId, error) {
	s := mc.GetSessionCopy()
	defer s.Close()

	reg.Created = db.MakeTimestamp()
	reg.ID = bson.NewObjectId()

	// Add the registration
	err := s.DB(mc.Database.Name).C(EXPORT_COLLECTION).Insert(reg)
	if err != nil {
		return reg.ID, err
	}

	return reg.ID, err
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (mc *ExportMongoClient) UpdateRegistration(reg export.Registration) error {
	s := mc.GetSessionCopy()
	defer s.Close()

	reg.Modified = db.MakeTimestamp()

	err := s.DB(mc.Database.Name).C(EXPORT_COLLECTION).UpdateId(reg.ID, reg)
	if err == mgo.ErrNotFound {
		return db.ErrNotFound
	}

	return err
}

// Get a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc *ExportMongoClient) RegistrationById(id string) (export.Registration, error) {
	if !bson.IsObjectIdHex(id) {
		return export.Registration{}, db.ErrInvalidObjectId
	}
	return mc.getRegistration(bson.M{"_id": bson.ObjectIdHex(id)})
}

// Get a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (mc *ExportMongoClient) RegistrationByName(name string) (export.Registration, error) {
	return mc.getRegistration(bson.M{"name": name})
}

// Delete a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc *ExportMongoClient) DeleteRegistrationById(id string) error {
	if !bson.IsObjectIdHex(id) {
		return db.ErrInvalidObjectId
	}
	return mc.deleteRegistration(bson.M{"_id": bson.ObjectIdHex(id)})
}

// Delete a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc *ExportMongoClient) DeleteRegistrationByName(name string) error {
	return mc.deleteRegistration(bson.M{"name": name})
}

// Get registrations for the passed query
func (mc *ExportMongoClient) getRegistrations(q bson.M) ([]export.Registration, error) {
	s := mc.GetSessionCopy()
	defer s.Close()

	var regs []export.Registration
	err := s.DB(mc.Database.Name).C(EXPORT_COLLECTION).Find(q).All(&regs)
	if err != nil {
		return regs, err
	}

	return regs, nil
}

// Get a single registration for the passed query
func (mc *ExportMongoClient) getRegistration(q bson.M) (export.Registration, error) {
	s := mc.GetSessionCopy()
	defer s.Close()

	var reg export.Registration
	err := s.DB(mc.Database.Name).C(EXPORT_COLLECTION).Find(q).One(&reg)
	if err == mgo.ErrNotFound {
		return reg, db.ErrNotFound
	}

	return reg, err
}

// Delete from the collection based on ID
func (mc *ExportMongoClient) deleteRegistration(q bson.M) error {
	s := mc.GetSessionCopy()
	defer s.Close()

	err := s.DB(mc.Database.Name).C(EXPORT_COLLECTION).Remove(q)
	if err == mgo.ErrNotFound {
		return db.ErrNotFound
	}
	return err
}
