package metadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
)

var TestError = errors.New("test error")
var TestDeviceProfileID = "TestProfileID"
var TestDeviceProfileName = "TestProfileName"
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
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithBody(TestDeviceProfile),
			createDBClient(),
			http.StatusNoContent,
		},
		{
			"Multiple devices associated with device profile",
			createRequestWithBody(TestDeviceProfile),
			createDBClientMultipleDevicesFoundError(),
			http.StatusConflict,
		},
		{
			"Multiple provision watchers associated with device profile",
			createRequestWithBody(TestDeviceProfile),
			createDBClientMultipleProvisionWatchersFoundError(),
			http.StatusConflict,
		},
		{
			"Device profile duplicate name",
			createRequestWithBody(TestDeviceProfile),
			createDBClientDuplicateDeviceProfileNameError(),
			http.StatusConflict,
		},
		{
			"GetAllDeviceProfilesError database error ",
			createRequestWithBody(TestDeviceProfile),
			createDBClientGetAllDeviceProfilesError(),
			http.StatusInternalServerError,
		},
		{
			"Invalid request body",
			createRequestWithInvalidBody(),
			createDBClient(), http.StatusBadRequest,
		},
		{
			"Device Profile Not Found",
			createRequestWithBody(TestDeviceProfile),
			createDBClientDeviceProfileNotFoundError(),
			http.StatusNotFound,
		},
		{
			"Multiple devices associated with device profile",
			createRequestWithBody(TestDeviceProfile),
			createDBClientMultipleDevicesFoundError(),
			http.StatusConflict,
		},
		{
			"Multiple provision watchers associated with device profile",
			createRequestWithBody(TestDeviceProfile),
			createDBClientMultipleProvisionWatchersFoundError(),
			http.StatusConflict,
		},
		{
			"Device profile duplicate name",
			createRequestWithBody(TestDeviceProfile),
			createDBClientDuplicateDeviceProfileNameError(),
			http.StatusConflict,
		},
		{
			"GetAllDeviceProfiles database error ",
			createRequestWithBody(TestDeviceProfile),
			createDBClientGetAllDeviceProfilesError(),
			http.StatusInternalServerError,
		},
		{
			"GetProvisionWatchersByProfileId database error ",
			createRequestWithBody(TestDeviceProfile),
			createDBClientGetProvisionWatchersByProfileIdError(),
			http.StatusInternalServerError,
		},
		{
			"UpdateDeviceProfile database error ",
			createRequestWithBody(TestDeviceProfile),
			createDBClientUpdateDeviceProfileError(),
			http.StatusInternalServerError,
		},
		{
			"GetDevicesByProfileId database error",
			createRequestWithBody(TestDeviceProfile),
			createDBClientGetDevicesByProfileIdError(),
			http.StatusInternalServerError,
		},
		{
			"Duplicate commands error ",
			createRequestWithBody(createTestDeviceProfileWithCommands(TestCommand, TestCommand)),
			createDBClient(),
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restUpdateDeviceProfile)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func createRequestWithBody(d contract.DeviceProfile) *http.Request {
	body, err := d.MarshalJSON()
	if err != nil {
		panic("Failed to create test JSON:" + err.Error())
	}

	return httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBuffer(body))
}

func createRequestWithInvalidBody() *http.Request {
	return httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBufferString("Bad JSON"))
}

func createDBClient() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("GetAllDeviceProfiles").Return(make([]contract.DeviceProfile, 0), nil)
	d.On("UpdateDeviceProfile", TestDeviceProfile).Return(nil)

	return d
}

func createDBClientDeviceProfileNotFoundError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(contract.DeviceProfile{}, TestError)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, TestError)

	return d
}
func createDBClientMultipleDevicesFoundError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(TestDevices, nil)

	return d
}

func createDBClientMultipleProvisionWatchersFoundError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(TestProvisionWatchers, nil)

	return d
}
func createDBClientDuplicateDeviceProfileNameError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("GetAllDeviceProfiles").Return([]contract.DeviceProfile{{Name: TestDeviceProfile.Name, Id: "SomethingElse"}}, nil)
	d.On("UpdateDeviceProfile", TestDeviceProfile).Return(nil)

	return d
}

func createDBClientGetDevicesByProfileIdError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), TestError)

	return d
}
func createDBClientGetAllDeviceProfilesError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("GetAllDeviceProfiles").Return([]contract.DeviceProfile{}, TestError)

	return d
}

func createDBClientGetProvisionWatchersByProfileIdError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), TestError)
	d.On("GetAllDeviceProfiles").Return([]contract.DeviceProfile{}, TestError)

	return d
}

func createDBClientUpdateDeviceProfileError() interfaces.DBClient {
	d := &mocks.DBClient{}
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
	return createTestDeviceProfileWithCommands(TestCommand)
}

// createTestDeviceProfileWithCommands creates a device profile to be used during testing.
// This function handles some of the necessary creation nuances which need to take place for proper mocking and equality
// verifications.
func createTestDeviceProfileWithCommands(commands ...contract.Command) contract.DeviceProfile {
	return contract.DeviceProfile{
		Id:   TestDeviceProfileID,
		Name: TestDeviceProfileName,
		DescribedObject: contract.DescribedObject{
			Description: "Some test data",
			Timestamps: contract.Timestamps{
				Origin:   123,
				Created:  456,
				Modified: 789,
			},
		},
		Labels:       []string{"test", "data"},
		Manufacturer: "Test Manufacturer",
		Model:        "Test Model",
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
