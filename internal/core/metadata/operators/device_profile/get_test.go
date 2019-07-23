package device_profile

import (
	"reflect"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_profile/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
)

var TestLabelError1 = "TestErrorLabel1"
var TestLabelError2 = "TestErrorLabel2"

var TestDeviceProfileError = createTestDeviceProfileWithCommands("TestErrorID", "TestErrorName", []string{TestLabelError1, TestLabelError2}, "TestErrorManufacturer", "TestErrorModel", TestCommand)
var TestDeviceProfiles = []contract.DeviceProfile{
	TestDeviceProfile,
	createTestDeviceProfileWithCommands("TestDeviceProfileID2", "TestDeviceProfileName2", []string{TestDeviceProfileLabel1, TestDeviceProfileLabel2}, TestDeviceProfileManufacturer, TestDeviceProfileModel, TestCommand),
	createTestDeviceProfileWithCommands("TestErrorID", "TestErrorName", []string{TestLabelError1, TestLabelError2}, "TestErrorManufacturer", "TestErrorModel", TestCommand),
}

func TestGetAllExecutor(t *testing.T) {
	tests := []struct {
		name           string
		dl             DeviceProfileLoader
		maxResultCount int
		expectedResult []contract.DeviceProfile
		expectError    bool
	}{
		{
			"Successfully get all device profiles",
			createDeviceProfileLoaderMock(),
			len(TestDeviceProfiles),
			TestDeviceProfiles,
			false,
		},
		{
			"Database error",
			createDeviceProfileLoaderMockGetAllError(),
			len(TestDeviceProfiles),
			TestDeviceProfiles,
			true,
		},
		{
			"Max limit exceeded error",
			createDeviceProfileLoaderMock(),
			len(TestDeviceProfiles) - 1,
			TestDeviceProfiles,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetAllExecutor(config.ServiceInfo{MaxResultCount: test.maxResultCount}, test.dl, logger.MockLogger{})
			actual, err := op.Execute()
			if err != nil && !test.expectError {
				t.Error(err)
				return
			}

			if err == nil && test.expectError {
				t.Errorf("error was expected, none occurred")
				return
			}

			if !test.expectError && !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}
func TestGetModelExecutor(t *testing.T) {
	tests := []struct {
		name           string
		dl             DeviceProfileLoader
		mod            string
		expectedResult []contract.DeviceProfile
		expectError    bool
	}{
		{
			"Successfully get all device profiles",
			createDeviceProfileLoaderMock(),
			TestDeviceProfileModel,
			TestDeviceProfiles,
			false,
		},
		{
			"Database error",
			createDeviceProfileLoaderMock(),
			TestDeviceProfileError.Model,
			nil,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetModelExecutor(test.mod, test.dl)
			actual, err := op.Execute()
			if err != nil && !test.expectError {
				t.Error(err)
				return
			}

			if err == nil && test.expectError {
				t.Errorf("error was expected, none occurred")
				return
			}

			if !test.expectError && !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestGetLabelExecutor(t *testing.T) {
	tests := []struct {
		name           string
		dl             DeviceProfileLoader
		l              string
		expectedResult []contract.DeviceProfile
		expectError    bool
	}{
		{
			"Successfully get all device profiles",
			createDeviceProfileLoaderMock(),
			TestDeviceProfileLabel1,
			TestDeviceProfiles,
			false,
		},
		{
			"Database error",
			createDeviceProfileLoaderMock(),
			TestLabelError1,
			nil,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetLabelExecutor(test.l, test.dl)
			actual, err := op.Execute()
			if err != nil && !test.expectError {
				t.Error(err)
				return
			}

			if err == nil && test.expectError {
				t.Errorf("error was expected, none occurred")
				return
			}

			if !test.expectError && !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestGetManufacturerModelExecutor(t *testing.T) {
	tests := []struct {
		name           string
		dl             DeviceProfileLoader
		man            string
		mod            string
		expectedResult []contract.DeviceProfile
		expectError    bool
	}{
		{
			"Successfully get all device profiles",
			createDeviceProfileLoaderMock(),
			TestDeviceProfile.Manufacturer,
			TestDeviceProfile.Model,
			TestDeviceProfiles,
			false,
		},
		{
			"Database error",
			createDeviceProfileLoaderMock(),
			TestDeviceProfileError.Manufacturer,
			TestDeviceProfileError.Model,
			nil,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetManufacturerModelExecutor(test.man, test.mod, test.dl)
			actual, err := op.Execute()
			if err != nil && !test.expectError {
				t.Error(err)
				return
			}

			if err == nil && test.expectError {
				t.Errorf("error was expected, none occurred")
				return
			}

			if !test.expectError && !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestGetManufacturerExecutor(t *testing.T) {
	tests := []struct {
		name           string
		dl             DeviceProfileLoader
		man            string
		expectedResult []contract.DeviceProfile
		expectError    bool
	}{
		{
			"Successfully get all device profiles",
			createDeviceProfileLoaderMock(),
			TestDeviceProfile.Manufacturer,
			TestDeviceProfiles,
			false,
		},
		{
			"Database error",
			createDeviceProfileLoaderMock(),
			TestDeviceProfileError.Manufacturer,
			nil,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetManufacturerExecutor(test.man, test.dl)
			actual, err := op.Execute()
			if err != nil && !test.expectError {
				t.Error(err)
				return
			}

			if err == nil && test.expectError {
				t.Errorf("error was expected, none occurred")
				return
			}

			if !test.expectError && !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestGetIdExecutor(t *testing.T) {
	tests := []struct {
		name           string
		dl             DeviceProfileLoader
		id             string
		expectedResult contract.DeviceProfile
		expectError    bool
	}{
		{
			"Successfully get all device profiles",
			createDeviceProfileLoaderMock(),
			TestDeviceProfile.Id,
			TestDeviceProfile,
			false,
		},
		{
			"Database error",
			createDeviceProfileLoaderMock(),
			TestDeviceProfileError.Id,
			contract.DeviceProfile{},
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetProfileID(test.id, test.dl)
			actual, err := op.Execute()
			if err != nil && !test.expectError {
				t.Error(err)
				return
			}

			if err == nil && test.expectError {
				t.Errorf("error was expected, none occurred")
				return
			}

			if !test.expectError && !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func TestGetNameExecutor(t *testing.T) {
	tests := []struct {
		name           string
		dl             DeviceProfileLoader
		profileName    string
		expectedResult contract.DeviceProfile
		expectError    bool
	}{
		{
			"Successfully get all device profiles",
			createDeviceProfileLoaderMock(),
			TestDeviceProfile.Name,
			TestDeviceProfile,
			false,
		},
		{
			"Database error",
			createDeviceProfileLoaderMock(),
			TestDeviceProfileError.Name,
			contract.DeviceProfile{},
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewGetProfileName(test.profileName, test.dl)
			actual, err := op.Execute()
			if err != nil && !test.expectError {
				t.Error(err)
				return
			}

			if err == nil && test.expectError {
				t.Errorf("error was expected, none occurred")
				return
			}

			if !test.expectError && !reflect.DeepEqual(test.expectedResult, actual) {
				t.Errorf("Expected result does not match the observed.\nExpected: %v\nObserved: %v\n", test.expectedResult, actual)
				return
			}
		})
	}
}

func createDeviceProfileLoaderMock() DeviceProfileLoader {
	mock := &mocks.DeviceProfileLoader{}

	// Successful mock calls
	mock.On("GetAllDeviceProfiles").Return(TestDeviceProfiles, nil)
	mock.On("GetDeviceProfilesByModel", TestDeviceProfile.Model).Return(TestDeviceProfiles, nil)
	mock.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel1).Return(TestDeviceProfiles, nil)
	mock.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel2).Return(TestDeviceProfiles, nil)
	mock.On("GetDeviceProfilesByManufacturer", TestDeviceProfile.Manufacturer).Return(TestDeviceProfiles, nil)
	mock.On("GetDeviceProfilesByManufacturerModel", TestDeviceProfile.Manufacturer, TestDeviceProfile.Model).Return(TestDeviceProfiles, nil)
	mock.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(TestDeviceProfile, nil)
	mock.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(TestDeviceProfile, nil)

	// Mock calls to simulate errors
	mock.On("GetDeviceProfilesByModel", TestDeviceProfileError.Model).Return(make([]contract.DeviceProfile, 0), TestError)
	mock.On("GetDeviceProfilesWithLabel", TestLabelError1).Return(make([]contract.DeviceProfile, 0), TestError)
	mock.On("GetDeviceProfilesWithLabel", TestLabelError2).Return(make([]contract.DeviceProfile, 0), TestError)
	mock.On("GetDeviceProfilesByManufacturer", TestDeviceProfileError.Manufacturer).Return(make([]contract.DeviceProfile, 0), TestError)
	mock.On("GetDeviceProfilesByManufacturerModel", TestDeviceProfileError.Manufacturer, TestDeviceProfileError.Model).Return(make([]contract.DeviceProfile, 0), TestError)
	mock.On("GetDeviceProfileById", TestDeviceProfileError.Id).Return(contract.DeviceProfile{}, TestError)
	mock.On("GetDeviceProfileByName", TestDeviceProfileError.Name).Return(contract.DeviceProfile{}, TestError)

	return mock
}

func createDeviceProfileLoaderMockGetAllError() DeviceProfileLoader {
	mock := &mocks.DeviceProfileLoader{}

	// Mock calls to simulate errors
	mock.On("GetAllDeviceProfiles").Return(make([]contract.DeviceProfile, 0), TestError)

	return mock
}
