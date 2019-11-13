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
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestUpdateDevice(t *testing.T) {
	tests := []struct {
		name        string
		dbMock      DeviceUpdater
		expectError bool
	}{
		{"UpdateDevice", createUpdateDeviceDbMock(), false},
		{"UpdateDeviceError", createUpdateDeviceErrorDbMock(), true},
		{"UpdateDeviceByName", createUpdateDeviceByNameDbMock(), false},
		{"UpdateDeviceNotFound", createUpdateDeviceNotFoundDbMock(), true},
		{"UpdateDeviceServiceByName", createUpdateDeviceByServiceNameDbMock(), false},
		{"UpdateDeviceServiceNotFound", createUpdateDeviceServiceNotFoundDbMock(), true},
		{"UpdateDeviceProfileNotFound", createUpdateDeviceDeviceProfileNotFoundDbMock(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup

			ch := make(chan DeviceEvent)
			defer close(ch)

			wg.Add(1)
			go func(wg *sync.WaitGroup, t *testing.T) {
				defer wg.Done()
				msg, ok := <-ch
				if !ok {
					t.Errorf("%s unsuccessful read from channel", t.Name())
					return
				}
				if msg.Error != nil && !tt.expectError {
					t.Errorf("%s error reported via channel: %s", t.Name(), msg.Error.Error())
					return
				}
				// Ensure that all successful operations result in the correct action.
				if !tt.expectError && msg.HttpMethod != http.MethodPut {
					t.Errorf("Expected HTTP method 'PUT', but recieved '%s'", msg.HttpMethod)
				}
			}(&wg, t)

			op := NewUpdateDevice(ch, tt.dbMock, testDevice, logger.MockLogger{})
			err := op.Execute()

			if err != nil && !tt.expectError {
				t.Error(err)
				return
			}

			if err == nil && tt.expectError {
				t.Errorf("Expected an error but didn't receive one.")
				return
			}

			if !tt.expectError {
				// assert that a call to dbClient.UpdateDevice() was made
				tt.dbMock.(*mocks.DeviceUpdater).AssertCalled(t, "UpdateDevice", testDevice)
			}

			wg.Wait()
		})
	}
}

func createUpdateDeviceDbMock() DeviceUpdater {
	var dbMock mocks.DeviceUpdater
	dbMock.On("UpdateDevice", testDevice).Return(nil)
	dbMock.On("GetDeviceById", testDevice.Id).Return(testDevice, nil)
	dbMock.On("GetDeviceByName", testDevice.Name).Return(testDevice, nil)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return &dbMock
}

func createUpdateDeviceErrorDbMock() DeviceUpdater {
	var dbMock mocks.DeviceUpdater
	dbMock.On("UpdateDevice", testDevice).Return(db.ErrInvalidObjectId)
	dbMock.On("GetDeviceById", testDevice.Id).Return(testDevice, nil)
	dbMock.On("GetDeviceByName", testDevice.Name).Return(testDevice, nil)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return &dbMock
}

func createUpdateDeviceByNameDbMock() DeviceUpdater {
	var dbMock mocks.DeviceUpdater
	dbMock.On("UpdateDevice", testDevice).Return(nil)
	dbMock.On("GetDeviceById", testDevice.Id).Return(testDevice, fmt.Errorf("err"))
	dbMock.On("GetDeviceByName", testDevice.Name).Return(testDevice, nil)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return &dbMock
}

func createUpdateDeviceNotFoundDbMock() DeviceUpdater {
	var dbMock mocks.DeviceUpdater
	dbMock.On("UpdateDevice", testDevice).Return(nil)
	dbMock.On("GetDeviceById", testDevice.Id).Return(models.Device{}, db.ErrNotFound)
	dbMock.On("GetDeviceByName", testDevice.Name).Return(models.Device{}, db.ErrNotFound)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return &dbMock
}

func createUpdateDeviceByServiceNameDbMock() DeviceUpdater {
	var dbMock mocks.DeviceUpdater
	dbMock.On("UpdateDevice", testDevice).Return(nil)
	dbMock.On("GetDeviceById", testDevice.Id).Return(testDevice, nil)
	dbMock.On("GetDeviceByName", testDevice.Name).Return(testDevice, nil)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, db.ErrNotFound)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return &dbMock
}

func createUpdateDeviceServiceNotFoundDbMock() DeviceUpdater {
	var dbMock mocks.DeviceUpdater
	dbMock.On("UpdateDevice", testDevice).Return(nil)
	dbMock.On("GetDeviceById", testDevice.Id).Return(testDevice, nil)
	dbMock.On("GetDeviceByName", testDevice.Name).Return(testDevice, nil)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, db.ErrNotFound)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, db.ErrNotFound)
	return &dbMock
}

func createUpdateDeviceDeviceProfileNotFoundDbMock() DeviceUpdater {
	var dbMock mocks.DeviceUpdater
	dbMock.On("UpdateDevice", testDevice).Return(nil)
	dbMock.On("GetDeviceById", testDevice.Id).Return(testDevice, nil)
	dbMock.On("GetDeviceByName", testDevice.Name).Return(testDevice, nil)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(testDeviceProfile, db.ErrNotFound)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, db.ErrNotFound)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return &dbMock
}
