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
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// ****************************** REGISTRATIONS ********************************

// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (mc *MongoClient) Registrations() ([]export.Registration, error) {
	return mc.getRegistrations(bson.M{})
}

// Add a new registration
// UnexpectedError - failed to add to database
func (mc *MongoClient) AddRegistration(reg *export.Registration) (bson.ObjectId, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	reg.Created = db.MakeTimestamp()
	reg.ID = bson.NewObjectId()

	// Add the registration
	err := s.DB(mc.database.Name).C(db.ExportCollection).Insert(reg)
	if err != nil {
		return reg.ID, err
	}

	return reg.ID, err
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (mc *MongoClient) UpdateRegistration(reg export.Registration) error {
	s := mc.getSessionCopy()
	defer s.Close()

	reg.Modified = db.MakeTimestamp()

	err := s.DB(mc.database.Name).C(db.ExportCollection).UpdateId(reg.ID, reg)
	if err == mgo.ErrNotFound {
		return db.ErrNotFound
	}

	return err
}

// Get a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc *MongoClient) RegistrationById(id string) (export.Registration, error) {
	if !bson.IsObjectIdHex(id) {
		return export.Registration{}, db.ErrInvalidObjectId
	}
	return mc.getRegistration(bson.M{"_id": bson.ObjectIdHex(id)})
}

// Get a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (mc *MongoClient) RegistrationByName(name string) (export.Registration, error) {
	return mc.getRegistration(bson.M{"name": name})
}

// Delete a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc *MongoClient) DeleteRegistrationById(id string) error {
	if !bson.IsObjectIdHex(id) {
		return db.ErrInvalidObjectId
	}
	return mc.deleteRegistration(bson.M{"_id": bson.ObjectIdHex(id)})
}

// Delete a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc *MongoClient) DeleteRegistrationByName(name string) error {
	return mc.deleteRegistration(bson.M{"name": name})
}

// Delete all registrations
func (mc *MongoClient) ScrubAllRegistrations() error {
	s := mc.getSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(db.ExportCollection).RemoveAll(nil)
	return err
}

// Get registrations for the passed query
func (mc *MongoClient) getRegistrations(q bson.M) ([]export.Registration, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var regs []export.Registration
	err := s.DB(mc.database.Name).C(db.ExportCollection).Find(q).All(&regs)
	if err != nil {
		return regs, err
	}

	return regs, nil
}

// Get a single registration for the passed query
func (mc *MongoClient) getRegistration(q bson.M) (export.Registration, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var reg export.Registration
	err := s.DB(mc.database.Name).C(db.ExportCollection).Find(q).One(&reg)
	if err == mgo.ErrNotFound {
		return reg, db.ErrNotFound
	}

	return reg, err
}

// Delete from the collection based on ID
func (mc *MongoClient) deleteRegistration(q bson.M) error {
	s := mc.getSessionCopy()
	defer s.Close()

	err := s.DB(mc.database.Name).C(db.ExportCollection).Remove(q)
	if err == mgo.ErrNotFound {
		return db.ErrNotFound
	}
	return err
}
