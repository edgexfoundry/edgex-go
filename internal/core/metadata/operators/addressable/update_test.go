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

func TestUpdateExecutor(t *testing.T) {
	successNoId := SuccessfulDatabaseResult[0]
	successNoId.Name = successNoId.Id
	successNoId.Id = ""

	successNewName := SuccessfulDatabaseResult[0]
	successNewName.Name = "something different"

	tests := []struct {
		name             string
		mockUpdater      AddressUpdater
		addr             contract.Addressable
		expectedError    bool
		expectedErrorVal error
	}{
		{
			name: "Successful database call, search by ID",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetAddressableById", SuccessfulDatabaseResult[0].Id, SuccessfulDatabaseResult[0], nil},
				{"UpdateAddressable", mock.Anything, nil, nil}}),
			addr:             SuccessfulDatabaseResult[0],
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Successful database call, search by name",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetAddressableByName", successNoId.Name, successNoId, nil},
				{"UpdateAddressable", mock.Anything, nil, nil}}),
			addr:             successNoId,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Successful database call, updated name",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetAddressableById", SuccessfulDatabaseResult[0].Id, SuccessfulDatabaseResult[0], nil},
				{"UpdateAddressable", mock.Anything, nil, nil},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{}, nil}}),
			addr:             successNewName,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			name: "Unsuccessful device service database call, updated name",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetAddressableById", SuccessfulDatabaseResult[0].Id, SuccessfulDatabaseResult[0], nil},
				{"GetDeviceServicesByAddressableId", mock.Anything, nil, Error}}),
			addr:             successNewName,
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name: "Addressable in use",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetAddressableById", SuccessfulDatabaseResult[0].Id, SuccessfulDatabaseResult[0], nil},
				{"UpdateAddressable", mock.Anything, nil, nil},
				{"GetDeviceServicesByAddressableId", mock.Anything, []contract.DeviceService{{}}, nil}}),
			addr:             successNewName,
			expectedError:    true,
			expectedErrorVal: errors.NewErrAddressableInUse(SuccessfulDatabaseResult[0].Name),
		},
		{
			name: "Unsuccessful database call retrieving addressable",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetAddressableById", SuccessfulDatabaseResult[0].Id, contract.Addressable{}, Error}}),
			addr:             SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name: "Unsuccessful database call updating addressable",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetAddressableById", SuccessfulDatabaseResult[0].Id, SuccessfulDatabaseResult[0], nil},
				{"UpdateAddressable", mock.Anything, Error, nil}}),
			addr:             SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: Error,
		},
		{
			name: "Addressable not found",
			mockUpdater: createMockUpdater([]mockOutline{
				{"GetAddressableById", SuccessfulDatabaseResult[0].Id, contract.Addressable{}, db.ErrNotFound}}),
			addr:             SuccessfulDatabaseResult[0],
			expectedError:    true,
			expectedErrorVal: errors.NewErrAddressableNotFound(SuccessfulDatabaseResult[0].Id, SuccessfulDatabaseResult[0].Name),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			op := NewUpdateExecutor(test.mockUpdater, test.addr)
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

func createMockUpdater(outlines []mockOutline) AddressUpdater {
	dbMock := mocks.AddressUpdater{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg).Return(o.ret, o.err)
	}

	return &dbMock
}
