package device_profile

import (
	"context"
	"reflect"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	dataErrors "github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	mocks2 "github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_profile/mocks"
)

var TestDeviceResource1 = contract.DeviceResource{
	Name:        "TestDeviceResource1Name",
	Description: "TestDeviceResource1Description",
	Properties: contract.ProfileProperty{
		Value: contract.PropertyValue{
			Maximum:       "TestValueMax",
			Minimum:       "TestValueMin",
			DefaultValue:  "TestValueDefault",
			Type:          "TestValueType",
			FloatEncoding: "TestValueFloatEncoding",
			MediaType:     "TestValueMediaType",
		},
		Units: contract.Units{
			Type:         "TestUnitType",
			DefaultValue: "TestDefaultUnitValue",
			ReadWrite:    "TestReadWrite",
		},
	},
}

var TestDeviceResource2 = contract.DeviceResource{
	Name:        "TestDeviceResource2Name",
	Description: "TestDeviceResource2Description",
	Properties: contract.ProfileProperty{
		Value: contract.PropertyValue{
			Maximum:       "TestValueMax",
			Minimum:       "TestValueMin",
			DefaultValue:  "TestValueDefault",
			Type:          "TestValueType",
			FloatEncoding: "TestValueFloatEncoding",
			MediaType:     "TestValueMediaType",
		},
		Units: contract.Units{
			Type:         "TestUnitType",
			DefaultValue: "TestDefaultUnitValue",
			ReadWrite:    "TestReadWrite",
		},
	},
}

var TestDeviceResourceError = contract.DeviceResource{
	Name:        "TestDeviceResourceError",
	Description: "TestDeviceResource1Description",
	Properties: contract.ProfileProperty{
		Value: contract.PropertyValue{
			Maximum:       "TestValueMax",
			Minimum:       "TestValueMin",
			DefaultValue:  "TestValueDefault",
			Type:          "TestValueType",
			FloatEncoding: "TestValueFloatEncoding",
			MediaType:     "TestValueMediaType",
		},
		Units: contract.Units{
			Type:         "TestUnitType",
			DefaultValue: "TestDefaultUnitValue",
			ReadWrite:    "TestReadWrite",
		},
	},
}

var TestValueDescriptor1 = contract.From(TestDeviceResource1)
var TestValueDescriptor2 = contract.From(TestDeviceResource2)
var TestValueDescriptorError = contract.From(TestDeviceResourceError)
var TestDeviceResources = []contract.DeviceResource{
	TestDeviceResource1,
	TestDeviceResource2,
}

var TestContext = context.Background()

var TestUpdatedDeviceProfile = contract.DeviceProfile{
	DescribedObject: contract.DescribedObject{},
	Id:              "TestDeivceProfile",
	Name:            "TestDeviceProfileName",
	DeviceResources: []contract.DeviceResource{TestCreateDeviceResource, TestUpdateDeviceResource},
}

var TestExistingDeviceProfile = contract.DeviceProfile{
	Id:              TestUpdatedDeviceProfile.Id,
	Name:            TestUpdatedDeviceProfile.Name,
	DeviceResources: []contract.DeviceResource{TestDeleteDeviceResource, TestUpdateDeviceResource},
}

var TestDeleteDeviceResource = contract.DeviceResource{Name: "TestDeleteDeviceResource"}
var TestUpdateDeviceResource = contract.DeviceResource{Name: "TestUpdateDeviceResource"}
var TestCreateDeviceResource = contract.DeviceResource{Name: "TestCreateDeviceResource"}
var TestDeleteValueDescriptor = contract.From(TestDeleteDeviceResource)
var TestUpdateValueDescriptor = contract.From(TestUpdateDeviceResource)
var TestCreateValueDescriptor = contract.From(TestCreateDeviceResource)
var TestInUseDeviceResource = contract.DeviceResource{Name: "TestInUseDeviceResource"}

func TestAddValueDescriptors(t *testing.T) {
	tests := []struct {
		name              string
		dr                []contract.DeviceResource
		expectError       bool
		expectedErrorType error
	}{
		{
			"Add 1 ValueDescriptor successfully",
			[]contract.DeviceResource{TestDeviceResource1},
			false,
			nil,
		},
		{
			"Add multiple ValueDescriptor successfully",
			TestDeviceResources,
			false,
			nil,
		},
		{
			"Add 1 ValueDescriptor with client error",
			[]contract.DeviceResource{TestDeviceResourceError},
			true,
			TestError,
		},
		{
			"Add multiple ValueDescriptor with client error",
			[]contract.DeviceResource{TestDeviceResource1, TestDeviceResource2, TestDeviceResourceError},
			true,
			TestError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewAddValueDescriptorExecutor(TestContext, createMockValueDescriptorAdder(), logger.MockLogger{}, test.dr...)
			err := op.Execute()

			if test.expectError && err == nil {
				t.Error("We expected an error but did not get one")
			}

			if !test.expectError && err != nil {
				t.Errorf("We do not expect an error but got one. %s", err.Error())
			}

			if test.expectError {
				eet := reflect.TypeOf(test.expectedErrorType)
				aet := reflect.TypeOf(err)
				if !aet.AssignableTo(eet) {
					t.Errorf("Expected error of type %v, but got an error of type %v", eet, aet)
				}
			}

			return
		})
	}
}

func TestUpdateValueDescriptors(t *testing.T) {
	tests := []struct {
		name              string
		oldDP             contract.DeviceProfile
		newDP             contract.DeviceProfile
		loader            DeviceProfileUpdater
		client            ValueDescriptorUpdater
		expectError       bool
		expectedErrorType error
	}{
		{
			"Successfully update",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockDBClient(),
			createMockValueDescriptorUpdater(),
			false,
			nil,
		},
		{
			"DeviceProfile in use",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockDBClientDeviceProfileInUse(),
			createMockValueDescriptorUpdater(),
			true,
			errors.ErrDeviceProfileInvalidState{},
		},
		{
			"ValueDescriptor in use",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockDBClient(),
			createMockValueDescriptorUpdaterInUseError(),
			true,
			dataErrors.ErrValueDescriptorsInUse{},
		},
		{
			"ValueDescriptorUsage error",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockDBClient(),
			createMockValueDescriptorUsageError(),
			true,
			TestError,
		},
		{
			"GetDeviceProfileByName error",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockErrorDBClient(),
			createMockValueDescriptorUpdater(),
			true,
			TestError,
		},
		{
			"ValueDescriptor add error",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockDBClient(),
			createMockValueDescriptorClientError(TestError, nil, nil, nil),
			true,
			TestError,
		},
		{
			"ValueDescriptor get error",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockDBClient(),
			createMockValueDescriptorClientError(nil, TestError, nil, nil),
			true,
			TestError,
		},
		{
			"ValueDescriptor update error",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockDBClient(),
			createMockValueDescriptorClientError(nil, nil, TestError, nil),
			true,
			TestError,
		},
		{
			"ValueDescriptor delete error",
			TestExistingDeviceProfile,
			TestUpdatedDeviceProfile,
			createMockDBClient(),
			createMockValueDescriptorClientError(nil, nil, nil, TestError),
			true,
			TestError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewUpdateValueDescriptorExecutor(TestContext, test.newDP, test.loader, test.client, logger.MockLogger{})
			err := op.Execute()

			if test.expectError && err == nil {
				t.Error("We expected an error but did not get one")
			}

			if !test.expectError && err != nil {
				t.Errorf("We do not expect an error but got one. %s", err.Error())
			}

			if test.expectError {
				eet := reflect.TypeOf(test.expectedErrorType)
				aet := reflect.TypeOf(err)
				if !aet.AssignableTo(eet) {
					t.Errorf("Expected error of type %v, but got an error of type %v", eet, aet)
				}
			}

			return
		})
	}
}

func createMockValueDescriptorAdder() ValueDescriptorAdder {
	mockClient := mocks.ValueDescriptorAdder{}
	mockClient.On("Add", &TestValueDescriptor1, TestContext).Return("1", nil)
	mockClient.On("Add", &TestValueDescriptor2, TestContext).Return("2", nil)
	mockClient.On("Add", &TestValueDescriptorError, TestContext).Return("", TestError)

	return &mockClient
}

func createMockValueDescriptorUpdater() ValueDescriptorUpdater {
	mockUpdater := &mocks.ValueDescriptorUpdater{}
	mockUpdater.On("ValueDescriptorsUsage", []string{TestDeleteDeviceResource.Name, TestUpdateDeviceResource.Name}, TestContext).Return(map[string]bool{TestUpdateDeviceResource.Name: false}, nil)
	mockUpdater.On("Add", &TestCreateValueDescriptor, TestContext).Return(TestCreateDeviceResource.Name, nil)
	mockUpdater.On("ValueDescriptorForName", TestUpdateDeviceResource.Name, TestContext).Return(TestUpdateValueDescriptor, nil)
	mockUpdater.On("Update", &TestUpdateValueDescriptor, TestContext).Return(nil)
	mockUpdater.On("DeleteByName", TestDeleteDeviceResource.Name, TestContext).Return(nil)

	return mockUpdater
}

func createMockValueDescriptorUpdaterInUseError() ValueDescriptorUpdater {
	mockUpdater := &mocks.ValueDescriptorUpdater{}
	mockUpdater.On("ValueDescriptorsUsage", []string{TestDeleteDeviceResource.Name, TestUpdateDeviceResource.Name}, TestContext).Return(map[string]bool{TestUpdateDeviceResource.Name: true}, nil)

	return mockUpdater
}
func createMockValueDescriptorUsageError() ValueDescriptorUpdater {
	mockUpdater := &mocks.ValueDescriptorUpdater{}
	mockUpdater.On("ValueDescriptorsUsage", []string{TestDeleteDeviceResource.Name, TestUpdateDeviceResource.Name}, TestContext).Return(nil, TestError)

	return mockUpdater
}

func createMockValueDescriptorClientError(addError, getError, updateError, deleteError error) ValueDescriptorUpdater {
	mockUpdater := &mocks.ValueDescriptorUpdater{}
	mockUpdater.On("ValueDescriptorsUsage", []string{TestDeleteDeviceResource.Name, TestUpdateDeviceResource.Name}, TestContext).Return(map[string]bool{TestUpdateDeviceResource.Name: false}, nil)
	mockUpdater.On("Add", &TestCreateValueDescriptor, TestContext).Return(TestCreateDeviceResource.Name, addError)
	mockUpdater.On("ValueDescriptorForName", TestUpdateDeviceResource.Name, TestContext).Return(contract.ValueDescriptor{}, getError)
	mockUpdater.On("Update", &TestUpdateValueDescriptor, TestContext).Return(updateError)
	mockUpdater.On("DeleteByName", TestDeleteDeviceResource.Name, TestContext).Return(deleteError)

	return mockUpdater
}

func createMockDBClient() interfaces.DBClient {
	mockDb := &mocks2.DBClient{}
	mockDb.On("GetDeviceProfileByName", TestExistingDeviceProfile.Name).Return(TestExistingDeviceProfile, nil)
	mockDb.On("GetDevicesByProfileId", TestExistingDeviceProfile.Id).Return([]contract.Device{}, nil)

	return mockDb
}
func createMockDBClientDeviceProfileInUse() interfaces.DBClient {
	mockDb := &mocks2.DBClient{}
	mockDb.On("GetDeviceProfileByName", TestExistingDeviceProfile.Name).Return(TestExistingDeviceProfile, nil)
	mockDb.On("GetDevicesByProfileId", TestExistingDeviceProfile.Id).Return(TestDevices, nil)

	return mockDb
}

func createMockErrorDBClient() interfaces.DBClient {
	mockDb := &mocks2.DBClient{}
	mockDb.On("GetDeviceProfileByName", TestExistingDeviceProfile.Name).Return(contract.DeviceProfile{}, TestError)

	return mockDb
}
