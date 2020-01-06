/*
 * ******************************************************************************
 *  Copyright 2019 Dell Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software distributed under the License
 *  is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 *  or implied. See the License for the specific language governing permissions and limitations under
 *  the License.
 *  ******************************************************************************
 */

package value_descriptor

import (
	goErrors "errors"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/operators/value_descriptor/mocks"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var TestSuccessfulConfig = bootstrapConfig.ServiceInfo{MaxResultCount: 5}
var TestError = goErrors.New("test error")

var TestVDDescription = "test description"
var TestVDName = "Temperature"
var TestMin = -70
var TestMax = 140
var TestUoMLabel = "C"
var TestDefaultValue = 32
var TestFormatting = "%d"
var TestVDLabels = []string{"temp", "room temp"}
var TestVDFloatEncoding = contract.ENotation
var TestMediaType = "TestMediaType"
var TestValueDescriptor = contract.ValueDescriptor{Created: 123, Modified: 123, Origin: 123, Name: TestVDName, Description: TestVDDescription, Min: TestMin, Max: TestMax, DefaultValue: TestDefaultValue, Formatting: TestFormatting, Labels: TestVDLabels, UomLabel: TestUoMLabel, MediaType: TestMediaType, FloatEncoding: TestVDFloatEncoding}

// Value descriptors which share the same name.
var TestVDNameSharedName = "TestSharedName"
var TestValueDescriptor2 = contract.ValueDescriptor{Created: 123, Modified: 123, Origin: 123, Name: TestVDNameSharedName, Description: TestVDDescription, Min: TestMin, Max: TestMax, DefaultValue: TestDefaultValue, Formatting: TestFormatting, Labels: TestVDLabels, UomLabel: TestUoMLabel, MediaType: TestMediaType, FloatEncoding: TestVDFloatEncoding}
var TestValueDescriptor3 = contract.ValueDescriptor{Created: 123, Modified: 123, Origin: 123, Name: TestVDNameSharedName, Description: TestVDDescription, Min: TestMin, Max: TestMax, DefaultValue: TestDefaultValue, Formatting: TestFormatting, Labels: TestVDLabels, UomLabel: TestUoMLabel, MediaType: TestMediaType, FloatEncoding: TestVDFloatEncoding}

var TestVDErrorNames = []string{"TestVDError1", "TestVDError2"}

func TestGetValueDescriptorsByNames(t *testing.T) {
	tests := []struct {
		name              string
		loader            Loader
		vdNames           []string
		config            bootstrapConfig.ServiceInfo
		expectedResult    []contract.ValueDescriptor
		expectError       bool
		expectedErrorType error
	}{
		{
			"Get by one name",
			createMockLoader(),
			[]string{TestVDName},
			TestSuccessfulConfig,
			[]contract.ValueDescriptor{TestValueDescriptor},
			false,
			nil,
		},
		{
			"Get by multiple names",
			createMockLoader(),
			[]string{TestVDName, TestVDNameSharedName},
			TestSuccessfulConfig,
			[]contract.ValueDescriptor{TestValueDescriptor, TestValueDescriptor2, TestValueDescriptor3},
			false,
			nil,
		},
		{
			"Database error",
			createMockLoader(),
			TestVDErrorNames,
			TestSuccessfulConfig,
			nil,
			true,
			TestError,
		},
		{
			"Max result count exceeded error",
			createMockLoader(),
			[]string{TestVDName, TestVDNameSharedName},
			bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			nil,
			true,
			errors.ErrLimitExceeded{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetValueDescriptorsNameExecutor(test.vdNames, test.loader, logger.MockLogger{}, test.config)
			observed, err := op.Execute()

			if test.expectError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectError && test.expectedErrorType != nil {
				eet := reflect.TypeOf(test.expectedErrorType)
				aet := reflect.TypeOf(err)
				if !aet.AssignableTo(eet) {
					t.Errorf("Expected error of type %v, but got an error of type %v", eet, aet)
				}
			}

			if !reflect.DeepEqual(test.expectedResult, observed) {
				t.Errorf("Observed result doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedResult, observed)
			}
		})
	}
}

func TestGetAllValueDescriptors(t *testing.T) {
	tests := []struct {
		name              string
		loader            Loader
		config            bootstrapConfig.ServiceInfo
		expectedResult    []contract.ValueDescriptor
		expectError       bool
		expectedErrorType error
	}{
		{
			"One matching result",
			createMockLoaderSingleResult(),
			TestSuccessfulConfig,
			[]contract.ValueDescriptor{TestValueDescriptor},
			false,
			nil,
		},
		{
			"Multiple matching results",
			createMockLoader(),
			TestSuccessfulConfig,
			[]contract.ValueDescriptor{TestValueDescriptor, TestValueDescriptor2, TestValueDescriptor3},
			false,
			nil,
		},
		{
			"Database error",
			createErrorMockLoader(),
			TestSuccessfulConfig,
			nil,
			true,
			TestError,
		},
		{
			"Max result count exceeded error",
			createMockLoader(),
			bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			nil,
			true,
			errors.ErrLimitExceeded{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetValueDescriptorsExecutor(test.loader, logger.MockLogger{}, test.config)
			observed, err := op.Execute()

			if test.expectError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectError && test.expectedErrorType != nil {
				eet := reflect.TypeOf(test.expectedErrorType)
				aet := reflect.TypeOf(err)
				if !aet.AssignableTo(eet) {
					t.Errorf("Expected error of type %v, but got an error of type %v", eet, aet)
				}
			}

			if !reflect.DeepEqual(test.expectedResult, observed) {
				t.Errorf("Observed result doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedResult, observed)
			}
		})
	}
}

func createMockLoader() Loader {
	mock := &mocks.Loader{}
	mock.On("ValueDescriptors").Return([]contract.ValueDescriptor{TestValueDescriptor, TestValueDescriptor2, TestValueDescriptor3}, nil)
	mock.On("ValueDescriptorsByName", []string{TestVDName}).Return([]contract.ValueDescriptor{TestValueDescriptor}, nil)
	mock.On("ValueDescriptorsByName", []string{TestVDNameSharedName}).Return([]contract.ValueDescriptor{TestValueDescriptor2, TestValueDescriptor3}, nil)
	mock.On("ValueDescriptorsByName", []string{TestVDName, TestVDNameSharedName}).Return([]contract.ValueDescriptor{TestValueDescriptor, TestValueDescriptor2, TestValueDescriptor3}, nil)
	mock.On("ValueDescriptorsByName", TestVDErrorNames).Return(nil, TestError)

	return mock
}

func createMockLoaderSingleResult() Loader {
	mock := &mocks.Loader{}
	mock.On("ValueDescriptors").Return([]contract.ValueDescriptor{TestValueDescriptor}, nil)

	return mock
}

func createErrorMockLoader() Loader {
	mock := &mocks.Loader{}
	mock.On("ValueDescriptors").Return(nil, TestError)

	return mock
}
