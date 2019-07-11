package metadata

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
	"testing"

	types "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestGetCommandsByDeviceId(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		request        *http.Request
		expectedStatus int
	}{
		{"OK", createCommandByDeviceIdLoaderMock(), createCommandRequest(cmdsByDeviceIdURL, ID, deviceId), http.StatusOK},
		{"Unexpected", createCommandByDeviceIdLoaderMockUnexpectedFail(), createCommandRequest(cmdsByDeviceIdURL, ID, deviceId), http.StatusInternalServerError},
		{"NotFound", createCommandByDeviceIdLoaderMockNotFoundFail(), createCommandRequest(cmdsByDeviceIdURL, ID, deviceId), http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetCommandsByDeviceId)

			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetAllCommands(t *testing.T) {
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
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
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAllCommands)

			handler.ServeHTTP(rr, tt.request)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
	Configuration = &ConfigurationStruct{}
}

var cmdsURL = clients.ApiBase + "/" + COMMAND
var cmdsByDeviceIdURL = clients.ApiBase + "/" + COMMAND + "/" + DEVICE + "/{" + ID + "}"

var deviceId = "b3445cc6-87df-48f4-b8b0-587dc8a4e1c2"

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

//TestGetCommandsByDeviceId Mocks
func createCommandByDeviceIdLoaderMock() interfaces.DBClient {
	commands := []contract.Command{}
	for i := 0; i < 3; i++ {
		commands = append(commands, contract.Command{Name: fmt.Sprintf("Command %v", i)})
	}

	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", deviceId).Return(commands, nil)
	return dbMock
}

func createCommandByDeviceIdLoaderMockUnexpectedFail() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", deviceId).Return(nil, errors.New("unexpected error"))
	return dbMock
}

func createCommandByDeviceIdLoaderMockNotFoundFail() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", deviceId).Return(nil, types.NewErrItemNotFound(fmt.Sprintf("device with id %s not found", deviceId)))
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
