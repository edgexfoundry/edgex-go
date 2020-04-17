/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package command

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
	commandMocks "github.com/edgexfoundry/edgex-go/internal/core/command/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var knownDeviceName = "Test Device"
var unknownDeviceName = "?"
var cmdURI = clients.ApiBase + "/" + DEVICE
var deviceId = "b3445cc6-87df-48f4-b8b0-587dc8a4e1c2"
var deviceName = ""
var TestCommandId = "TestCommandID"
var testTimestamps = models.Timestamps{Created: 123, Modified: 123, Origin: 123}
var exampleCommand = models.Command{
	Timestamps: testTimestamps,
	Id:         TestCommandId,
	Name:       "testName",
}

var unlockedDevice = models.Device{
	Id:         deviceId,
	Name:       deviceName,
	AdminState: models.Unlocked,
	Service: models.DeviceService{
		Addressable: models.Addressable{Protocol: "http", Address: "localhost",
			Port: 8080}}}
var lockedDevice = models.Device{
	Id:         deviceId,
	Name:       NAME,
	AdminState: models.Locked,
	Service: models.DeviceService{
		Addressable: models.Addressable{Protocol: "http", Address: "localhost",
			Port: 8080}}}

type mockOutline struct {
	methodName string
	arg        []interface{}
	ret        []interface{}
}

func TestRestGetCommandsByDeviceName(t *testing.T) {
	tests := []struct {
		name           string
		dcMock         *mocks.DeviceClient
		dbMock         interfaces.DBClient
		request        *http.Request
		expectedErr    error
		expectedStatus int
	}{
		{
			"ok json content type in header",
			createMockUnlockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodGet,
				clients.ContentTypeJSON,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusOK,
		},
		{
			"ok cbor content type in header",
			createMockUnlockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodGet,
				clients.ContentTypeCBOR,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusOK,
		},
		{
			"no device for requested name",
			createMockUnlockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{unknownDeviceName}, []interface{}{models.DeviceService{}, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{}, db.ErrNotFound}},
			}),
			createGetCommandRequest(),
			nil,
			http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			configuration := config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			}
			restGetCommandsByDeviceName(
				rr,
				tt.request,
				tt.dbMock,
				tt.dcMock,
				&configuration,
				errorconcept.NewErrorHandler(loggerMock))
			response := rr.Result()
			require.Equal(t, tt.expectedStatus, response.StatusCode)
		})
	}
}

func TestRestPutDeviceCommandByCommandID(t *testing.T) {
	tests := []struct {
		name           string
		dcMock         *mocks.DeviceClient
		dbMock         interfaces.DBClient
		request        *http.Request
		expectedErr    error
		expectedStatus int
	}{
		{
			"ok json content type in header",
			createMockUnlockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodPut,
				clients.ContentTypeJSON,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusOK,
		},
		{
			"ok cbor content type in header",
			createMockUnlockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodPut,
				clients.ContentTypeCBOR,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusOK,
		},

		{
			"device was locked with json content type in header",
			createMockLockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodPut,
				clients.ContentTypeJSON,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusLocked,
		},
		{
			"device was locked with cbor content type in header",
			createMockLockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodPut,
				clients.ContentTypeCBOR,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusLocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			httpCaller := createMockHttpCaller()
			restPutDeviceCommandByCommandID(
				rr,
				tt.request,
				loggerMock,
				tt.dbMock,
				tt.dcMock,
				errorconcept.NewErrorHandler(loggerMock),
				httpCaller)
			response := rr.Result()
			require.Equal(t, tt.expectedStatus, response.StatusCode)
		})
	}
}

func TestRestGetDeviceCommandByCommandID(t *testing.T) {
	tests := []struct {
		name           string
		dcMock         *mocks.DeviceClient
		dbMock         interfaces.DBClient
		request        *http.Request
		expectedErr    error
		expectedStatus int
	}{
		{
			"ok json content type in header",
			createMockUnlockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodGet,
				clients.ContentTypeJSON,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusOK,
		},
		{
			"ok cbor content type in header",
			createMockUnlockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodGet,
				clients.ContentTypeCBOR,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusOK,
		},
		{
			"device was locked with json content type in header",
			createMockLockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodGet,
				clients.ContentTypeJSON,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusLocked,
		},
		{
			"device was locked with cbor content type in header",
			createMockLockedDeviceCommandClient(deviceId),
			createMockWithOutlines([]mockOutline{
				{"GetCommandsByName", []interface{}{knownDeviceName}, []interface{}{nil, nil}},
				{"GetCommandsByDeviceId", []interface{}{deviceId}, []interface{}{[]models.Command{exampleCommand}, nil}},
			}),
			createRequestWithPathParameters(
				http.MethodGet,
				clients.ContentTypeCBOR,
				map[string]string{ID: deviceId, COMMANDID: TestCommandId},
				createTestDeviceWithPathUrl(TestCommandId, deviceId),
				exampleCommand),
			nil,
			http.StatusLocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			httpCaller := createMockHttpCaller()
			restGetDeviceCommandByCommandID(
				rr,
				tt.request,
				loggerMock,
				tt.dbMock,
				tt.dcMock,
				errorconcept.NewErrorHandler(loggerMock),
				httpCaller)
			response := rr.Result()
			require.Equal(t, tt.expectedStatus, response.StatusCode)
		})
	}
}

func createMockWithOutlines(outlines []mockOutline) interfaces.DBClient {
	dbMock := commandMocks.DBClient{}

	for _, o := range outlines {
		dbMock.On(o.methodName, o.arg...).Return(o.ret...)
	}
	return &dbMock
}

// An HttpCaller which returns a normal response when a mocked request is executed.
type NormalMockHttpCaller struct{}

// MockNormalBody is a body that will return...
type MockNormalBody struct{}

// Read returns ...
func (MockNormalBody) Read(p []byte) (n int, err error) {
	return 0, nil
}
func (MockNormalBody) ReadFrom(p []byte) (n int, err error) {
	return 0, nil
}

// Close implementation is not required
func (MockNormalBody) Close() error {
	panic("implement me")
}

func (NormalMockHttpCaller) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       MockNormalBody{},
	}, nil
}

func createMockHttpCaller() internal.HttpCaller {
	reqBody := ioutil.NopCloser(bytes.NewReader([]byte(exampleCommand.String())))
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       reqBody,
	}

	req := &mocks.HttpCaller{}
	req.On("Do", mock.Anything).Return(resp, nil)
	return req
}

func createMockUnlockedDeviceCommandClient(deviceId string) *mocks.DeviceClient {
	client := &mocks.DeviceClient{}
	client.On("DeviceForName", mock.Anything, knownDeviceName).Return(unlockedDevice, nil)
	client.On("DeviceForName", mock.Anything, unknownDeviceName).Return(unlockedDevice, db.ErrNotFound)
	client.On("DeviceForName", mock.Anything, "").Return(unlockedDevice, nil)
	client.On("Device", mock.Anything, deviceId).Return(unlockedDevice, nil)
	client.On("Device", mock.Anything, "").Return(unlockedDevice, nil)
	client.On("Device", mock.Anything, "").Return(unlockedDevice, nil)
	return client
}

func createMockLockedDeviceCommandClient(deviceId string) *mocks.DeviceClient {
	client := &mocks.DeviceClient{}
	client.On("DeviceForName", mock.Anything, knownDeviceName).Return(lockedDevice, nil)
	client.On("DeviceForName", mock.Anything, unknownDeviceName).Return(lockedDevice, db.ErrNotFound)
	client.On("DeviceForName", mock.Anything, "").Return(lockedDevice, nil)
	client.On("Device", mock.Anything, deviceId).Return(lockedDevice, nil)
	client.On("Device", mock.Anything, "").Return(lockedDevice, nil)
	client.On("Device", mock.Anything, "").Return(lockedDevice, nil)
	return client
}

func createGetCommandRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, cmdURI, nil)
}

func createRequestWithPathParameters(
	httpMethod string,
	contentType string,
	params map[string]string,
	sampleDevice models.Device,
	cmd models.Command) *http.Request {

	cmdBody, _ := json.Marshal(cmd)
	req, _ := http.NewRequest(httpMethod, cmdURI, strings.NewReader(string(cmdBody)))
	req.Header.Set(clients.ContentType, contentType)
	req.URL.Path = sampleDevice.Service.Addressable.Path
	req.Context()

	return mux.SetURLVars(req, params)
}
