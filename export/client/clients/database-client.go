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
 *
 * @microservice: export-client-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package clients

import (
	"errors"
	"fmt"

	"github.com/edgexfoundry/edgex-go/export"
	"gopkg.in/mgo.v2/bson"
)

type DatabaseType int8 // Database type enum
const (
	MONGO DatabaseType = iota
	MOCK
)

type DBClient interface {
	// ********************** REGISTRATION FUNCTIONS *****************************
	// Return all the registrations
	// UnexpectedError - failed to retrieve registrations from the database
	Registrations() ([]export.Registration, error)

	// Add a new registration
	// UnexpectedError - failed to add to database
	AddRegistration(reg *export.Registration) (bson.ObjectId, error)

	// Update a registration
	// UnexpectedError - problem updating in database
	// NotFound - no registration with the ID was found
	UpdateRegistration(reg export.Registration) error

	// Get a registration by ID
	// UnexpectedError - problem getting in database
	// NotFound - no registration with the ID was found
	RegistrationById(id string) (export.Registration, error)

	// Get a registration by name
	// UnexpectedError - problem getting in database
	// NotFound - no registration with the name was found
	RegistrationByName(name string) (export.Registration, error)

	// Delete a registration by ID
	// UnexpectedError - problem getting in database
	// NotFound - no registration with the ID was found
	DeleteRegistrationById(id string) error

	// Delete a registration by name
	// UnexpectedError - problem getting in database
	// NotFound - no registration with the ID was found
	DeleteRegistrationByName(name string) error
}

type DBConfiguration struct {
	DbType       DatabaseType
	Host         string
	Port         int
	Timeout      int
	DatabaseName string
	Username     string
	Password     string
}

var ErrNotFound error = errors.New("Item not found")
var ErrUnsupportedDatabase error = errors.New("Unsuppored database type")
var ErrInvalidObjectId error = errors.New("Invalid object ID")
var ErrNotUnique error = errors.New("Resource already exists")

// Return the dbClient interface
func NewDBClient(config DBConfiguration) (DBClient, error) {
	switch config.DbType {
	case MONGO:
		// Create the mongo client
		mc, err := newMongoClient(config)
		if err != nil {
			fmt.Println("Error creating the mongo client: " + err.Error())
			return nil, err
		}
		return mc, nil
	case MOCK:
		// Create the mock client
		mock := &MockDb{}
		return mock, nil
	default:
		return nil, ErrUnsupportedDatabase
	}
}
