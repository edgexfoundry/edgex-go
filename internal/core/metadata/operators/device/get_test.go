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

package device

import (
	"github.com/pkg/errors"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestGetAllDevices(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.ServiceInfo
		dbMock      DeviceLoader
		expectError bool
	}{
		{"GetAllPass", config.ServiceInfo{MaxResultCount: 1}, createGetDeviceLoaderMock(), false},
		{"GetAllFailCount", config.ServiceInfo{}, createGetDeviceLoaderMock(), true},
		{"GetAllFailUnexpected", config.ServiceInfo{MaxResultCount: 1}, createGetDeviceLoaderMockFail(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewDeviceLoadAll(tt.cfg, tt.dbMock, logger.MockLogger{})
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

func createGetDeviceLoaderMock() DeviceLoader {
	devices := []contract.Device{testDevice}
	dbMock := &mocks.DeviceLoader{}
	dbMock.On("GetAllDevices").Return(devices, nil)
	return dbMock
}

func createGetDeviceLoaderMockFail() DeviceLoader {
	dbMock := &mocks.DeviceLoader{}
	dbMock.On("GetAllDevices").Return(nil, errors.New("unexpected error"))
	return dbMock
}
