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
	"strconv"
	"testing"

	metadataConfig "github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/errorconcept"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var testDeviceServiceURI = clients.ApiBase + "/" + DEVICESERVICE
var testAddressable = contract.Addressable{Id: testDeviceServiceId, Name: testDeviceServiceName}
var testDeviceServiceId = uuid.New().String()
var testDeviceServiceName = "test service"
var testDeviceService = contract.DeviceService{Id: testDeviceServiceId, Name: testDeviceServiceName}
var testDeviceServices = []contract.DeviceService{testDeviceService}
var testOperatingState, _ = contract.GetOperatingState(contract.Enabled)
var testAdminState, _ = contract.GetAdminState(contract.Unlocked)
var testError = errors.New("some error")
var testDeviceServiceLastUpdated = contract.DeviceService{
	Id:           testDeviceServiceLastUpdatedId,
	Name:         testDeviceServiceName,
	LastReported: lastReported}
var testDeviceServiceLastUpdatedId = "b3445cc6-87df-48f4-b8b0-587dc8a4e1c2"
var testDeviceServiceLastUpdatedNotExist = contract.DeviceService{
	Id:           testDeviceServiceLastUpdatedNotExistId,
	Name:         testDeviceServiceName,
	LastReported: lastReported}
var testDeviceServiceLastUpdatedNotExistId = "0"

const (
	lastReported = 123546789
)

func TestGetAllDeviceServices(t *testing.T) {
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
			rr := httptest.NewRecorder()
			var loggerMock = logger.NewMockClient()
			configuration := metadataConfig.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{MaxResultCount: 1},
			}
			restGetAllDeviceServices(
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

func TestGetServiceByName(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		value          string
		expectedStatus int
	}{
		{
			"OK",
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{testDeviceService, nil}},
			}),
			testDeviceServiceName,
			http.StatusOK,
		},
		{
			"Invalid name",
			nil,
			"%ERR",
			http.StatusBadRequest,
		},
		{
			"Device service not found",
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{contract.DeviceService{}, db.ErrNotFound}},
			}),
			testDeviceServiceName,
			http.StatusNotFound,
		},
		{
			"Device services lookup error",
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{contract.DeviceService{}, testError}},
			}),
			testDeviceServiceName,
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetServiceByName(
				rr,
				createDeviceServiceRequest(http.MethodGet, NAME, tt.value),
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

func TestGetServiceById(t *testing.T) {
	req := createDeviceServiceRequest(http.MethodGet, ID, testDeviceServiceId)

	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"OK",
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{testDeviceService, nil}},
			}),
			http.StatusOK,
		},
		{
			"Device service not found",
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{contract.DeviceService{}, db.ErrNotFound}},
			}),
			http.StatusNotFound,
		},
		{
			"Device services lookup error",
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{contract.DeviceService{}, testError}},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetServiceById(rr, req, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
			response := rr.Result()
			if response.StatusCode != tt.expectedStatus {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, response.StatusCode)
				return
			}
		})
	}
}

func TestGetServiceByAddressableName(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		value          string
		expectedStatus int
	}{
		{
			"OK",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", []interface{}{testDeviceServiceName}, []interface{}{testAddressable, nil}},
				{"GetDeviceServicesByAddressableId", []interface{}{testDeviceServiceId}, []interface{}{testDeviceServices, nil}},
			}),
			testDeviceServiceName,
			http.StatusOK,
		},
		{
			"Addressable not found",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", []interface{}{testDeviceServiceName}, []interface{}{contract.Addressable{}, db.ErrNotFound}},
			}),
			testDeviceServiceName,
			http.StatusNotFound,
		},
		{
			"No name provided",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", []interface{}{""}, []interface{}{contract.Addressable{}, db.ErrNotFound}},
			}),
			"",
			http.StatusNotFound,
		},
		{
			"Invalid name",
			nil,
			"%ERR",
			http.StatusBadRequest,
		},
		{
			"Addressable lookup error",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", []interface{}{testDeviceServiceName}, []interface{}{contract.Addressable{}, testError}},
			}),
			testDeviceServiceName,
			http.StatusInternalServerError,
		},
		{
			"Device services lookup error",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", []interface{}{testDeviceServiceName}, []interface{}{testAddressable, nil}},
				{"GetDeviceServicesByAddressableId", []interface{}{testDeviceServiceId}, []interface{}{nil, testError}},
			}),
			testDeviceServiceName,
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetServiceByAddressableName(
				rr,
				createDeviceServiceRequest(http.MethodGet, ADDRESSABLENAME, tt.value),
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

func TestGetServiceByAddressableId(t *testing.T) {
	tests := []struct {
		name           string
		dbMock         interfaces.DBClient
		value          string
		expectedStatus int
	}{
		{
			"OK",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableById", []interface{}{testDeviceServiceId}, []interface{}{testAddressable, nil}},
				{"GetDeviceServicesByAddressableId", []interface{}{testDeviceServiceId}, []interface{}{testDeviceServices, nil}},
			}),
			testDeviceServiceId,
			http.StatusOK,
		},
		{
			"No ID provided",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableByName", []interface{}{""}, []interface{}{contract.Addressable{}, db.ErrNotFound}},
			}),
			"",
			http.StatusNotFound,
		},
		{
			"Addressable not found",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableById", []interface{}{testDeviceServiceId}, []interface{}{contract.Addressable{}, db.ErrNotFound}},
			}),
			testDeviceServiceId,
			http.StatusNotFound,
		},
		{
			"Addressable lookup error",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableById", []interface{}{testDeviceServiceId}, []interface{}{contract.Addressable{}, testError}},
			}),
			testDeviceServiceId,
			http.StatusInternalServerError,
		},
		{
			"Device services lookup error",
			createMockWithOutlines([]mockOutline{
				{"GetAddressableById", []interface{}{testDeviceServiceId}, []interface{}{testAddressable, nil}},
				{"GetDeviceServicesByAddressableId", []interface{}{testDeviceServiceId}, []interface{}{nil, testError}},
			}),
			testDeviceServiceId,
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restGetServiceByAddressableId(
				rr,
				createDeviceServiceRequest(http.MethodGet, ADDRESSABLEID, tt.value),
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

func TestUpdateOpStateById(t *testing.T) {
	operatingStateEnabled := testDeviceService
	operatingStateEnabled.OperatingState = testOperatingState

	tests := []struct {
		name           string
		req            *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{testDeviceService, nil}},
				{"UpdateDeviceService", []interface{}{operatingStateEnabled}, []interface{}{nil, nil}},
			}),
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
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{contract.DeviceService{}, db.ErrNotFound}},
			}),
			http.StatusNotFound,
		},
		{"Device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{contract.DeviceService{}, testError}},
			}),
			http.StatusInternalServerError,
		},
		{"Update error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{testDeviceService, nil}},
				{"UpdateDeviceService", []interface{}{operatingStateEnabled}, []interface{}{testError, nil}},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restUpdateServiceOpStateById(rr, tt.req, tt.dbMock, errorconcept.NewErrorHandler(logger.NewMockClient()))
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

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{testDeviceService, nil}},
				{"UpdateDeviceService", []interface{}{operatingStateEnabled}, []interface{}{nil, nil}},
			}),
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
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{contract.DeviceService{}, db.ErrNotFound}},
			}),
			http.StatusNotFound,
		},
		{"Device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{contract.DeviceService{}, testError}},
			}),
			http.StatusInternalServerError,
		},
		{"Update error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, OPSTATE: ENABLED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{testDeviceService, nil}},
				{"UpdateDeviceService", []interface{}{operatingStateEnabled}, []interface{}{testError, nil}},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restUpdateServiceOpStateByName(
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

func TestUpdateAdminStateById(t *testing.T) {
	adminStateEnabled := testDeviceService
	adminStateEnabled.AdminState = testAdminState

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{testDeviceService, nil}},
				{"UpdateDeviceService", []interface{}{adminStateEnabled}, []interface{}{nil, nil}},
			}),
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
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{contract.DeviceService{}, db.ErrNotFound}},
			}),
			http.StatusNotFound,
		},
		{"Device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{contract.DeviceService{}, testError}},
			}),
			http.StatusInternalServerError,
		},
		{"Update error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{ID: testDeviceServiceId, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceId}, []interface{}{testDeviceService, nil}},
				{"UpdateDeviceService", []interface{}{adminStateEnabled}, []interface{}{testError, nil}},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restUpdateServiceAdminStateById(
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

func TestUpdateAdminStateByName(t *testing.T) {
	adminStateEnabled := testDeviceService
	adminStateEnabled.AdminState = testAdminState

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{"OK",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{testDeviceService, nil}},
				{"UpdateDeviceService", []interface{}{adminStateEnabled}, []interface{}{nil, nil}},
			}),
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
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{contract.DeviceService{}, db.ErrNotFound}},
			}),
			http.StatusNotFound,
		},
		{"Device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{contract.DeviceService{}, testError}},
			}),
			http.StatusInternalServerError,
		},
		{"Update error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceService, map[string]string{NAME: testDeviceServiceName, ADMINSTATE: UNLOCKED}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceByName", []interface{}{testDeviceServiceName}, []interface{}{testDeviceService, nil}},
				{"UpdateDeviceService", []interface{}{adminStateEnabled}, []interface{}{testError, nil}},
			}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			restUpdateServiceAdminStateByName(
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

func TestUpdateServiceLastReportedById(t *testing.T) {

	tests := []struct {
		name           string
		request        *http.Request
		dbMock         interfaces.DBClient
		expectedStatus int
	}{
		{
			"ok",
			createDeviceServiceRequestWithBody(
				http.MethodPut, testDeviceServiceLastUpdated, map[string]string{ID: testDeviceServiceLastUpdatedId, LASTREPORTED: strconv.Itoa(lastReported)}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceLastUpdatedId}, []interface{}{testDeviceServiceLastUpdated, nil}},
				{"UpdateDeviceService", []interface{}{testDeviceServiceLastUpdated}, []interface{}{nil, nil}},
			}),
			http.StatusOK,
		},
		{
			"request validation error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceServiceLastUpdated, map[string]string{ID: testDeviceServiceLastUpdatedId, LASTREPORTED: ""}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceLastUpdatedId}, []interface{}{testDeviceServiceLastUpdated, testError}},
				{"UpdateDeviceService", []interface{}{testDeviceServiceLastUpdated}, []interface{}{nil, testError}},
			}),
			http.StatusBadRequest,
		},
		{
			"device service lookup error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceServiceLastUpdatedNotExist, map[string]string{ID: testDeviceServiceLastUpdatedNotExistId, LASTREPORTED: strconv.Itoa(lastReported)}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceLastUpdatedNotExistId}, []interface{}{contract.DeviceService{}, db.ErrNotFound}},
				{"UpdateDeviceService", []interface{}{testDeviceServiceLastUpdatedNotExist}, []interface{}{contract.DeviceService{}, db.ErrNotFound}},
			}),
			http.StatusNotFound,
		},
		{
			"update service last reported error",
			createDeviceServiceRequestWithBody(http.MethodPut, testDeviceServiceLastUpdated, map[string]string{ID: testDeviceServiceLastUpdatedId, LASTREPORTED: strconv.Itoa(lastReported)}),
			createMockWithOutlines([]mockOutline{
				{"GetDeviceServiceById", []interface{}{testDeviceServiceLastUpdatedId}, []interface{}{testDeviceServiceLastUpdated, nil}},
				{"UpdateDeviceService", []interface{}{testDeviceServiceLastUpdated}, []interface{}{testError}},
			}),
			http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			loggerMock := logger.NewMockClient()
			restUpdateServiceLastReportedById(
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

func createDeviceServiceRequest(httpMethod string, pathParamName string, pathParamValue string) *http.Request {
	req := httptest.NewRequest(httpMethod, testDeviceServiceURI, nil)
	return mux.SetURLVars(req, map[string]string{pathParamName: pathParamValue})
}

func createDeviceServiceRequestWithBody(
	httpMethod string,
	deviceService contract.DeviceService,
	pathParams map[string]string) *http.Request {

	// if your JSON marshalling fails you've got bigger problems
	body, _ := json.Marshal(deviceService)

	req := httptest.NewRequest(httpMethod, testDeviceServiceURI, bytes.NewReader(body))

	return mux.SetURLVars(req, pathParams)
}
