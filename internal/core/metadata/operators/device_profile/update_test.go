package device_profile

import (
	"encoding/json"
	"errors"
	"testing"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_profile/mocks"
)

var TestError = errors.New("test error")
var TestDeviceProfileID = "TestProfileID"
var TestDeviceProfileName = "TestProfileName"
var TestDeviceProfileLabel1 = "TestLabel1"
var TestDeviceProfileLabel2 = "TestLabel2"
var TestDeviceProfileLabels = []string{TestDeviceProfileLabel1, TestDeviceProfileLabel2}
var TestDeviceProfileManufacturer = "TestManufacturer"
var TestDeviceProfileModel = "TestModel"
var TestDeviceProfile = createTestDeviceProfile()
var TestCommand = contract.Command{Name: "TestCommand", Id: "TestCommandId"}
var TestDevices = []contract.Device{
	{
		Name: "TestDevice1",
	},
	{
		Name: "TestDevice2",
	},
}

var TestProvisionWatchers = []contract.ProvisionWatcher{
	{
		Name: "TestProvisionWatcher1",
	},
	{
		Name: "TestProvisionWatcher2",
	},
}

func TestUpdateDeviceProfile(t *testing.T) {
	tests := []struct {
		name        string
		dbMock      DeviceProfileUpdater
		dp          contract.DeviceProfile
		expectError bool
	}{
		{
			"Multiple devices associated with device profile",
			createDBClientMultipleDevicesFoundError(),
			TestDeviceProfile,
			true,
		},
		{
			"Multiple provision watchers associated with device profile",
			createDBClientMultipleProvisionWatchersFoundError(),
			TestDeviceProfile,
			true,
		},
		{
			"Device Profile Not Found",
			createDBClientDeviceProfileNotFoundError(),
			TestDeviceProfile,
			true,
		},
		{
			"Multiple devices associated with device profile",
			createDBClientMultipleDevicesFoundError(),
			TestDeviceProfile,
			true,
		},
		{
			"Multiple provision watchers associated with device profile",
			createDBClientMultipleProvisionWatchersFoundError(),
			TestDeviceProfile,
			true,
		},
		{
			"GetProvisionWatchersByProfileId db error ",
			createDBClientGetProvisionWatchersByProfileIdError(),
			TestDeviceProfile,
			true,
		},
		{
			"UpdateDeviceProfile db error ",
			createDBClientUpdateDeviceProfileError(),
			TestDeviceProfile,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewUpdateDeviceProfileExecutor(tt.dbMock, tt.dp)
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

func createDBClient() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("GetAllDeviceProfiles").Return(make([]contract.DeviceProfile, 0), nil)
	d.On("UpdateDeviceProfile", TestDeviceProfile).Return(nil)

	return d
}

func createDBClientDeviceProfileNotFoundError() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(contract.DeviceProfile{}, TestError)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, TestError)

	return d
}
func createDBClientMultipleDevicesFoundError() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(TestDevices, nil)

	return d
}

func createDBClientMultipleProvisionWatchersFoundError() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(TestProvisionWatchers, nil)

	return d
}
func createDBClientDuplicateDeviceProfileNameError() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("GetAllDeviceProfiles").Return([]contract.DeviceProfile{{Name: TestDeviceProfile.Name, Id: "SomethingElse"}}, nil)
	d.On("UpdateDeviceProfile", TestDeviceProfile).Return(nil)

	return d
}

func createDBClientGetDevicesByProfileIdError() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), TestError)

	return d
}
func createDBClientGetAllDeviceProfilesError() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("GetAllDeviceProfiles").Return([]contract.DeviceProfile{}, TestError)
	d.On("UpdateDeviceProfile", TestDeviceProfile).Return(nil)

	return d
}

func createDBClientGetProvisionWatchersByProfileIdError() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), TestError)
	d.On("GetAllDeviceProfiles").Return([]contract.DeviceProfile{}, TestError)

	return d
}

func createDBClientUpdateDeviceProfileError() DeviceProfileUpdater {
	d := &mocks.DeviceProfileUpdater{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("GetAllDeviceProfiles").Return(make([]contract.DeviceProfile, 0), nil)
	d.On("UpdateDeviceProfile", TestDeviceProfile).Return(TestError)

	return d
}

// createTestDeviceProfile creates a device profile to be used during testing.
// This function handles some of the necessary creation nuances which need to take place for proper mocking and equality
// verifications.
func createTestDeviceProfile() contract.DeviceProfile {
	return createTestDeviceProfileWithCommands(TestDeviceProfileID, TestDeviceProfileName, TestDeviceProfileLabels, TestDeviceProfileManufacturer, TestDeviceProfileModel, TestCommand)
}

// createTestDeviceProfileWithCommands creates a device profile to be used during testing.
// This function handles some of the necessary creation nuances which need to take place for proper mocking and equality
// verifications.
func createTestDeviceProfileWithCommands(id string, name string, labels []string, manufacturer string, model string, commands ...contract.Command) contract.DeviceProfile {
	return contract.DeviceProfile{
		Id:   id,
		Name: name,
		DescribedObject: contract.DescribedObject{
			Description: "Some test data",
			Timestamps: contract.Timestamps{
				Origin:   123,
				Created:  456,
				Modified: 789,
			},
		},
		Labels:       labels,
		Manufacturer: manufacturer,
		Model:        model,
		CoreCommands: createCoreCommands(commands),
		DeviceResources: []contract.DeviceResource{
			{
				Name: "TestDeviceResource",
			},
		},
		DeviceCommands: []contract.ProfileResource{
			{
				Name: "TestProfileResource",
			},
		},
	}
}

// createCoreCommands creates Command instances which can be used during testing.
// This function is necessary due to the internal field 'isValidated', which is not exported, being false when created
// manually and true when serialized. This causes the mocking infrastructure to not match when Commands are involved
// with matching parameters or verifying results.
func createCoreCommands(commands []contract.Command) []contract.Command {
	cs := make([]contract.Command, 0)
	for _, command := range commands {
		b, _ := command.MarshalJSON()
		var temp contract.Command
		err := json.Unmarshal(b, &temp)
		if err != nil {
			panic(err.Error())
		}

		cs = append(cs, temp)
	}

	return cs
}
