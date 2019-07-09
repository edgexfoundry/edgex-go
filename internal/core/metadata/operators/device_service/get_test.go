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

package device_service

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device_service/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func TestGetAllDeviceServices(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.ServiceInfo
		dbMock      DeviceServiceLoader
		expectError bool
	}{
		{"GetAllPass", config.ServiceInfo{MaxResultCount: 1}, createGetDeviceServiceLoaderMock(), false},
		{"GetAllFailCount", config.ServiceInfo{}, createGetDeviceServiceLoaderMock(), true},
		{"GetAllFailUnexpected", config.ServiceInfo{MaxResultCount: 1}, createGetDeviceServiceLoaderMockFail(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewDeviceServiceLoadAll(tt.cfg, tt.dbMock, logger.MockLogger{})
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

func createGetDeviceServiceLoaderMock() DeviceServiceLoader {
	services := []contract.DeviceService{testDeviceService}
	dbMock := &mocks.DeviceServiceLoader{}
	dbMock.On("GetAllDeviceServices").Return(services, nil)
	return dbMock
}

func createGetDeviceServiceLoaderMockFail() DeviceServiceLoader {
	dbMock := &mocks.DeviceServiceLoader{}
	dbMock.On("GetAllDeviceServices").Return(nil, errors.New("unexpected error"))
	return dbMock
}

var testDeviceServiceId = uuid.New().String()
var testDeviceServiceName = "test service"
var testDeviceService = contract.DeviceService{Id: testDeviceServiceId, Name: testDeviceServiceName}
