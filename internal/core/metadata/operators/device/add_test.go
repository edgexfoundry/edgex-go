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
	"sync"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func TestAddNewDevice(t *testing.T) {
	tests := []struct {
		name        string
		dbMock      DeviceAdder
		expectError bool
	}{
		{"AddDeviceCheckByName", createAddDeviceDbMockForName(), false},
		{"AddDeviceCheckById", createAddDeviceDbMockForId(), false},
		{"AddDeviceFailNoDeviceService", createAddDeviceDbMockFailDeviceService(), true},
		{"AddDeviceFailNoDeviceProfile", createAddDeviceDbMockFailDeviceProfile(), true},
		{"AddDeviceFailDuplicateName", createAddDeviceDbMockDuplicateName(), true},
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
			}(&wg, t)

			op := NewAddDevice(ch, tt.dbMock, testDevice)
			newId, err := op.Execute()
			// Unexpected error, something weird happened
			if err != nil && !tt.expectError {
				t.Error(err)
				return
			}
			_, err = uuid.Parse(newId)
			// Returned value doesn't parse to a UUID
			if err != nil && !tt.expectError {
				t.Error(err)
				return
			}
			wg.Wait()
		})
	}
}

func createAddDeviceDbMockForName() DeviceAdder {
	dbMock := &mocks.DeviceAdder{}
	dbMock.On("AddDevice", testDevice).Return(uuid.New().String(), nil)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return dbMock
}

func createAddDeviceDbMockForId() DeviceAdder {
	dbMock := &mocks.DeviceAdder{}
	dbMock.On("AddDevice", testDevice).Return(uuid.New().String(), nil)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(models.DeviceProfile{}, db.ErrNotFound)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(models.DeviceService{}, db.ErrNotFound)
	return dbMock
}

func createAddDeviceDbMockFailDeviceService() DeviceAdder {
	dbMock := &mocks.DeviceAdder{}
	dbMock.On("AddDevice", testDevice).Return(uuid.New().String(), nil)
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(models.DeviceService{}, db.ErrNotFound)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(models.DeviceService{}, db.ErrNotFound)
	return dbMock
}

func createAddDeviceDbMockFailDeviceProfile() DeviceAdder {
	dbMock := &mocks.DeviceAdder{}
	dbMock.On("AddDevice", testDevice).Return(uuid.New().String(), nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	dbMock.On("GetDeviceProfileById", testDeviceProfileId).Return(models.DeviceProfile{}, db.ErrNotFound)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(models.DeviceProfile{}, db.ErrNotFound)
	return dbMock
}

func createAddDeviceDbMockDuplicateName() DeviceAdder {
	dbMock := &mocks.DeviceAdder{}
	dbMock.On("AddDevice", testDevice).Return("", db.ErrNotUnique)
	dbMock.On("GetDeviceProfileByName", testDeviceProfileName).Return(testDeviceProfile, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return dbMock
}

// Device Profile mock instance setup
var testExpectedvalues = []string{"temperature", "humidity"}
var testAction = models.Action{Path: "test/path", Responses: []models.Response{{Code: "200", Description: "ok", ExpectedValues: testExpectedvalues}}, URL: ""}

var testTimestamps = models.Timestamps{Created: 123, Modified: 123, Origin: 123}
var testDescribedObject = models.DescribedObject{Timestamps: testTimestamps, Description: "This is a description"}

var testUnits = models.Units{Type: "String", ReadWrite: "R", DefaultValue: "Degrees Fahrenheit"}
var testPropertyValue = models.PropertyValue{Type: "Float", ReadWrite: "RW", Minimum: "-99.99", Maximum: "199.99",
	DefaultValue: "0.00", Size: "8", Mask: "0x00", Shift: "0", Scale: "1.0", Offset: "0.0", Base: "0",
	Assertion: "0", Precision: "1", FloatEncoding: models.Base64Encoding, MediaType: clients.ContentTypeJSON}
var testProfileProperty = models.ProfileProperty{Value: testPropertyValue, Units: testUnits}
var testDeviceResource = models.DeviceResource{Description: "test device object description",
	Name: "test device object name", Tag: "test device object tag", Properties: testProfileProperty}

var testResourceOperation = models.ResourceOperation{Index: "test index", Operation: "test operation", Object: "test resource object",
	Parameter: "test parameter", Resource: "test resource", Secondary: []string{"test secondary"}, Mappings: make(map[string]string)}
var testProfileResource = models.ProfileResource{Name: "test profile resource name", Get: []models.ResourceOperation{testResourceOperation}, Set: []models.ResourceOperation{testResourceOperation}}

var testCommand = models.Command{Timestamps: testTimestamps, Name: "test command name", Get: models.Get{Action: testAction},
	Put: models.Put{Action: testAction, ParameterNames: testExpectedvalues}}

var testDeviceProfileId = uuid.New().String()
var testDeviceProfileName = "Test Profile.NAME"
var testDeviceProfile = models.DeviceProfile{Id: testDeviceProfileId, DescribedObject: testDescribedObject, Name: testDeviceProfileName,
	Manufacturer: "Test Manufacturer", Model: "Test Model", Labels: []string{"labe1", "label2"},
	DeviceResources: []models.DeviceResource{testDeviceResource}, DeviceCommands: []models.ProfileResource{testProfileResource},
	CoreCommands: []models.Command{testCommand}}

// Device Service mock instance setup
var testAddressable = models.Addressable{Timestamps: testTimestamps, Name: "TEST_ADDR.NAME", Protocol: "HTTP", HTTPMethod: "Get",
	Address: "localhost", Port: 48089, Path: clients.ApiDeviceRoute, Publisher: "TEST_PUB", User: "edgexer", Password: "password",
	Topic: "device_topic"}

var testDeviceServiceId = uuid.New().String()
var testDeviceServiceName = "test service"
var testDeviceService = models.DeviceService{Id: testDeviceServiceId, DescribedObject: testDescribedObject, Name: testDeviceServiceName,
	LastConnected: int64(1000000), LastReported: int64(1000000), OperatingState: "ENABLED", Labels: []string{"MODBUS", "TEMP"},
	Addressable: testAddressable, AdminState: "UNLOCKED"}

// Device mock instance setup
var testDeviceName = "test device name"
var testDevice = models.Device{DescribedObject: testDescribedObject, Name: testDeviceName, AdminState: "UNLOCKED",
	OperatingState: "ENABLED", Protocols: newTestProtocols(), LastReported: int64(1000000), LastConnected: int64(1000000),
	Labels: []string{"MODBUS", "TEMP"}, Location: "{40lat;45long}", Service: testDeviceService, Profile: testDeviceProfile,
	AutoEvents: newAutoEvent()}

func newAutoEvent() []models.AutoEvent {
	a := []models.AutoEvent{}
	a = append(a, models.AutoEvent{Resource: "TestDevice", Frequency: "300ms", OnChange: true})
	return a
}

func newTestProtocols() map[string]models.ProtocolProperties {
	p1 := make(models.ProtocolProperties)
	p1["host"] = "localhost"
	p1["port"] = "1234"
	p1["unitID"] = "1"

	p2 := make(models.ProtocolProperties)
	p2["serialPort"] = "/dev/USB0"
	p2["baudRate"] = "19200"
	p2["dataBits"] = "8"
	p2["stopBits"] = "1"
	p2["parity"] = "0"
	p2["unitID"] = "2"

	wrap := make(map[string]models.ProtocolProperties)
	wrap["modbus-ip"] = p1
	wrap["modbus-rtu"] = p2

	return wrap
}
