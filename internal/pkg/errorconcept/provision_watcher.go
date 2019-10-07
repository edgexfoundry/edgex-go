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

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

var ProvisionWatcher provisionWatcherErrorConcept

// ProvisionWatcherErrorConcept represents the accessor for provision-watcher-specific error concepts
type provisionWatcherErrorConcept struct {
	DeleteError_StatusInternalServer     provisionWatcherDeleteError_StatusInternalServer
	DeviceProfileNotFound_StatusConflict provisionWatcherDeviceProfileNotFound_StatusConflict
	DeviceProfileNotFound_StatusNotFound provisionWatcherDeviceProfileNotFound_StatusNotFound
	DeviceServiceNotFound_StatusConflict provisionWatcherDeviceServiceNotFound_StatusConflict
	DeviceServiceNotFound_StatusNotFound provisionWatcherDeviceServiceNotFound_StatusNotFound
	NotFoundById                         provisionWatcherNotFoundById
	NotFoundByName                       provisionWatcherNotFoundByName
	NotUnique                            provisionWatcherNotUnique
	RetrieveError_StatusNotFound         provisionWatcherRetrieveError_StatusNotFound
}

type provisionWatcherDeleteError_StatusInternalServer struct{}

func (r provisionWatcherDeleteError_StatusInternalServer) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r provisionWatcherDeleteError_StatusInternalServer) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r provisionWatcherDeleteError_StatusInternalServer) message(err error) string {
	return "Error deleting provision watcher"
}

type provisionWatcherDeviceProfileNotFound_StatusConflict struct{}

func (r provisionWatcherDeviceProfileNotFound_StatusConflict) httpErrorCode() int {
	return http.StatusConflict
}

func (r provisionWatcherDeviceProfileNotFound_StatusConflict) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r provisionWatcherDeviceProfileNotFound_StatusConflict) message(err error) string {
	return "Device profile not found for provision watcher"
}

type provisionWatcherDeviceProfileNotFound_StatusNotFound struct{}

func (r provisionWatcherDeviceProfileNotFound_StatusNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r provisionWatcherDeviceProfileNotFound_StatusNotFound) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r provisionWatcherDeviceProfileNotFound_StatusNotFound) message(err error) string {
	return "Device profile not found"
}

type provisionWatcherDeviceServiceNotFound_StatusConflict struct{}

func (r provisionWatcherDeviceServiceNotFound_StatusConflict) httpErrorCode() int {
	return http.StatusConflict
}

func (r provisionWatcherDeviceServiceNotFound_StatusConflict) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r provisionWatcherDeviceServiceNotFound_StatusConflict) message(err error) string {
	return "Device service not found for provision watcher"
}

type provisionWatcherDeviceServiceNotFound_StatusNotFound struct{}

func (r provisionWatcherDeviceServiceNotFound_StatusNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r provisionWatcherDeviceServiceNotFound_StatusNotFound) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r provisionWatcherDeviceServiceNotFound_StatusNotFound) message(err error) string {
	return "Device service not found"
}

type provisionWatcherNotFoundById struct{}

func (r provisionWatcherNotFoundById) httpErrorCode() int {
	return http.StatusNotFound
}

func (r provisionWatcherNotFoundById) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r provisionWatcherNotFoundById) message(err error) string {
	return "Provision Watcher not found by ID: " + err.Error()
}

type provisionWatcherNotFoundByName struct{}

func (r provisionWatcherNotFoundByName) httpErrorCode() int {
	return http.StatusNotFound
}

func (r provisionWatcherNotFoundByName) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r provisionWatcherNotFoundByName) message(err error) string {
	return "Provision Watcher not found: " + err.Error()
}

// NewProvisionWatcherDuplicateErrorConcept instantiates a new error concept for a given set of ids
func NewProvisionWatcherDuplicateErrorConcept(currentId string, newId string) provisionWatcherDuplicate {
	return provisionWatcherDuplicate{currentId: currentId, newId: newId}
}

type provisionWatcherDuplicate struct {
	currentId string
	newId     string
}

func (r provisionWatcherDuplicate) httpErrorCode() int {
	return http.StatusConflict
}

func (r provisionWatcherDuplicate) isA(err error) bool {
	return err != db.ErrNotFound && r.currentId != r.newId
}

func (r provisionWatcherDuplicate) message(err error) string {
	return "Duplicate name for the provision watcher"
}

type provisionWatcherNotUnique struct{}

func (r provisionWatcherNotUnique) httpErrorCode() int {
	return http.StatusConflict
}

func (r provisionWatcherNotUnique) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r provisionWatcherNotUnique) message(err error) string {
	return "Duplicate name for the provision watcher"
}

type provisionWatcherRetrieveError_StatusNotFound struct{}

func (r provisionWatcherRetrieveError_StatusNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r provisionWatcherRetrieveError_StatusNotFound) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r provisionWatcherRetrieveError_StatusNotFound) message(err error) string {
	return err.Error()
}
