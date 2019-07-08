package metadata

import (
	"fmt"
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
		{"OK", createCommandLoaderMock(), createCommandRequest(cmdsByDeviceIdURL, ID, deviceId), http.StatusOK},
		{"Unexpected", createCommandLoaderMockUnexpectedFail(), createCommandRequest(cmdsByDeviceIdURL, ID, deviceId), http.StatusInternalServerError},
		{"NotFound", createCommandLoaderMockNotFoundFail(), createCommandRequest(cmdsByDeviceIdURL, ID, deviceId), http.StatusNotFound},
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

var deviceId = "b3445cc6-87df-48f4-b8b0-587dc8a4e1c2"
var cmdsByDeviceIdURL = clients.ApiBase + "/" + COMMAND + "/" + DEVICE + "/{" + ID + "}"

func createCommandRequest(url string, pathParamName string, pathParamValue string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	return mux.SetURLVars(req, map[string]string{pathParamName: pathParamValue})
}

func createCommandLoaderMock() interfaces.DBClient {
	commands := []contract.Command{}
	for i := 0; i < 3; i++ {
		commands = append(commands, contract.Command{Name: fmt.Sprintf("Command %v", i)})
	}

	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", deviceId).Return(commands, nil)
	return dbMock
}

func createCommandLoaderMockUnexpectedFail() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", deviceId).Return(nil, errors.New("unexpected error"))
	return dbMock
}

func createCommandLoaderMockNotFoundFail() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", deviceId).Return(nil, types.NewErrItemNotFound(fmt.Sprintf("device with id %s not found", deviceId)))
	return dbMock
}
