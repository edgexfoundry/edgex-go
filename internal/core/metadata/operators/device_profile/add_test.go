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

package device_profile

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_profile/mocks"
)

type mockOutline struct {
	methodName string
	arg        interface{}
	ret        interface{}
	err        error
}

func TestDeleteExecutor(t *testing.T) {
	tests := []struct {
		testName         string
		mockAdder      DeviceProfileAdder
		deviceProfileBytes []byte
		expectedReturn string
		expectedError    bool
		expectedErrorVal error
	}{
		{
			testName: "Successful database call",
			mockAdder: createMockDeviceProfileAdder([]mockOutline{}),
			deviceProfileBytes: []byte{},
			expectedReturn: TestDeviceProfileID,
			expectedError:    false,
			expectedErrorVal: nil,
		},
		{
			testName: "Duplicate name",
			mockAdder: createMockDeviceProfileAdder([]mockOutline{}),
			deviceProfileBytes: []byte{},
			expectedReturn: "",
			expectedError:    true,
			expectedErrorVal: errors.NewErrDuplicateName("Duplicate profile name " + TestDeviceProfileName ),
		},
		{
			testName: "Empty device profile name",
			mockAdder: createMockDeviceProfileAdder([]mockOutline{}),
			deviceProfileBytes: []byte{},
			expectedReturn: "",
			expectedError:    true,
			expectedErrorVal: errors.NewErrEmptyDeviceProfileName(),
		},
		{
			testName: "Unsuccessful database call",
			mockAdder: createMockDeviceProfileAdder([]mockOutline{}),
			deviceProfileBytes: []byte{},
			expectedReturn: "",
			expectedError:    true,
			expectedErrorVal: TestError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(tt *testing.T) {
			op := NewAddDeviceProfileExecutor(test.deviceProfileBytes, test.mockAdder)
			id, err := op.Execute()

			if test.expectedReturn != id {
				t.Errorf("Observed return value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedReturn, id)
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

func createMockDeviceProfileAdder(outlines []mockOutline) DeviceProfileAdder {
	dbMock := mocks.DeviceProfileAdder{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg).Return(o.ret, o.err)
	}

	return &dbMock
}
