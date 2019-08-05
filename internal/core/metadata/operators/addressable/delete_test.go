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

package addressable

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/addressable/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/mock"
)

func TestDeleteExecutor(t *testing.T) {
	success := SuccessfulDatabaseResult[0]

	tests := []struct {
		testName         string
		mockDeleter      AddressDeleter
		id               string
		name             string
		expectedError    bool
		expectedErrorVal error
	}{
		{
			testName: "Successful database call by ID",
			mockDeleter: createMockDeleter([]mockOutline{
				{"GetAddressableById", success.Id, success, nil},
				{"GetAddressableByName", mock.Anything, contract.Addressable{}, Error},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{}, nil},
				{"DeleteAddressableById", mock.Anything, nil, nil}}),
			id:               success.Id,
			name:             "",
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			testName: "Successful database call by name",
			mockDeleter: createMockDeleter([]mockOutline{
				{"GetAddressableByName", success.Name, success, nil},
				{"GetAddressableById", mock.Anything, contract.Addressable{}, Error},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{}, nil},
				{"DeleteAddressableById", mock.Anything, nil, nil}}),
			id:               "",
			name:             success.Name,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			testName: "Addressable not found",
			mockDeleter: createMockDeleter([]mockOutline{
				{"GetAddressableByName", success.Name, contract.Addressable{}, db.ErrNotFound},
				{"GetAddressableById", success.Id, contract.Addressable{}, db.ErrNotFound}}),
			id:               success.Id,
			name:             success.Name,
			expectedError:    true,
			expectedErrorVal: errors.NewErrAddressableNotFound(success.Id, success.Name),
		},
		{
			testName:         "No identifiers provided",
			mockDeleter:      nil,
			id:               "",
			name:             "",
			expectedError:    true,
			expectedErrorVal: errors.NewErrAddressableNotFound("", ""),
		},
		{
			testName: "Unsuccessful database call retrieving addressable",
			mockDeleter: createMockDeleter([]mockOutline{
				{"GetAddressableById", mock.Anything, contract.Addressable{}, Error},
				{"GetAddressableByName", mock.Anything, contract.Addressable{}, Error}}),
			id:               success.Id,
			name:             "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			testName: "Addressable in use",
			mockDeleter: createMockDeleter([]mockOutline{
				{"GetAddressableById", success.Id, success, nil},
				{"GetAddressableByName", mock.Anything, contract.Addressable{}, Error},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{{}}, nil},
				{"DeleteAddressableById", mock.Anything, nil, nil}}),
			id:               success.Id,
			name:             "",
			expectedError:    true,
			expectedErrorVal: errors.NewErrAddressableInUse(success.Name),
		},
		{
			testName: "Unsuccessful database call retrieving device services",
			mockDeleter: createMockDeleter([]mockOutline{
				{"GetAddressableById", success.Id, success, nil},
				{"GetAddressableByName", mock.Anything, contract.Addressable{}, Error},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{}, Error}}),
			id:               success.Id,
			name:             "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			testName: "Unsuccessful database call deleting addressable",
			mockDeleter: createMockDeleter([]mockOutline{
				{"GetAddressableById", success.Id, success, nil},
				{"GetAddressableByName", mock.Anything, contract.Addressable{}, Error},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{}, nil},
				{"DeleteAddressableById", mock.Anything, Error, nil}}),
			id:               success.Id,
			name:             "",
			expectedError:    true,
			expectedErrorVal: Error,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(tt *testing.T) {
			op := NewDeleteExecutor(test.mockDeleter, test.id, test.name)
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

func createMockDeleter(outlines []mockOutline) AddressDeleter {
	dbMock := mocks.AddressDeleter{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg).Return(o.ret, o.err)
	}

	return &dbMock
}
