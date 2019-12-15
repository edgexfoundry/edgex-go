package metadata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	metadataConfig "github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	types "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func TestGetCommandsByDeviceId(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		request        *http.Request
		expectedStatus int
	}{
		{
			"OK",
			createCommandByDeviceIdLoaderMock(),
			createCommandRequest(cmdsByDeviceIdURL, ID, deviceId),
			http.StatusOK,
		},
		{
			"Unexpected",
			createMockCommandLoaderMock("GetCommandsByDeviceId", deviceId, nil, unExpectedError),
			createCommandRequest(cmdsByDeviceIdURL, ID, deviceId),
			http.StatusInternalServerError,
		},
		{
			"NotFound",
			createMockCommandLoaderMock("GetCommandsByDeviceId", deviceId, nil, deviceNotFoundErr),
			createCommandRequest(cmdsByDeviceIdURL, ID, deviceId),
			http.StatusNotFound,
		},
		{
			"BadRequest",
			createMockCommandLoaderMock("GetCommandsByDeviceId", deviceId, nil, nil),
			createCommandRequest(cmdsByDeviceIdURL, ID, ErrorPathParam),
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			restGetCommandsByDeviceId(
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

func TestGetAllCommands(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		request        *http.Request
		expectedStatus int
	}{
		{"OK", createCommandsLoaderMock(1), createPlainCommandRequest(cmdsURL), http.StatusOK},
		{"Unexpected", createCommandsLoaderMockUnexpectedFail(), createPlainCommandRequest(cmdsURL), http.StatusInternalServerError},
		{"MaxExceeded", createCommandsLoaderMock(3), createPlainCommandRequest(cmdsURL), http.StatusRequestEntityTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			configuration := metadataConfig.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			}
			restGetAllCommands(
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

func TestGetCommandById(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		request        *http.Request
		expectedStatus int
	}{

		{
			"OK",
			createMockCommandLoaderMock("GetCommandById", commandId, contract.Command{Name: fmt.Sprintf("CommandName"), Id: commandId}, nil),
			createCommandRequest(cmdByIdURL, ID, commandId),
			http.StatusOK,
		},
		{
			"Unexpected",
			createMockCommandLoaderMock("GetCommandById", commandId, contract.Command{}, unExpectedError),
			createCommandRequest(cmdByIdURL, ID, commandId),
			http.StatusInternalServerError,
		},
		{
			"NotFound",
			createMockCommandLoaderMock("GetCommandById", commandId, contract.Command{}, cmdNotFoundErr),
			createCommandRequest(cmdByIdURL, ID, commandId),
			http.StatusNotFound,
		},
		{
			"BadRequest",
			createMockCommandLoaderMock("GetCommandById", commandId, contract.Command{}, nil),
			createCommandRequest(cmdByIdURL, ID, ErrorPathParam),
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			restGetCommandById(
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

func TestGetCommandsByName(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		request        *http.Request
		expectedStatus int
	}{

		{
			"OK",
			createMockCommandLoaderMock("GetCommandsByName", commandName, []contract.Command{{Name: commandName}}, nil),
			createCommandRequest(cmdByNameURL, NAME, commandName),
			http.StatusOK,
		},
		{
			"Unexpected",
			createMockCommandLoaderMock("GetCommandsByName", commandName, nil, unExpectedError),
			createCommandRequest(cmdByNameURL, NAME, commandName),
			http.StatusInternalServerError,
		},
		{
			"BadRequest",
			createMockCommandLoaderMock("GetCommandsByName", commandName, nil, nil),
			createCommandRequest(cmdByNameURL, NAME, ErrorPathParam),
			http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			restGetCommandsByName(
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

//should be common constants
var cmdsURL = clients.ApiBase + "/" + COMMAND
var cmdByIdURL = cmdsURL + "/{" + ID + "}"
var cmdByNameURL = cmdsURL + "/" + NAME + "/{" + NAME + "}"
var cmdsByDeviceIdURL = cmdsURL + "/" + DEVICE + "/{" + ID + "}"

var commandName = "Command 0"
var commandId = "f97b5f0a-ec32-4e96-bd36-02210af16f8c"
var deviceId = "b3445cc6-87df-48f4-b8b0-587dc8a4e1c2"

var unExpectedError = errors.New("unexpected error")
var cmdNotFoundErr = types.NewErrItemNotFound(fmt.Sprintf("command with id %s not found", commandId))
var deviceNotFoundErr = types.NewErrItemNotFound(fmt.Sprintf("device with id %s not found", deviceId))

var commands = []contract.Command{
	{Name: fmt.Sprintf(commandName)},
	{Name: fmt.Sprintf("Command 1")},
	{Name: fmt.Sprintf("Command 2")},
}

func createPlainCommandRequest(url string) *http.Request {

	return createCommandRequest(url, "", "")
}

func createCommandRequest(url string, pathParamName string, pathParamValue string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	if pathParamName == "" && pathParamValue == "" {
		return req
	}
	return mux.SetURLVars(req, map[string]string{pathParamName: pathParamValue})
}

func createCommandByDeviceIdLoaderMock() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", deviceId).Return(commands, nil)
	return dbMock
}

//TestGetAllCommands Mocks
func createCommandsLoaderMock(howMany int) interfaces.DBClient {
	commands := []contract.Command{}
	for i := 0; i < howMany; i++ {
		commands = append(commands, contract.Command{Name: fmt.Sprintf("Command %v", i)})
	}

	dbMock := &mocks.DBClient{}
	dbMock.On("GetAllCommands").Return(commands, nil)
	return dbMock
}

func createCommandsLoaderMockUnexpectedFail() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetAllCommands").Return(nil, errors.New("unexpected error"))
	return dbMock
}

func createMockCommandLoaderMock(methodName string, arg string, result interface{}, err error) interfaces.DBClient {
	myMock := &mocks.DBClient{}
	myMock.On(methodName, arg).Return(result, err)
	return myMock
}
