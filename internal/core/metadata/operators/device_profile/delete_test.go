package device_profile

import (
	"reflect"
	"testing"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_profile/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

func TestDeleteProfileById(t *testing.T) {
	tests := []struct {
		name              string
		database          DeviceProfileDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createDeviceDeleter(),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Device profile not found",
			database:          createDeviceDeleterNotFound(),
			expectError:       true,
			expectedErrorType: errors.ErrDeviceProfileNotFound{},
		},
		{
			name:              "Delete error",
			database:          createDeviceDeleterDeleteError(),
			expectError:       true,
			expectedErrorType: TestError,
		},
		{
			name:              "Devices associated",
			database:          createDeviceDeleterWithDevicesAssociated(),
			expectError:       true,
			expectedErrorType: errors.ErrDeviceProfileInvalidState{},
		},
		{
			name:              "Provision watchers associated",
			database:          createDeviceDeleterWithProvisionWatchersAssociated(),
			expectError:       true,
			expectedErrorType: errors.ErrDeviceProfileInvalidState{},
		},
		{
			name:              "GetDeviceProfileError",
			database:          createDeviceDeleterGetDeviceProfileError(),
			expectError:       true,
			expectedErrorType: TestError,
		},

		{
			name:              "GetProvisionWatchersByProfileIdError",
			database:          createDeviceDeleterGetProvisionWatchersByProfileIdError(),
			expectError:       true,
			expectedErrorType: TestError,
		},

		{
			name:              "GetDevicesByProfileIdError",
			database:          createDeviceDeleterGetDevicesByProfileIdError(),
			expectError:       true,
			expectedErrorType: TestError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteByIDExecutor(test.database, TestDeviceProfile.Id)
			err := op.Execute()

			if test.expectError && err == nil {
				t.Error("We expected an error but did not get one")
			}

			if !test.expectError && err != nil {
				t.Errorf("We do not expected an error but got one. %s", err.Error())
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
func TestDeleteProfileByName(t *testing.T) {
	tests := []struct {
		name              string
		database          DeviceProfileDeleter
		expectError       bool
		expectedErrorType error
	}{
		{
			name:              "Successful Delete",
			database:          createDeviceDeleter(),
			expectError:       false,
			expectedErrorType: nil,
		},
		{
			name:              "Device profile not found",
			database:          createDeviceDeleterNotFound(),
			expectError:       true,
			expectedErrorType: errors.ErrDeviceProfileNotFound{},
		},
		{
			name:              "Delete error",
			database:          createDeviceDeleterDeleteError(),
			expectError:       true,
			expectedErrorType: TestError,
		},
		{
			name:              "Devices associated",
			database:          createDeviceDeleterWithDevicesAssociated(),
			expectError:       true,
			expectedErrorType: errors.ErrDeviceProfileInvalidState{},
		},
		{
			name:              "Provision watchers associated",
			database:          createDeviceDeleterWithProvisionWatchersAssociated(),
			expectError:       true,
			expectedErrorType: errors.ErrDeviceProfileInvalidState{},
		},
		{
			name:              "GetDeviceProfileError",
			database:          createDeviceDeleterGetDeviceProfileError(),
			expectError:       true,
			expectedErrorType: TestError,
		},

		{
			name:              "GetProvisionWatchersByProfileIdError",
			database:          createDeviceDeleterGetProvisionWatchersByProfileIdError(),
			expectError:       true,
			expectedErrorType: TestError,
		},

		{
			name:              "GetDevicesByProfileIdError",
			database:          createDeviceDeleterGetDevicesByProfileIdError(),
			expectError:       true,
			expectedErrorType: TestError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := NewDeleteByNameExecutor(test.database, TestDeviceProfile.Name)
			err := op.Execute()

			if test.expectError && err == nil {
				t.Error("We expected an error but did not get one")
			}

			if !test.expectError && err != nil {
				t.Errorf("We do not expected an error but got one. %s", err.Error())
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

func createDeviceDeleter() DeviceProfileDeleter {
	d := mocks.DeviceProfileDeleter{}
	d.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfile.Id).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfile.Id).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("DeleteDeviceProfileById", TestDeviceProfile.Id).Return(nil)

	return &d
}

func createDeviceDeleterNotFound() DeviceProfileDeleter {
	d := mocks.DeviceProfileDeleter{}
	d.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(contract.DeviceProfile{}, db.ErrNotFound)
	d.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(contract.DeviceProfile{}, db.ErrNotFound)

	return &d
}

func createDeviceDeleterGetDeviceProfileError() DeviceProfileDeleter {
	d := mocks.DeviceProfileDeleter{}
	d.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(contract.DeviceProfile{}, TestError)
	d.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(contract.DeviceProfile{}, TestError)

	return &d
}

func createDeviceDeleterWithDevicesAssociated() DeviceProfileDeleter {
	d := mocks.DeviceProfileDeleter{}
	d.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfile.Id).Return(TestDevices, nil)

	return &d
}
func createDeviceDeleterGetDevicesByProfileIdError() DeviceProfileDeleter {
	d := mocks.DeviceProfileDeleter{}
	d.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfile.Id).Return(make([]contract.Device, 0), TestError)

	return &d
}
func createDeviceDeleterWithProvisionWatchersAssociated() DeviceProfileDeleter {
	d := mocks.DeviceProfileDeleter{}
	d.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfile.Id).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfile.Id).Return(TestProvisionWatchers, nil)

	return &d
}
func createDeviceDeleterGetProvisionWatchersByProfileIdError() DeviceProfileDeleter {
	d := mocks.DeviceProfileDeleter{}
	d.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfile.Id).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfile.Id).Return(make([]contract.ProvisionWatcher, 0), TestError)

	return &d
}
func createDeviceDeleterDeleteError() DeviceProfileDeleter {
	d := mocks.DeviceProfileDeleter{}
	d.On("GetDeviceProfileById", TestDeviceProfile.Id).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfile.Name).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfile.Id).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfile.Id).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("DeleteDeviceProfileById", TestDeviceProfile.Id).Return(TestError)

	return &d
}
