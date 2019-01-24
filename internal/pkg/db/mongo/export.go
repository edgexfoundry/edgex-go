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
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo/models"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
)

// ****************************** REGISTRATIONS ********************************

// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (mc MongoClient) Registrations() ([]contract.Registration, error) {
	return mapRegistrations(mc.getRegistrations(bson.M{}))
}

// Add a new registration
// UnexpectedError - failed to add to database
func (mc MongoClient) AddRegistration(r contract.Registration) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var mapped models.Registration
	id, err := mapped.FromContract(r)
	if err != nil {
		return "", err
	}

	mapped.TimestampForAdd()

	if err = s.DB(mc.database.Name).C(db.ExportCollection).Insert(mapped); err != nil {
		return "", errorMap(err)
	}
	return id, nil
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (mc MongoClient) UpdateRegistration(reg contract.Registration) error {
	var mapped models.Registration
	id, err := mapped.FromContract(reg)
	if err != nil {
		return err
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.ExportCollection, id, mapped)
}

// Get a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc MongoClient) RegistrationById(id string) (contract.Registration, error) {
	reg, err := mc.registrationById(id)
	if err != nil {
		return contract.Registration{}, err
	}
	return reg.ToContract(), nil
}

func (mc MongoClient) registrationById(id string) (models.Registration, error) {
	query, err := idToBsonM(id)
	if err != nil {
		return models.Registration{}, err
	}

	reg, err := mc.getRegistration(query)
	if err != nil {
		return models.Registration{}, err
	}

	return reg, nil
}

// Get a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (mc MongoClient) RegistrationByName(name string) (contract.Registration, error) {
	reg, err := mc.registrationByName(name)
	if err != nil {
		return contract.Registration{}, err
	}
	return reg.ToContract(), nil
}

func (mc MongoClient) registrationByName(name string) (models.Registration, error) {
	reg, err := mc.getRegistration(bson.M{"name": name})
	if err != nil {
		return models.Registration{}, errorMap(err)
	}
	return reg, nil
}

// Delete a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc MongoClient) DeleteRegistrationById(id string) error {
	return mc.deleteById(db.ExportCollection, id)
}

// Delete a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (mc MongoClient) DeleteRegistrationByName(name string) error {
	return mc.deleteRegistration(bson.M{"name": name})
}

// Delete all registrations
func (mc MongoClient) ScrubAllRegistrations() error {
	s := mc.getSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(db.ExportCollection).RemoveAll(nil)
	return errorMap(err)
}

// Get registrations for the passed query
func (mc MongoClient) getRegistrations(q bson.M) ([]models.Registration, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var regs []models.Registration
	err := s.DB(mc.database.Name).C(db.ExportCollection).Find(q).All(&regs)
	if err != nil {
		return []models.Registration{}, errorMap(err)
	}

	return regs, nil
}

// Get a single registration for the passed query
func (mc MongoClient) getRegistration(q bson.M) (models.Registration, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var reg models.Registration
	err := s.DB(mc.database.Name).C(db.ExportCollection).Find(q).One(&reg)
	if err != nil {
		return models.Registration{}, errorMap(err)
	}
	return reg, nil
}

// Delete from the collection based on ID
func (mc MongoClient) deleteRegistration(q bson.M) error {
	s := mc.getSessionCopy()
	defer s.Close()

	return errorMap(s.DB(mc.database.Name).C(db.ExportCollection).Remove(q))
}

func mapRegistrations(registrations []models.Registration, err error) ([]contract.Registration, error) {
	if err != nil {
		return []contract.Registration{}, err
	}

	mapped := make([]contract.Registration, 0)
	for _, r := range registrations {
		mapped = append(mapped, r.ToContract())
	}
	return mapped, nil
}
