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
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_service/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestGetAllDeviceServices(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.ServiceInfo
		dbMock      DeviceServiceLoader
		expectError bool
	}{
		{
			"GetAllPass",
			config.ServiceInfo{MaxResultCount: 1},
			createMockLoader([]mockOutline{
				{"GetAllDeviceServices", mock.Anything, testDeviceServices, nil},
			}),
			false,
		},
		{
			"GetAllFailCount",
			config.ServiceInfo{},
			createMockLoader([]mockOutline{
				{"GetAllDeviceServices", mock.Anything, testDeviceServices, nil},
			}),
			true,
		},
		{
			"GetAllFailUnexpected",
			config.ServiceInfo{MaxResultCount: 1},
			createMockLoader([]mockOutline{
				{"GetAllDeviceServices", mock.Anything, nil, testError},
			}),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewDeviceServiceLoadAll(tt.cfg, tt.dbMock, logger.MockLogger{})
			_, err := op.Execute()
			if err != nil && !tt.expectError {
				t.Error(err)
				return
			}
			if err == nil && tt.expectError {
				t.Errorf("error was expected, none occurred")
				return
			}
		})
	}
}

func TestGetDeviceServiceByName(t *testing.T) {
	tests := []struct {
		name             string
		mockLoader       DeviceServiceLoader
		expectedVal      contract.DeviceService
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call",
			mockLoader: createMockLoader([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
			}),
			expectedVal:      testDeviceService,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Device service not found",
			mockLoader: createMockLoader([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, db.ErrNotFound},
			}),
			expectedVal:      contract.DeviceService{},
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(testDeviceServiceName),
		},
		{
			name: "Device services lookup error",
			mockLoader: createMockLoader([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, testError},
			}),
			expectedVal:      contract.DeviceService{},
			expectedError:    true,
			expectedErrorVal: testError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewDeviceServiceLoadByName(testDeviceServiceName, test.mockLoader)
			actualVal, err := op.Execute()
			if !reflect.DeepEqual(test.expectedVal, actualVal) {
				t.Errorf("Observed value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedVal, actualVal)
				return
			}

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

func TestGetDeviceServiceById(t *testing.T) {
	tests := []struct {
		name             string
		mockLoader       DeviceServiceLoader
		expectedVal      contract.DeviceService
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call",
			mockLoader: createMockLoader([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
			}),
			expectedVal:      testDeviceService,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Device service not found",
			mockLoader: createMockLoader([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, db.ErrNotFound},
			}),
			expectedVal:      contract.DeviceService{},
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(testDeviceServiceId),
		},
		{
			name: "Device services lookup error",
			mockLoader: createMockLoader([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, testError},
			}),
			expectedVal:      contract.DeviceService{},
			expectedError:    true,
			expectedErrorVal: testError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewDeviceServiceLoadById(testDeviceServiceId, test.mockLoader)
			actualVal, err := op.Execute()
			if !reflect.DeepEqual(test.expectedVal, actualVal) {
				t.Errorf("Observed value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedVal, actualVal)
				return
			}

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

func TestGetDeviceServiceByAddressableId(t *testing.T) {
	tests := []struct {
		name             string
		mockLoader       DeviceServiceLoader
		value            string
		expectedVal      []contract.DeviceService
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableById", testDeviceServiceId, testAddressable, nil},
				{"GetDeviceServicesByAddressableId", testDeviceServiceId, testDeviceServices, nil},
			}),
			value:            testDeviceServiceId,
			expectedVal:      testDeviceServices,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "No ID provided",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableByName", "", contract.Addressable{}, db.ErrNotFound},
			}),
			value:            "",
			expectedVal:      nil,
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(""),
		},
		{
			name: "Addressable not found",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableById", testDeviceServiceId, contract.Addressable{}, db.ErrNotFound},
			}),
			value:            testDeviceServiceId,
			expectedVal:      nil,
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(testDeviceServiceId),
		},
		{
			name: "Addressable lookup error",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableById", testDeviceServiceId, contract.Addressable{}, testError},
			}),
			value:            testDeviceServiceId,
			expectedVal:      nil,
			expectedError:    true,
			expectedErrorVal: testError,
		},
		{
			name: "Device services lookup error",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableById", testDeviceServiceId, testAddressable, nil},
				{"GetDeviceServicesByAddressableId", testDeviceServiceId, nil, testError},
			}),
			value:            testDeviceServiceId,
			expectedVal:      nil,
			expectedError:    true,
			expectedErrorVal: testError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewDeviceServiceLoadByAddressableID(test.value, test.mockLoader)
			actualVal, err := op.Execute()
			if !reflect.DeepEqual(test.expectedVal, actualVal) {
				t.Errorf("Observed value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedVal, actualVal)
				return
			}

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

func TestGetDeviceServiceByAddressableName(t *testing.T) {
	tests := []struct {
		name             string
		mockLoader       DeviceServiceLoader
		value            string
		expectedVal      []contract.DeviceService
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableByName", testDeviceServiceName, testAddressable, nil},
				{"GetDeviceServicesByAddressableId", testDeviceServiceId, testDeviceServices, nil},
			}),
			value:            testDeviceServiceName,
			expectedVal:      testDeviceServices,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "No name provided",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableByName", "", contract.Addressable{}, db.ErrNotFound},
			}),
			value:            "",
			expectedVal:      nil,
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(""),
		},
		{
			name: "Addressable not found",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableByName", testDeviceServiceName, contract.Addressable{}, db.ErrNotFound},
			}),
			value:            testDeviceServiceName,
			expectedVal:      nil,
			expectedError:    true,
			expectedErrorVal: errors.NewErrItemNotFound(testDeviceServiceName),
		},
		{
			name: "Addressable lookup error",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableByName", testDeviceServiceName, contract.Addressable{}, testError},
			}),
			value:            testDeviceServiceName,
			expectedVal:      nil,
			expectedError:    true,
			expectedErrorVal: testError,
		},
		{
			name: "Device services lookup error",
			mockLoader: createMockLoader([]mockOutline{
				{"GetAddressableByName", testDeviceServiceName, testAddressable, nil},
				{"GetDeviceServicesByAddressableId", testDeviceServiceId, nil, testError},
			}),
			value:            testDeviceServiceName,
			expectedVal:      nil,
			expectedError:    true,
			expectedErrorVal: testError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewDeviceServiceLoadByAddressableName(test.value, test.mockLoader)
			actualVal, err := op.Execute()
			if !reflect.DeepEqual(test.expectedVal, actualVal) {
				t.Errorf("Observed value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedVal, actualVal)
				return
			}

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

func createMockLoader(outlines []mockOutline) DeviceServiceLoader {
	dbMock := mocks.DeviceServiceLoader{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg).Return(o.ret, o.err)
	}

	return &dbMock
}

var testAddressable = contract.Addressable{Id: testDeviceServiceId, Name: testDeviceServiceName}
var testDeviceServiceId = uuid.New().String()
var testDeviceServiceName = "test service"
var testDeviceService = contract.DeviceService{Id: testDeviceServiceId, Name: testDeviceServiceName}
var testDeviceServices = []contract.DeviceService{testDeviceService}
var testError = goErrors.New("some error")
