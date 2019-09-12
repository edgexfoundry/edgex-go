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

var DeviceReport deviceReportErrorConcept

// DeviceReportErrorConcept represents the accessor for the device-report-specific error concepts
type deviceReportErrorConcept struct {
	DeviceNotFound deviceReportDeviceNotFound
	NotUnique      deviceReportNotUnique
}

type deviceReportDeviceNotFound struct{}

func (r deviceReportDeviceNotFound) httpErrorCode() int {
	return http.StatusNotFound
}

func (r deviceReportDeviceNotFound) isA(err error) bool {
	return err == db.ErrNotFound
}

func (r deviceReportDeviceNotFound) message(err error) string {
	return "Device referenced by Device Report doesn't exist"
}

type deviceReportNotUnique struct{}

func (r deviceReportNotUnique) httpErrorCode() int {
	return http.StatusConflict
}

func (r deviceReportNotUnique) isA(err error) bool {
	return err == db.ErrNotUnique
}

func (r deviceReportNotUnique) message(err error) string {
	return "Duplicate Name for the device report"
}
