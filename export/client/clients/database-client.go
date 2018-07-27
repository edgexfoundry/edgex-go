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
package clients

import (
	"errors"

	"github.com/edgexfoundry/edgex-go/export"
	"gopkg.in/mgo.v2/bson"
)

type DatabaseType int8 // Database type enum
const (
	INVALID DatabaseType = iota
	MONGO
	MEMORY
	COUCH
)

const (
	invalidStr = "invalid"
	mongoStr   = "mongodb"
	memoryStr  = "memorydb"
	couchStr   = "couchdb"
)

// Add in order declared in Struct for string value
var databaseArr = [...]string{invalidStr, mongoStr, memoryStr}

func (db DatabaseType) String() string {
	if db >= INVALID && db <= MEMORY {
		return databaseArr[db]
	}
	return invalidStr
}

// Return enum value of the Database Type
func GetDatabaseType(db string) DatabaseType {
	switch db {
	case mongoStr:
		return MONGO
	case memoryStr:
		return MEMORY
	case couchStr:
		return COUCH
	default:
		return INVALID
	}
}

type DBClient interface {
	CloseSession()

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
		return newMongoClient(config)
	case MEMORY:
		return &memDB{}, nil
	case COUCH:
		return newCouchClient(config)
	default:
		return nil, ErrUnsupportedDatabase
	}
}
