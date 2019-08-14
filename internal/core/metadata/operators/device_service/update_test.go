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

package device_service

import (
	goErrors "errors"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_service/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var testOperatingState, _ = contract.GetOperatingState(contract.Enabled)
var testAdminState, _ = contract.GetAdminState(contract.Enabled)
var testError = goErrors.New("some error")

func TestUpdateOperatingStateByIdExecutor(t *testing.T) {
	operatingStateEnabled := testDeviceService
	operatingStateEnabled.OperatingState = testOperatingState

	tests := []struct {
		name             string
		mockUpdater      DeviceServiceUpdater
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
				{"UpdateDeviceService", operatingStateEnabled, nil, nil}}),
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Device service not found",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, db.ErrNotFound}}),
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(testDeviceServiceId),
		},
		{
			name: "Device service lookup error",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, testError}}),
			expectedError:    true,
			expectedErrorVal: testError,
		},
		{
			name: "Update error",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
				{"UpdateDeviceService", operatingStateEnabled, testError, nil}}),
			expectedError:    true,
			expectedErrorVal: testError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewUpdateOpStateByIdExecutor(testDeviceServiceId, testOperatingState, test.mockUpdater)
			err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}
		})
	}
}

func TestUpdateOperatingStateByNameExecutor(t *testing.T) {
	operatingStateEnabled := testDeviceService
	operatingStateEnabled.OperatingState = testOperatingState

	tests := []struct {
		name             string
		mockUpdater      DeviceServiceUpdater
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
				{"UpdateDeviceService", operatingStateEnabled, nil, nil}}),
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Device service not found",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, db.ErrNotFound}}),
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(testDeviceServiceName),
		},
		{
			name: "Device service lookup error",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, testError}}),
			expectedError:    true,
			expectedErrorVal: testError,
		},
		{
			name: "Update error",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
				{"UpdateDeviceService", operatingStateEnabled, testError, nil}}),
			expectedError:    true,
			expectedErrorVal: testError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewUpdateOpStateByNameExecutor(testDeviceServiceName, testOperatingState, test.mockUpdater)
			err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}
		})
	}
}

func TestUpdateAdminStateByIdExecutor(t *testing.T) {
	adminStateEnabled := testDeviceService
	adminStateEnabled.AdminState = testAdminState

	tests := []struct {
		name             string
		mockUpdater      DeviceServiceUpdater
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
				{"UpdateDeviceService", adminStateEnabled, nil, nil}}),
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Device service not found",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, db.ErrNotFound}}),
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(testDeviceServiceId),
		},
		{
			name: "Device service lookup error",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, testError}}),
			expectedError:    true,
			expectedErrorVal: testError,
		},
		{
			name: "Update error",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
				{"UpdateDeviceService", adminStateEnabled, testError, nil}}),
			expectedError:    true,
			expectedErrorVal: testError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewUpdateAdminStateByIdExecutor(testDeviceServiceId, testAdminState, test.mockUpdater)
			err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}
		})
	}
}

func TestUpdateAdminStateByNameExecutor(t *testing.T) {
	adminStateEnabled := testDeviceService
	adminStateEnabled.AdminState = testAdminState

	tests := []struct {
		name             string
		mockUpdater      DeviceServiceUpdater
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
				{"UpdateDeviceService", adminStateEnabled, nil, nil}}),
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Device service not found",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, db.ErrNotFound}}),
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(testDeviceServiceName),
		},
		{
			name: "Device service lookup error",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, testError}}),
			expectedError:    true,
			expectedErrorVal: testError,
		},
		{
			name: "Update error",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
				{"UpdateDeviceService", adminStateEnabled, testError, nil}}),
			expectedError:    true,
			expectedErrorVal: testError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewUpdateAdminStateByNameExecutor(testDeviceServiceName, testAdminState, test.mockUpdater)
			err := op.Execute()
			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}
		})
	}
}

type mockOutline struct {
	methodName string
	arg        interface{}
	ret        interface{}
	err        error
}

func createMockUpdater(outlines []mockOutline) DeviceServiceUpdater {
	dbMock := mocks.DeviceServiceUpdater{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg).Return(o.ret, o.err)
	}

	return &dbMock
}
