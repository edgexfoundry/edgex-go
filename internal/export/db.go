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
package export

import (
	"github.com/globalsign/mgo/bson"
)

type DBClient interface {
	CloseSession()

	Connect() error

	// ********************** REGISTRATION FUNCTIONS *****************************
	// Return all the registrations
	// UnexpectedError - failed to retrieve registrations from the database
	Registrations() ([]Registration, error)

	// Add a new registration
	// UnexpectedError - failed to add to database
	AddRegistration(reg *Registration) (bson.ObjectId, error)

	// Update a registration
	// UnexpectedError - problem updating in database
	// NotFound - no registration with the ID was found
	UpdateRegistration(reg Registration) error

	// Get a registration by ID
	// UnexpectedError - problem getting in database
	// NotFound - no registration with the ID was found
	RegistrationById(id string) (Registration, error)

	// Get a registration by name
	// UnexpectedError - problem getting in database
	// NotFound - no registration with the name was found
	RegistrationByName(name string) (Registration, error)

	// Delete a registration by ID
	// UnexpectedError - problem getting in database
	// NotFound - no registration with the ID was found
	DeleteRegistrationById(id string) error

	// Delete a registration by name
	// UnexpectedError - problem getting in database
	// NotFound - no registration with the ID was found
	DeleteRegistrationByName(name string) error

	// Delete all registrations
	ScrubAllRegistrations() error
}
