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

	metadataErrors "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var DeviceProfile deviceProfileErrorConcept

// DeviceProfileErrorConcept represents the accessor for the device-profile-specific error concepts
type deviceProfileErrorConcept struct {
	ContractInvalid_StatusConflict         deviceProfileContractInvalid_StatusConflict
	DuplicateName                          deviceProfileDuplicateName
	EmptyName                              deviceProfileEmptyName
	InvalidState_StatusBadRequest          deviceProfileInvalidState_StatusBadRequest
	InvalidState_StatusConflict            deviceProfileInvalidState_StatusConflict
	MarshalYaml                            deviceProfileMarshalYaml
	MissingFile                            deviceProfileMissingFile
	NotFound                               deviceProfileNotFound
	ReadFile                               deviceProfileReadFile
	UnmarshalYaml_StatusInternalServer     deviceProfileUnmarshalYaml_StatusInternalServer
	UnmarshalYaml_StatusServiceUnavailable deviceProfileUnmarshalYaml_StatusServiceUnavailable
}

type deviceProfileContractInvalid_StatusConflict struct{}

func (r deviceProfileContractInvalid_StatusConflict) httpErrorCode() int {
	return http.StatusConflict
}

func (r deviceProfileContractInvalid_StatusConflict) isA(err error) bool {
	_, ok := err.(models.ErrContractInvalid)
	return ok
}

func (r deviceProfileContractInvalid_StatusConflict) message(err error) string {
	return err.Error()
}

type deviceProfileDuplicateName struct {
	err error
}

func (r deviceProfileDuplicateName) httpErrorCode() int {
	return http.StatusConflict
}

func (r deviceProfileDuplicateName) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceProfileDuplicateName) message(err error) string {
	return "Duplicate name for device profile"
}

type deviceProfileEmptyName struct{}

func (r deviceProfileEmptyName) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r deviceProfileEmptyName) isA(err error) bool {
	_, ok := err.(metadataErrors.ErrEmptyDeviceProfileName)
	return ok
}

func (r deviceProfileEmptyName) message(err error) string {
	return err.Error()
}

type deviceProfileInvalidState_StatusBadRequest struct{}

func (r deviceProfileInvalidState_StatusBadRequest) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r deviceProfileInvalidState_StatusBadRequest) isA(err error) bool {
	_, ok := err.(metadataErrors.ErrDeviceProfileInvalidState)
	return ok
}

func (r deviceProfileInvalidState_StatusBadRequest) message(err error) string {
	return err.Error()
}

type deviceProfileInvalidState_StatusConflict struct{}

func (r deviceProfileInvalidState_StatusConflict) httpErrorCode() int {
	return http.StatusConflict
}

func (r deviceProfileInvalidState_StatusConflict) isA(err error) bool {
	_, ok := err.(metadataErrors.ErrDeviceProfileInvalidState)
	return ok
}

func (r deviceProfileInvalidState_StatusConflict) message(err error) string {
	return err.Error()
}

type deviceProfileMarshalYaml struct{}

func (r deviceProfileMarshalYaml) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r deviceProfileMarshalYaml) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceProfileMarshalYaml) message(err error) string {
	return err.Error()
}

type deviceProfileMissingFile struct{}

func (r deviceProfileMissingFile) httpErrorCode() int {
	return http.StatusBadRequest
}

func (r deviceProfileMissingFile) isA(err error) bool {
	return err == http.ErrMissingFile
}

func (r deviceProfileMissingFile) message(err error) string {
	return metadataErrors.NewErrEmptyFile("YAML").Error()
}

type deviceProfileNotFound struct{}

func (r deviceProfileNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r deviceProfileNotFound) isA(err error) bool {
	_, ok := err.(metadataErrors.ErrDeviceProfileNotFound)
	return ok
}

func (r deviceProfileNotFound) message(err error) string {
	return err.Error()
}

type deviceProfileReadFile struct{}

func (r deviceProfileReadFile) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r deviceProfileReadFile) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceProfileReadFile) message(err error) string {
	return err.Error()
}

type deviceProfileUnmarshalYaml_StatusInternalServer struct{}

func (r deviceProfileUnmarshalYaml_StatusInternalServer) httpErrorCode() int {
	return http.StatusInternalServerError
}

func (r deviceProfileUnmarshalYaml_StatusInternalServer) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceProfileUnmarshalYaml_StatusInternalServer) message(err error) string {
	return err.Error()
}

type deviceProfileUnmarshalYaml_StatusServiceUnavailable struct{}

func (r deviceProfileUnmarshalYaml_StatusServiceUnavailable) httpErrorCode() int {
	return http.StatusServiceUnavailable
}

func (r deviceProfileUnmarshalYaml_StatusServiceUnavailable) isA(err error) bool {
	panic(METHOD_NOT_ALLOWED)
}

func (r deviceProfileUnmarshalYaml_StatusServiceUnavailable) message(err error) string {
	return err.Error()
}
