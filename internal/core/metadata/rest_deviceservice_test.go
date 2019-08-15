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

package metadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var testDeviceServiceId = uuid.New().String()
var testDeviceServiceName = "test service"
var testDeviceService = contract.DeviceService{Id: testDeviceServiceId, Name: testDeviceServiceName}
var testOperatingState, _ = contract.GetOperatingState(contract.Enabled)
var testAdminState, _ = contract.GetAdminState(contract.Unlocked)
var testError = errors.New("some error")

func TestGetAllDeviceServices(t *testing.T) {
	Configuration = &ConfigurationStruct{Service: config.ServiceInfo{MaxResultCount: 1}}
	req, err := http.NewRequest(http.MethodGet, clients.ApiBase+"/"+DEVICESERVICE, nil)
	if err != nil {
		t.Error(err)
		return
	}

	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK", createGetDeviceServiceLoaderMock(1), http.StatusOK},
		{"MaxExceeded", createGetDeviceServiceLoaderMock(2), http.StatusRequestEntityTooLarge},
		{"Unexpected", createGetDeviceServiceLoaderMockFail(), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restGetAllDeviceServices)

			handler.ServeHTTP(rr, req)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
	Configuration = &ConfigurationStruct{}
}

func TestUpdateOpStateById(t *testing.T) {
	operatingStateEnabled := testDeviceService
	operatingStateEnabled.OperatingState = testOperatingState

	Configuration = &ConfigurationStruct{}

	tests := []struct {
		name           string
		req            *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
				{"UpdateDeviceService", operatingStateEnabled, nil, nil}}),
			http.StatusOK,
		},
		{"Invalid operating state",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, OPSTATE: UNLOCKED}),
			nil,
			http.StatusBadRequest,
		},
		{"Device service not found",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, db.ErrNotFound}}),
			http.StatusNotFound,
		},
		{"Device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, testError}}),
			http.StatusInternalServerError,
		},
		{"Update error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
				{"UpdateDeviceService", operatingStateEnabled, testError, nil}}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restUpdateServiceOpStateById)

			handler.ServeHTTP(rr, tt.req)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestUpdateOpStateByName(t *testing.T) {
	operatingStateEnabled := testDeviceService
	operatingStateEnabled.OperatingState = testOperatingState

	Configuration = &ConfigurationStruct{}

	tests := []struct {
		name           string
		req            *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
				{"UpdateDeviceService", operatingStateEnabled, nil, nil}}),
			http.StatusOK,
		},
		{"Invalid operating state",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, OPSTATE: UNLOCKED}),
			nil,
			http.StatusBadRequest,
		},
		{"Invalid name",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: "%ERR", OPSTATE: ENABLED}),
			nil,
			http.StatusBadRequest,
		},
		{"Device service not found",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, db.ErrNotFound}}),
			http.StatusNotFound,
		},
		{"Device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, testError}}),
			http.StatusInternalServerError,
		},
		{"Update error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
				{"UpdateDeviceService", operatingStateEnabled, testError, nil}}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restUpdateServiceOpStateByName)

			handler.ServeHTTP(rr, tt.req)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestUpdateAdminStateById(t *testing.T) {
	adminStateEnabled := testDeviceService
	adminStateEnabled.AdminState = testAdminState

	Configuration = &ConfigurationStruct{}

	tests := []struct {
		name           string
		req            *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
				{"UpdateDeviceService", adminStateEnabled, nil, nil}}),
			http.StatusOK,
		},
		{"Invalid admin state",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, ADMINSTATE: ENABLED}),
			nil,
			http.StatusBadRequest,
		},
		{"Device service not found",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, db.ErrNotFound}}),
			http.StatusNotFound,
		},
		{"Device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, contract.DeviceService{}, testError}}),
			http.StatusInternalServerError,
		},
		{"Update error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", testDeviceServiceId, testDeviceService, nil},
				{"UpdateDeviceService", adminStateEnabled, testError, nil}}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restUpdateServiceAdminStateById)

			handler.ServeHTTP(rr, tt.req)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestUpdateAdminStateByName(t *testing.T) {
	adminStateEnabled := testDeviceService
	adminStateEnabled.AdminState = testAdminState

	Configuration = &ConfigurationStruct{}

	tests := []struct {
		name           string
		req            *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
				{"UpdateDeviceService", adminStateEnabled, nil, nil}}),
			http.StatusOK,
		},
		{"Invalid admin state",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, ADMINSTATE: ENABLED}),
			nil,
			http.StatusBadRequest,
		},
		{"Invalid name",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: "%ERR", ADMINSTATE: UNLOCKED}),
			nil,
			http.StatusBadRequest,
		},
		{"Device service not found",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, db.ErrNotFound}}),
			http.StatusNotFound,
		},
		{"Device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, contract.DeviceService{}, testError}}),
			http.StatusInternalServerError,
		},
		{"Update error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", testDeviceServiceName, testDeviceService, nil},
				{"UpdateDeviceService", adminStateEnabled, testError, nil}}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbClient = tt.dbMock
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(restUpdateServiceAdminStateByName)

			handler.ServeHTTP(rr, tt.req)
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func createGetDeviceServiceLoaderMock(howMany int) interfaces.DBClient {
	services := []contract.DeviceService{}
	for i := 0; i < howMany; i++ {
		services = append(services, contract.DeviceService{Name: "Test Device Service"})
	}

	dbMock := &mocks.DBClient{}
	dbMock.On("GetAllDeviceServices").Return(services, nil)
	return dbMock
}

func createGetDeviceServiceLoaderMockFail() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetAllDeviceServices").Return(nil, errors.New("unexpected error"))
	return dbMock
}

func createDeviceServiceRequestWithBody(httpMethod string, deviceService contract.DeviceService, pathParams map[string]string) *http.Request {
	// if your JSON marshalling fails you've got bigger problems
	body, _ := json.Marshal(deviceService)

	req := httptest.NewRequest(httpMethod, clients.ApiBase+"/"+DEVICESERVICE, bytes.NewReader(body))

	return mux.SetURLVars(req, pathParams)
}
