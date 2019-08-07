package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"

	errors2 "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

var TestLabelError1 = "TestErrorLabel1"
var TestLabelError2 = "TestErrorLabel2"
var TestDeviceProfileLabel1 = "TestLabel1"
var TestDeviceProfileLabel2 = "TestLabel2"
var TestDeviceProfileLabels = []string{TestDeviceProfileLabel1, TestDeviceProfileLabel2}
var TestDeviceProfileManufacturer = "TestManufacturer"
var TestDeviceProfileModel = "TestModel"
var TestDeviceProfiles = []contract.DeviceProfile{
	TestDeviceProfile,
	createTestDeviceProfileWithCommands("TestDeviceProfileID2", "TestDeviceProfileName2", []string{TestDeviceProfileLabel1, TestDeviceProfileLabel2}, TestDeviceProfileManufacturer, TestDeviceProfileModel, TestCommand),
	createTestDeviceProfileWithCommands("TestErrorID", "TestErrorName", []string{TestLabelError1, TestLabelError2}, "TestErrorManufacturer", "TestErrorModel", TestCommand),
}
var TestError = errors.New("test error")
var TestContext = context.WithValue(context.Background(), "TestKey", "TestValue")
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

var TestDescriptionError = "Error"
var TestDeleteDeviceResource = contract.DeviceResource{Name: "TestDeleteDeviceResource"}
var TestUpdateDeviceResource = contract.DeviceResource{Name: "TestUpdateDeviceResource"}
var TestCreateDeviceResource = contract.DeviceResource{Name: "TestCreateDeviceResource"}
var TestDeleteValueDescriptor = contract.From(TestDeleteDeviceResource)
var TestUpdateValueDescriptor = contract.From(TestUpdateDeviceResource)
var TestCreateValueDescriptor = contract.From(TestCreateDeviceResource)
var TestInUseDeviceResource = contract.DeviceResource{Name: "TestInUseDeviceResource"}

func TestGetAllProfiles(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet,
				map[string]string{}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Max result count exceeded",
			createRequestWithPathParameters(http.MethodGet,
				map[string]string{}),
			createDBClientGetDeviceProfileMaxLimitError(),
			http.StatusRequestEntityTooLarge,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet,
				map[string]string{}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusInternalServerError,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{ID: TestDeviceProfileID}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: len(TestDeviceProfiles)}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAllDeviceProfiles)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetProfileById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{ID: TestDeviceProfileID}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet, map[string]string{ID: TestDeviceProfileID}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusNotFound,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{ID: TestDeviceProfileID}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetProfileByProfileId)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetYamlProfileById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{ID: TestDeviceProfileID}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet, map[string]string{ID: TestDeviceProfileID}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusNotFound,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{ID: TestDeviceProfileID}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetYamlProfileById)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetProfileByName(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Invalid name",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: ErrorPathParam}),
			createDBClient(),
			http.StatusBadRequest,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusNotFound,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetProfileByName)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetYamlProfileByName(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Invalid name",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: ErrorPathParam}),
			createDBClient(),
			http.StatusBadRequest,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusNotFound,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetYamlProfileByName)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetProfileByLabel(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{LABEL: TestDeviceProfileLabel1}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Invalid name",
			createRequestWithPathParameters(http.MethodGet, map[string]string{LABEL: ErrorPathParam}),
			createDBClient(),
			http.StatusBadRequest,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet, map[string]string{LABEL: TestDeviceProfileLabel1}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusInternalServerError,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{LABEL: TestDeviceProfileLabel1}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetProfileWithLabel)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetProfileByManufacturer(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{MANUFACTURER: TestDeviceProfileManufacturer}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Invalid name",
			createRequestWithPathParameters(http.MethodGet, map[string]string{MANUFACTURER: ErrorPathParam}),
			createDBClient(),
			http.StatusBadRequest,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet, map[string]string{MANUFACTURER: TestDeviceProfileManufacturer}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusInternalServerError,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{MANUFACTURER: TestDeviceProfileManufacturer}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetProfileByManufacturer)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetProfileByModel(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{MODEL: TestDeviceProfileModel}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Invalid MODEL",
			createRequestWithPathParameters(http.MethodGet, map[string]string{MODEL: ErrorPathParam}),
			createDBClient(),
			http.StatusBadRequest,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet, map[string]string{MODEL: TestDeviceProfileModel}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusInternalServerError,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{MODEL: TestDeviceProfileModel}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetProfileByModel)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestGetProfileByManufacturerModel(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet,
				map[string]string{
					MANUFACTURER: TestDeviceProfileManufacturer,
					MODEL:        TestDeviceProfileModel,
				}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Invalid MANUFACTURER",
			createRequestWithPathParameters(http.MethodGet,
				map[string]string{
					MANUFACTURER: ErrorPathParam,
					MODEL:        TestDeviceProfileModel,
				}),
			createDBClient(),
			http.StatusBadRequest,
		},
		{
			"Invalid MODEL",
			createRequestWithPathParameters(http.MethodGet,
				map[string]string{
					MANUFACTURER: TestDeviceProfileManufacturer,
					MODEL:        ErrorPathParam,
				}),
			createDBClient(),
			http.StatusBadRequest,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet,
				map[string]string{
					MANUFACTURER: TestDeviceProfileManufacturer,
					MODEL:        TestDeviceProfileModel,
				}), createDBClientDeviceProfileErrorNotFound(),
			http.StatusInternalServerError,
		},
		{
			"Database error",
			createRequestWithPathParameters(http.MethodGet,
				map[string]string{
					MANUFACTURER: TestDeviceProfileManufacturer,
					MODEL:        TestDeviceProfileModel,
				}),
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetProfileByManufacturerModel)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestUpdateDeviceProfile(t *testing.T) {
	tests := []struct {
		name                            string
		request                         *http.Request
		dbMock                          interfaces.DBClient
		enableValueDescriptorManagement bool
		expectedStatus                  int
	}{
		{
			"OK",
			createRequestWithBody(TestDeviceProfile),
			createDBClient(),
			true,
			http.StatusOK,
		},
		{
			"ValueDescriptor in use error",
			createRequestWithBody(TestDeviceProfile),
			createDBClientPersistDeviceInUseError(),
			true,
			http.StatusConflict,
		},
		{
			"ValueDescriptor update error",
			createRequestWithBody(TestDeviceProfile),
			createDBClientUpdateValueDescriptorError(),
			true,
			http.StatusInternalServerError,
		},
		{
			"Multiple devices associated with device profile",
			createRequestWithBody(TestDeviceProfile),
			createDBClientMultipleDevicesFoundError(),
			true,
			http.StatusConflict,
		},
		{
			"Multiple provision watchers associated with device profile",
			createRequestWithBody(TestDeviceProfile),
			createDBClientMultipleProvisionWatchersFoundError(),
			true,
			http.StatusConflict,
		},
		{
			"Invalid request body",
			createRequestWithInvalidBody(),
			createDBClient(),
			true,
			http.StatusBadRequest,
		},
		{
			"Device Profile Not Found(ValueDescriptorExecutor)",
			createRequestWithBody(TestDeviceProfile),
			createMockErrDeviceProfileNotFound(),
			true,
			http.StatusNotFound,
		},
		{
			"Device Profile Not Found(UpdateDeviceProfileExecutor)",
			createRequestWithBody(TestDeviceProfile),
			createMockErrDeviceProfileNotFound(),
			false,
			http.StatusNotFound,
		},
		{
			"GetProvisionWatchersByProfileId database error ",
			createRequestWithBody(TestDeviceProfile),
			createDBClientGetProvisionWatchersByProfileIdError(),
			true,
			http.StatusInternalServerError,
		},
		{
			"Service client error ",
			createRequestWithBody(TestDeviceProfile),
			createDBServiceClientError(http.StatusTeapot),
			true,
			http.StatusTeapot,
		},
		{
			"UpdateDeviceProfile database error ",
			createRequestWithBody(TestDeviceProfile),
			createDBClientPersistDeviceProfileError(),
			true,
			http.StatusInternalServerError,
		},
		{
			"GetDevicesByProfileId database error",
			createRequestWithBody(TestDeviceProfile),
			createDBClientGetDevicesByProfileIdError(),
			true,
			http.StatusInternalServerError,
		},
		{
			"Duplicate commands error ",
			createRequestWithBody(createTestDeviceProfileWithCommands(TestDeviceProfileID, TestDeviceProfileName, TestDeviceProfileLabels, TestDeviceProfileManufacturer, TestDeviceProfileModel, TestCommand, TestCommand)),
			createDBClient(),
			true,
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Writable: WritableInfo{EnableValueDescriptorManagement: tt.enableValueDescriptorManagement}, Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			vdc = MockValueDescriptorClient{}
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

func TestDeleteDeviceProfileById(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodDelete, map[string]string{ID: TestDeviceProfileID}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Multiple devices associated with device profile",
			createRequestWithPathParameters(http.MethodDelete, map[string]string{ID: TestDeviceProfileID}),
			createDBClientMultipleDevicesFoundError(),
			http.StatusConflict,
		},
		{
			"Multiple provision watchers associated with device profile",
			createRequestWithPathParameters(http.MethodDelete, map[string]string{ID: TestDeviceProfileID}),
			createDBClientMultipleProvisionWatchersFoundError(),
			http.StatusConflict,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodDelete, map[string]string{ID: TestDeviceProfileID}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusNotFound,
		},
		{
			"GetProvisionWatchersByProfileId database error ",
			createRequestWithPathParameters(http.MethodDelete, map[string]string{ID: TestDeviceProfileID}),
			createDBClientGetProvisionWatchersByProfileIdError(),
			http.StatusInternalServerError,
		},
		{
			"DeleteDeviceProfileById database error ",
			createRequestWithPathParameters(http.MethodDelete, map[string]string{ID: TestDeviceProfileID}),
			createDBClientPersistDeviceProfileError(),
			http.StatusInternalServerError,
		},
		{
			"GetDevicesByProfileId database error",
			createRequestWithPathParameters(http.MethodDelete, map[string]string{ID: TestDeviceProfileID}),
			createDBClientGetDevicesByProfileIdError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteProfileByProfileId)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestDeleteDeviceProfileByName(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClient(),
			http.StatusOK,
		},
		{
			"Invalid name",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: ErrorPathParam}),
			createDBClient(),
			http.StatusBadRequest,
		},
		{
			"Multiple devices associated with device profile",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientMultipleDevicesFoundError(),
			http.StatusConflict,
		},
		{
			"Multiple provision watchers associated with device profile",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientMultipleProvisionWatchersFoundError(),
			http.StatusConflict,
		},
		{
			"Device Profile Not Found",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusNotFound,
		},
		{
			"GetProvisionWatchersByProfileId database error ",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientGetProvisionWatchersByProfileIdError(),
			http.StatusInternalServerError,
		},
		{
			"DeleteDeviceProfileById database error ",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientPersistDeviceProfileError(),
			http.StatusInternalServerError,
		},
		{
			"GetDevicesByProfileId database error",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientGetDevicesByProfileIdError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restDeleteProfileByName)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestAddDeviceProfileByYaml(t *testing.T) {
	emptyName := TestDeviceProfile
	emptyName.Name = ""

	duplicateCommandName := TestDeviceProfile
	duplicateCommandName.CoreCommands = createCoreCommands([]contract.Command{TestCommand, TestCommand})

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", TestDeviceProfile, TestDeviceProfileID, nil},
			}),
			http.StatusOK,
		},
		{
			"Duplicate command name",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", duplicateCommandName, "", db.ErrNotUnique},
			}),
			http.StatusConflict,
		},
		{
			"Empty device profile name",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", emptyName, "", db.ErrNameEmpty},
			}),
			http.StatusBadRequest,
		},
		{
			"Unsuccessful database call",
			createRequestWithPathParameters(http.MethodGet, map[string]string{NAME: TestDeviceProfileName}),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", emptyName, "", TestError},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restAddProfileByYaml)
			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestAddDeviceProfileByYamlRaw(t *testing.T) {
	_, _ = TestDeviceProfile.Validate()
	
	okBody, _ := yaml.Marshal(TestDeviceProfile)

	duplicateCommandName := TestDeviceProfile
	duplicateCommandName.CoreCommands = createCoreCommands([]contract.Command{TestCommand, TestCommand})
	dupeBody, _ := yaml.Marshal(duplicateCommandName)

	emptyName := TestDeviceProfile
	emptyName.Name = ""
	emptyBody, _ := yaml.Marshal(emptyName)

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBuffer(okBody)),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", TestDeviceProfile, TestDeviceProfileID, nil},
			}),
			http.StatusOK,
		},
		{
			"Duplicate command name",
			httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBuffer(dupeBody)),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", duplicateCommandName, "", db.ErrNotUnique},
			}),
			http.StatusConflict,
		},
		{
			"Empty device profile name",
			httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBuffer(emptyBody)),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", emptyName, "", db.ErrNameEmpty},
			}),
			http.StatusBadRequest,
		},
		{
			"Unsuccessful database call",
			httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBuffer(okBody)),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", TestDeviceProfile, "", TestError},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LoggingClient = logger.MockLogger{}
			Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restAddProfileByYamlRaw)
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

func createRequestWithPathParameters(httpMethod string, params map[string]string) *http.Request {
	req := httptest.NewRequest(httpMethod, TestURI, nil)
	return mux.SetURLVars(req, params)
}

func createRequestWithInvalidBody() *http.Request {
	return httptest.NewRequest(http.MethodPut, TestURI, bytes.NewBufferString("Bad JSON"))
}

func createDBClient() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetAllDeviceProfiles").Return(TestDeviceProfiles, nil)
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfilesByModel", TestDeviceProfile.Model).Return(TestDeviceProfiles, nil)
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel1).Return(TestDeviceProfiles, nil)
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel2).Return(TestDeviceProfiles, nil)
	d.On("GetDeviceProfilesByManufacturer", TestDeviceProfile.Manufacturer).Return(TestDeviceProfiles, nil)
	d.On("GetDeviceProfilesByManufacturerModel", TestDeviceProfile.Manufacturer, TestDeviceProfile.Model).Return(TestDeviceProfiles, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)

	// Methods which need to return empty slices so that the business logic does not return a conflict due to the
	// DeviceProfile being in use. This is for update and deletion functionality.
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("UpdateDeviceProfile", mock.Anything).Return(nil)
	d.On("DeleteDeviceProfileById", TestDeviceProfileID).Return(nil)
	d.On("DeleteDeviceProfileByName", TestDeviceProfileName).Return(nil)

	return d
}

func createDBClientDeviceProfileErrorNotFound() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, db.ErrNotFound)
	d.On("GetAllDeviceProfiles").Return(nil, db.ErrNotFound)
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(contract.DeviceProfile{}, db.ErrNotFound)
	d.On("GetDeviceProfilesByModel", TestDeviceProfile.Model).Return(nil, db.ErrNotFound)
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel1).Return(nil, db.ErrNotFound)
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel2).Return(nil, db.ErrNotFound)
	d.On("GetDeviceProfilesByManufacturer", TestDeviceProfile.Manufacturer).Return(nil, db.ErrNotFound)
	d.On("GetDeviceProfilesByManufacturerModel", TestDeviceProfile.Manufacturer, TestDeviceProfile.Model).Return(nil, db.ErrNotFound)

	return d
}

func createMockErrDeviceProfileNotFound() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, errors2.ErrDeviceProfileNotFound{})
	d.On("GetDeviceProfileById", TestDeviceProfileName).Return(contract.DeviceProfile{}, errors2.ErrDeviceProfileNotFound{})
	d.On("GetDeviceProfileByName", TestDeviceProfileID).Return(contract.DeviceProfile{}, errors2.ErrDeviceProfileNotFound{})
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(contract.DeviceProfile{}, errors2.ErrDeviceProfileNotFound{})

	return d
}

func createDBClientMultipleDevicesFoundError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(TestDevices, nil)

	return d
}

func createDBClientMultipleProvisionWatchersFoundError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(TestProvisionWatchers, nil)

	return d
}

func createDBClientGetDevicesByProfileIdError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), TestError)

	return d
}

func createDBClientGetProvisionWatchersByProfileIdError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), TestError)
	d.On("GetAllDeviceProfiles").Return([]contract.DeviceProfile{}, TestError)

	return d
}

func createDBClientPersistDeviceProfileError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(TestDeviceProfile, nil)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(TestDeviceProfile, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return(make([]contract.Device, 0), nil)
	d.On("GetProvisionWatchersByProfileId", TestDeviceProfileID).Return(make([]contract.ProvisionWatcher, 0), nil)
	d.On("GetAllDeviceProfiles").Return(make([]contract.DeviceProfile, 0), nil)

	// Mock both persistence functions so this can be used for updates and delete tests
	d.On("UpdateDeviceProfile", mock.Anything).Return(TestError)
	d.On("DeleteDeviceProfileById", mock.Anything).Return(TestError)

	return d
}

func createDBClientGetDeviceProfileError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetAllDeviceProfiles").Return(nil, TestError)
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(contract.DeviceProfile{}, TestError)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, TestError)
	d.On("GetDeviceProfilesByModel", TestDeviceProfile.Model).Return(nil, TestError)
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel1).Return(nil, TestError)
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel2).Return(nil, TestError)
	d.On("GetDeviceProfilesByManufacturer", TestDeviceProfile.Manufacturer).Return(nil, TestError)
	d.On("GetDeviceProfilesByManufacturerModel", TestDeviceProfile.Manufacturer, TestDeviceProfile.Model).Return(nil, TestError)

	return d
}

func createDBClientGetDeviceProfileMaxLimitError() interfaces.DBClient {
	d := &mocks.DBClient{}
	d.On("GetAllDeviceProfiles").Return(nil, errors2.ErrLimitExceeded{})
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(contract.DeviceProfile{}, errors2.ErrLimitExceeded{})
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, errors2.ErrLimitExceeded{})
	d.On("GetDeviceProfilesByModel", TestDeviceProfile.Model).Return(nil, errors2.ErrLimitExceeded{})
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel1).Return(nil, errors2.ErrLimitExceeded{})
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel2).Return(nil, errors2.ErrLimitExceeded{})
	d.On("GetDeviceProfilesByManufacturer", TestDeviceProfile.Manufacturer).Return(nil, errors2.ErrLimitExceeded{})
	d.On("GetDeviceProfilesByManufacturerModel", TestDeviceProfile.Manufacturer, TestDeviceProfile.Model).Return(nil, errors2.ErrLimitExceeded{})

	return d
}

func createDBClientPersistDeviceInUseError() interfaces.DBClient {
	d := &mocks.DBClient{}
	inUse := TestDeviceProfile
	inUse.DeviceResources = append(inUse.DeviceResources, TestInUseDeviceResource)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(inUse, nil)

	return d
}

func createDBClientUpdateValueDescriptorError() interfaces.DBClient {
	d := &mocks.DBClient{}
	inUse := TestDeviceProfile
	inUse.DeviceResources = append(inUse.DeviceResources, TestInUseDeviceResource)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, TestError)

	return d
}

func createDBServiceClientError(statusCode int) interfaces.DBClient {
	d := &mocks.DBClient{}
	inUse := TestDeviceProfile
	inUse.DeviceResources = append(inUse.DeviceResources, TestInUseDeviceResource)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, types.ErrServiceClient{
		StatusCode: statusCode,
	})

	return d
}

func createDBClientWithOutlines(outlines []mockOutline) interfaces.DBClient {
	dbMock := mocks.DBClient{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg).Return(o.ret, o.err)
	}

	return &dbMock
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
			TestDeleteDeviceResource,
			TestUpdateDeviceResource,
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

type MockValueDescriptorClient struct{}

func (MockValueDescriptorClient) ValueDescriptorsUsage(names []string, ctx context.Context) (map[string]bool, error) {
	usage := map[string]bool{}
	for _, n := range names {
		if n == TestInUseDeviceResource.Name {
			usage[n] = true
			continue
		}

		usage[n] = false
	}
	return usage, nil
}

func (MockValueDescriptorClient) Add(vdr *contract.ValueDescriptor, ctx context.Context) (string, error) {
	return "", nil
}

func (MockValueDescriptorClient) Update(vdr *contract.ValueDescriptor, ctx context.Context) error {
	return nil
}

func (MockValueDescriptorClient) DeleteByName(name string, ctx context.Context) error {
	return nil
}

func (MockValueDescriptorClient) ValueDescriptorForName(name string, ctx context.Context) (contract.ValueDescriptor, error) {
	return contract.ValueDescriptor{Id: name}, nil
}

func (MockValueDescriptorClient) ValueDescriptors(ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("not expected to be invoked")
}

func (MockValueDescriptorClient) ValueDescriptor(id string, ctx context.Context) (contract.ValueDescriptor, error) {
	panic("not expected to be invoked")
}

func (MockValueDescriptorClient) ValueDescriptorsByLabel(label string, ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("not expected to be invoked")
}

func (MockValueDescriptorClient) ValueDescriptorsForDevice(deviceId string, ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("not expected to be invoked")
}

func (MockValueDescriptorClient) ValueDescriptorsForDeviceByName(deviceName string, ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("not expected to be invoked")
}

func (MockValueDescriptorClient) ValueDescriptorsByUomLabel(uomLabel string, ctx context.Context) ([]contract.ValueDescriptor, error) {
	panic("not expected to be invoked")
}

func (MockValueDescriptorClient) Delete(id string, ctx context.Context) error {
	panic("not expected to be invoked")
}
