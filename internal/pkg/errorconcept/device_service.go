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

var DeviceService deviceServiceErrorConcept

// DeviceServiceErrorConcept represents the accessor for the device-service-specific error concepts
type deviceServiceErrorConcept struct {
	AddressableNotFound                 deviceServiceAddressableNotFound
	EmptyAddressable                    deviceServiceEmptyAddressable
	InvalidRequest_StatusInternalServer deviceServiceInvalidRequest_StatusInternalServer
	InvalidState                        deviceServiceInvalidState
	NotUnique                           deviceServiceNotUnique
	NotFound                            deviceServiceNotFound
}

type deviceServiceAddressableNotFound struct{}

func (r deviceServiceAddressableNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r deviceServiceAddressableNotFound) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r deviceServiceAddressableNotFound) message(err error) string {
	return "Addressable not found by ID or Name"
}

type deviceServiceEmptyAddressable struct{}

func (r deviceServiceEmptyAddressable) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r deviceServiceEmptyAddressable) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceServiceEmptyAddressable) message(err error) string {
	return "Must provide an Addressable for Device Service"
}

type deviceServiceInvalidRequest_StatusInternalServer struct{}

func (r deviceServiceInvalidRequest_StatusInternalServer) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r deviceServiceInvalidRequest_StatusInternalServer) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceServiceInvalidRequest_StatusInternalServer) message(err error) string {
	return err.Error()
}

type deviceServiceInvalidState struct{}

func (r deviceServiceInvalidState) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r deviceServiceInvalidState) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceServiceInvalidState) message(err error) string {
	return err.Error()
}

// NewDeviceServiceDuplicate instantiates a new deviceServiceDuplicate error concept in effort to handle stateful
// DeviceService duplicate errors
func NewDeviceServiceDuplicate(currentDSId string, newDSId string) deviceServiceDuplicate {
	return deviceServiceDuplicate{CurrentDSId: currentDSId, NewDSId: newDSId}
}

type deviceServiceDuplicate struct {
	CurrentDSId string
	NewDSId     string
}

func (r deviceServiceDuplicate) httpErrorCode() int {
	return http.StatusConflict
}

func (r deviceServiceDuplicate) isA(err error) bool {
	return r.CurrentDSId != r.NewDSId
}

func (r deviceServiceDuplicate) message(err error) string {
	return "Duplicate name for Device Service"
}

type deviceServiceNotUnique struct{}

func (r deviceServiceNotUnique) httpErrorCode() int {
	return http.StatusConflict
}

func (r deviceServiceNotUnique) isA(err error) bool {
	return err == db.ErrNotUnique
}

func (r deviceServiceNotUnique) message(err error) string {
	return "Duplicate name for the device service"
}

type deviceServiceNotFound struct{}

func (r deviceServiceNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r deviceServiceNotFound) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r deviceServiceNotFound) message(err error) string {
	return "Device service not found"
}
