package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	goErrors "errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	metadataConfig "github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
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
	createTestDeviceProfileWithCommands(
		"TestDeviceProfileID2",
		"TestDeviceProfileName2",
		[]string{TestDeviceProfileLabel1,
			TestDeviceProfileLabel2},
		TestDeviceProfileManufacturer,
		TestDeviceProfileModel,
		TestCommand),
	createTestDeviceProfileWithCommands(
		"TestErrorID",
		"TestErrorName",
		[]string{TestLabelError1,
			TestLabelError2},
		"TestErrorManufacturer",
		"TestErrorModel",
		TestCommand),
}
var TestDeviceProfileValidated = createValidatedTestDeviceProfile()
var TestError = goErrors.New("test error")
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

var TestDeleteDeviceResource = contract.DeviceResource{Name: "TestDeleteDeviceResource"}
var TestUpdateDeviceResource = contract.DeviceResource{Name: "TestUpdateDeviceResource"}
var TestInUseDeviceResource = contract.DeviceResource{Name: "TestInUseDeviceResource"}

func TestGetAllProfiles(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createDBClient(),
			http.StatusOK,
		},
		{
			"Max result count exceeded",
			createDBClientGetDeviceProfileMaxLimitError(),
			http.StatusRequestEntityTooLarge,
		},
		{
			"Device Profile Not Found",
			createDBClientDeviceProfileErrorNotFound(),
			http.StatusInternalServerError,
		},
		{
			"Database error",
			createDBClientGetDeviceProfileError(),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			configuration := metadataConfig.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{MaxResultCount: len(TestDeviceProfiles)},
			}
			restGetAllDeviceProfiles(
				rr,
				loggerMock,
				tt.dbMock,
				errorconcept.NewErrorHandler(loggerMock),
				&configuration)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestAddDeviceProfile(t *testing.T) {
	// this one uses validated because it needs it for input and the dbMock
	emptyName := TestDeviceProfileValidated
	emptyName.Name = ""

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		vdcMock        MockValueDescriptorClient
		expectedStatus int
	}{
		{
			"OK",
			createRequestWithBody(http.MethodPost, TestDeviceProfile),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{TestDeviceProfileValidated}, []interface{}{TestDeviceProfileID, nil}},
				{"GetDeviceProfileByName", []interface{}{TestDeviceProfileName}, []interface{}{contract.DeviceProfile{}, db.ErrNotFound}},
			}),
			MockValueDescriptorClient{},
			http.StatusOK,
		},
		{
			"OK with value descriptor management",
			createRequestWithBody(http.MethodPost, TestDeviceProfile),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{TestDeviceProfileValidated}, []interface{}{TestDeviceProfileID, nil}},
				{"GetDeviceProfileByName", []interface{}{TestDeviceProfileName}, []interface{}{contract.DeviceProfile{}, db.ErrNotFound}},
			}),
			MockValueDescriptorClient{},
			http.StatusOK,
		},
		{
			"Value descriptor management service client error",
			createRequestWithBody(http.MethodPost, TestDeviceProfile),
			createDBClientWithOutlines([]mockOutline{
				{"GetDeviceProfileByName", []interface{}{TestDeviceProfileName}, []interface{}{contract.DeviceProfile{}, db.ErrNotFound}},
			}),
			MockValueDescriptorClient{types.ErrServiceClient{StatusCode: http.StatusTeapot}},
			http.StatusTeapot,
		},
		{
			"Value descriptor management service other error",
			createRequestWithBody(http.MethodPost, TestDeviceProfile),
			createDBClientWithOutlines([]mockOutline{
				{"GetDeviceProfileByName", []interface{}{TestDeviceProfileName}, []interface{}{contract.DeviceProfile{}, db.ErrNotFound}},
			}),
			MockValueDescriptorClient{TestError},
			http.StatusInternalServerError,
		},
		{
			"YAML unmarshal error",
			httptest.NewRequest(http.MethodPost, AddressableTestURI, bytes.NewBuffer(nil)),
			nil,
			MockValueDescriptorClient{},
			http.StatusBadRequest,
		},
		{
			"Duplicate commands",
			createRequestWithBody(http.MethodPost, createTestDeviceProfileWithCommands(TestDeviceProfileID, TestDeviceProfileName, TestDeviceProfileLabels, TestDeviceProfileManufacturer, TestDeviceProfileModel, TestCommand, TestCommand)),
			nil,
			MockValueDescriptorClient{},
			http.StatusBadRequest,
		},
		{
			"Empty device profile name",
			createRequestWithBody(http.MethodPost, emptyName),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{emptyName}, []interface{}{"", db.ErrNameEmpty}},
				{"GetDeviceProfileByName", []interface{}{""}, []interface{}{contract.DeviceProfile{}, db.ErrNotFound}},
			}),
			MockValueDescriptorClient{},
			http.StatusBadRequest,
		},
		{
			"Unsuccessful database call",
			createRequestWithBody(http.MethodPost, TestDeviceProfile),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{TestDeviceProfileValidated}, []interface{}{"", TestError}},
				{"GetDeviceProfileByName", []interface{}{TestDeviceProfileName}, []interface{}{contract.DeviceProfile{}, db.ErrNotFound}},
			}),
			MockValueDescriptorClient{},
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			configuration := metadataConfig.ConfigurationStruct{
				Writable: metadataConfig.WritableInfo{EnableValueDescriptorManagement: true},
				Service:  bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			}
			restAddDeviceProfile(
				rr,
				tt.request,
				loggerMock,
				tt.dbMock,
				errorconcept.NewErrorHandler(loggerMock),
				tt.vdcMock,
				&configuration)
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
			rr := httptest.NewRecorder()
			restGetProfileByProfileId(rr, tt.request, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
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
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			restGetYamlProfileById(
				rr,
				tt.request,
				loggerMock,
				tt.dbMock,
				errorconcept.NewErrorHandler(loggerMock))
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
			rr := httptest.NewRecorder()
			restGetProfileByName(rr, tt.request, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
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
			rr := httptest.NewRecorder()
			restGetYamlProfileByName(rr, tt.request, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
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
			rr := httptest.NewRecorder()
			restGetProfileWithLabel(rr, tt.request, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
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
			rr := httptest.NewRecorder()
			restGetProfileByManufacturer(
				rr,
				tt.request,
				tt.dbMock,
				errorconcept.NewErrorHandler(logger.NewMockClient()))
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
			rr := httptest.NewRecorder()
			restGetProfileByModel(rr, tt.request, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
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
			rr := httptest.NewRecorder()
			restGetProfileByManufacturerModel(
				rr,
				tt.request,
				tt.dbMock,
				errorconcept.NewErrorHandler(logger.NewMockClient()))
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
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createDBClient(),
			true,
			http.StatusOK,
		},
		{
			"ValueDescriptor in use error",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createDBClientPersistDeviceInUseError(),
			true,
			http.StatusConflict,
		},
		{
			"ValueDescriptor update error",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createDBClientUpdateValueDescriptorError(),
			true,
			http.StatusInternalServerError,
		},
		{
			"Multiple devices associated with device profile",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createDBClientMultipleDevicesFoundError(),
			true,
			http.StatusConflict,
		},
		{
			"Multiple provision watchers associated with device profile",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
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
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createMockErrDeviceProfileNotFound(),
			true,
			http.StatusNotFound,
		},
		{
			"Device Profile Not Found(UpdateDeviceProfileExecutor)",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createMockErrDeviceProfileNotFound(),
			false,
			http.StatusNotFound,
		},
		{
			"GetProvisionWatchersByProfileId database error ",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createDBClientGetProvisionWatchersByProfileIdError(),
			true,
			http.StatusInternalServerError,
		},
		{
			"Service client error ",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createDBServiceClientError(http.StatusTeapot),
			true,
			http.StatusTeapot,
		},
		{
			"UpdateDeviceProfile database error ",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createDBClientPersistDeviceProfileError(),
			true,
			http.StatusInternalServerError,
		},
		{
			"GetDevicesByProfileId database error",
			createRequestWithBody(http.MethodPut, TestDeviceProfile),
			createDBClientGetDevicesByProfileIdError(),
			true,
			http.StatusInternalServerError,
		},
		{
			"Duplicate commands error ",
			createRequestWithBody(http.MethodPut, createTestDeviceProfileWithCommands(TestDeviceProfileID, TestDeviceProfileName, TestDeviceProfileLabels, TestDeviceProfileManufacturer, TestDeviceProfileModel, TestCommand, TestCommand)),
			createDBClient(),
			true,
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			configuration := metadataConfig.ConfigurationStruct{
				Writable: metadataConfig.WritableInfo{
					EnableValueDescriptorManagement: tt.enableValueDescriptorManagement},
				Service: bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			}
			restUpdateDeviceProfile(
				rr,
				tt.request,
				loggerMock,
				tt.dbMock,
				errorconcept.NewErrorHandler(loggerMock),
				MockValueDescriptorClient{},
				&configuration)
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
			rr := httptest.NewRecorder()
			restDeleteProfileByProfileId(
				rr,
				tt.request,
				tt.dbMock,
				errorconcept.NewErrorHandler(logger.NewMockClient()))
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
			rr := httptest.NewRecorder()
			restDeleteProfileByName(rr, tt.request, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestAddProfileByYaml(t *testing.T) {
	// we have to overwrite the CoreCommands so that their isValidated is false
	dp := TestDeviceProfile
	dp.CoreCommands = []contract.Command{TestCommand}

	okBody, _ := yaml.Marshal(dp)

	duplicateCommandName := dp
	duplicateCommandName.CoreCommands = []contract.Command{TestCommand, TestCommand}
	dupeBody, _ := yaml.Marshal(duplicateCommandName)

	emptyName := dp
	emptyName.Name = ""
	emptyBody, _ := yaml.Marshal(emptyName)

	emptyFileRequest := createDeviceProfileRequestWithFile(okBody)
	emptyFileRequest.MultipartForm = new(multipart.Form)
	emptyFileRequest.MultipartForm.File = nil

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createDeviceProfileRequestWithFile(okBody),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{dp}, []interface{}{TestDeviceProfileID, nil}},
			}),
			http.StatusOK,
		},
		{
			"Wrong content type",
			httptest.NewRequest(http.MethodPut, AddressableTestURI, bytes.NewBuffer(okBody)),
			nil,
			http.StatusInternalServerError,
		},
		{
			"Missing file",
			emptyFileRequest,
			nil,
			http.StatusBadRequest,
		},
		{
			"Empty file",
			createDeviceProfileRequestWithFile([]byte{}),
			nil,
			http.StatusBadRequest,
		},
		{
			"Duplicate command name",
			createDeviceProfileRequestWithFile(dupeBody),
			nil,
			http.StatusConflict,
		},
		{
			"Duplicate profile name",
			createDeviceProfileRequestWithFile(okBody),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{dp}, []interface{}{"", db.ErrNotUnique}},
			}),
			http.StatusConflict,
		},
		{
			"Empty device profile name",
			createDeviceProfileRequestWithFile(emptyBody),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{emptyName}, []interface{}{"", db.ErrNameEmpty}},
			}),
			http.StatusBadRequest,
		},
		{
			"Unsuccessful database call",
			createDeviceProfileRequestWithFile(okBody),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{dp}, []interface{}{"", TestError}},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restAddProfileByYaml(rr, tt.request, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func TestAddProfileByYamlRaw(t *testing.T) {
	// we have to overwrite the CoreCommands so that their isValidated is false
	dp := TestDeviceProfile
	dp.CoreCommands = []contract.Command{TestCommand}

	okBody, _ := yaml.Marshal(dp)

	duplicateCommandName := dp
	duplicateCommandName.CoreCommands = []contract.Command{TestCommand, TestCommand}
	dupeBody, _ := yaml.Marshal(duplicateCommandName)

	emptyName := dp
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
			httptest.NewRequest(http.MethodPut, AddressableTestURI, bytes.NewBuffer(okBody)),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{dp}, []interface{}{TestDeviceProfileID, nil}},
			}),
			http.StatusOK,
		},
		{
			"YAML unmarshal error",
			createDeviceProfileRequestWithFile(dupeBody),
			nil,
			http.StatusServiceUnavailable,
		},
		{
			"Empty device profile name",
			httptest.NewRequest(http.MethodPut, AddressableTestURI, bytes.NewBuffer(emptyBody)),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{emptyName}, []interface{}{"", db.ErrNameEmpty}},
			}),

			http.StatusBadRequest,
		},
		{
			"Unsuccessful database call",
			httptest.NewRequest(http.MethodPut, AddressableTestURI, bytes.NewBuffer(okBody)),
			createDBClientWithOutlines([]mockOutline{
				{"AddDeviceProfile", []interface{}{dp}, []interface{}{"", TestError}},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restAddProfileByYamlRaw(rr, tt.request, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)

				return
			}
		})
	}
}

func createDeviceProfileRequestWithFile(fileContents []byte) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "deviceProfile.yaml")
	if err != nil {
		return nil
	}
	_, err = part.Write(fileContents)
	if err != nil {
		return nil
	}
	boundary := writer.Boundary()

	err = writer.Close()
	if err != nil {
		return nil
	}

	req, _ := http.NewRequest(http.MethodPost, AddressableTestURI, body)
	req.Header.Set(clients.ContentType, "multipart/form-data; boundary="+boundary)
	return req
}

func createRequestWithBody(method string, d contract.DeviceProfile) *http.Request {
	body, err := json.Marshal(d)
	if err != nil {
		panic("Failed to create test JSON:" + err.Error())
	}

	return httptest.NewRequest(method, AddressableTestURI, bytes.NewBuffer(body))
}

func createRequestWithPathParameters(httpMethod string, params map[string]string) *http.Request {
	req := httptest.NewRequest(httpMethod, AddressableTestURI, nil)
	return mux.SetURLVars(req, params)
}

func createRequestWithInvalidBody() *http.Request {
	return httptest.NewRequest(http.MethodPut, AddressableTestURI, bytes.NewBufferString("Bad JSON"))
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
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, errors.ErrDeviceProfileNotFound{})
	d.On("GetDeviceProfileById", TestDeviceProfileName).Return(contract.DeviceProfile{}, errors.ErrDeviceProfileNotFound{})
	d.On("GetDeviceProfileByName", TestDeviceProfileID).Return(contract.DeviceProfile{}, errors.ErrDeviceProfileNotFound{})
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(contract.DeviceProfile{}, errors.ErrDeviceProfileNotFound{})

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
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, TestError)
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
	d.On("GetAllDeviceProfiles").Return(nil, errors.ErrLimitExceeded{})
	d.On("GetDeviceProfileById", TestDeviceProfileID).Return(contract.DeviceProfile{}, errors.ErrLimitExceeded{})
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(contract.DeviceProfile{}, errors.ErrLimitExceeded{})
	d.On("GetDeviceProfilesByModel", TestDeviceProfile.Model).Return(nil, errors.ErrLimitExceeded{})
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel1).Return(nil, errors.ErrLimitExceeded{})
	d.On("GetDeviceProfilesWithLabel", TestDeviceProfileLabel2).Return(nil, errors.ErrLimitExceeded{})
	d.On("GetDeviceProfilesByManufacturer", TestDeviceProfile.Manufacturer).Return(nil, errors.ErrLimitExceeded{})
	d.On("GetDeviceProfilesByManufacturerModel", TestDeviceProfile.Manufacturer, TestDeviceProfile.Model).Return(nil, errors.ErrLimitExceeded{})

	return d
}

func createDBClientPersistDeviceInUseError() interfaces.DBClient {
	d := &mocks.DBClient{}
	inUse := TestDeviceProfile
	inUse.DeviceResources = append(inUse.DeviceResources, TestInUseDeviceResource)
	d.On("GetDeviceProfileByName", TestDeviceProfileName).Return(inUse, nil)
	d.On("GetDevicesByProfileId", TestDeviceProfileID).Return([]contract.Device{}, nil)

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
		dbMock.On(o.methodName, o.arg...).Return(o.ret...)
	}

	return &dbMock
}

// createTestDeviceProfile creates a device profile to be used during testing.
// This function handles some of the necessary creation nuances which need to take place for proper mocking and equality
// verifications.
func createTestDeviceProfile() contract.DeviceProfile {
	return createTestDeviceProfileWithCommands(TestDeviceProfileID, TestDeviceProfileName, TestDeviceProfileLabels, TestDeviceProfileManufacturer, TestDeviceProfileModel, TestCommand)
}

// createValidatedTestDeviceProfile creates an object by deserializing it from JSON
// so that its unexported field isValidated will be true.
func createValidatedTestDeviceProfile() contract.DeviceProfile {
	bytes, _ := json.Marshal(TestDeviceProfile)
	var dp contract.DeviceProfile
	_ = json.Unmarshal(bytes, &dp)

	return dp
}

// createTestDeviceProfileWithCommands creates a device profile to be used during testing.
// This function handles some of the necessary creation nuances which need to take place for proper mocking and equality
// verifications.
func createTestDeviceProfileWithCommands(
	id string,
	name string,
	labels []string,
	manufacturer string,
	model string,
	commands ...contract.Command) contract.DeviceProfile {

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

type MockValueDescriptorClient struct {
	errorToThrow error
}

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

func (mvdc MockValueDescriptorClient) Add(vdr *contract.ValueDescriptor, ctx context.Context) (string, error) {
	if mvdc.errorToThrow != nil {
		return "", mvdc.errorToThrow
	}

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
