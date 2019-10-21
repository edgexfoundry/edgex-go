/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package errorconcept

import (
	"net/http"

	data "github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

var Database databaseErrorConcept

// DatabaseErrorConcept represents a collection of database error concepts
type databaseErrorConcept struct {
	NotFound        dbNotFound
	NotFoundTyped   dbNotFoundTyped
	NotUnique       dbNotUnique
	InvalidObjectId dbInvalidObjectId
}

type dbNotFound struct{}

func (r dbNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r dbNotFound) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r dbNotFound) message(err error) string {
	return err.Error()
}

type dbNotFoundTyped struct{}

func (r dbNotFoundTyped) httpErrorCode() int {
	return http.StatusNotFound
}

func (r dbNotFoundTyped) isA(err error) bool {
	_, ok := err.(data.ErrDbNotFound)
	return ok
}

func (r dbNotFoundTyped) message(err error) string {
	return err.Error()
}

type dbNotUnique struct{}

func (r dbNotUnique) httpErrorCode() int {
	return http.StatusConflict
}

func (r dbNotUnique) isA(err error) bool {
	return err == db.ErrNotUnique
}

func (r dbNotUnique) message(err error) string {
	return err.Error()
}

type dbInvalidObjectId struct{}

func (r dbInvalidObjectId) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r dbInvalidObjectId) isA(err error) bool {
	return err == db.ErrInvalidObjectId
}

func (r dbInvalidObjectId) message(err error) string {
	return err.Error()
}
