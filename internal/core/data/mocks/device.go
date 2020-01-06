/********************************************************************************
 *  Copyright 2019 Dell Inc.
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

package mocks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

const testDeviceName string = "Test Device"

// NewMockDeviceClient creates a new mock DeviceClient which has some general mocking behavior defined.
func NewMockDeviceClient() *mocks.DeviceClient {
	client := &mocks.DeviceClient{}

	protocols := getProtocols()

	mockDeviceResultFn := func(id string, ctx context.Context) contract.Device {
		if bson.IsObjectIdHex(id) {
			return contract.Device{Id: id, Name: testDeviceName, Protocols: protocols}
		}
		return contract.Device{}
	}
	client.On("Device", "valid", context.Background()).Return(mockDeviceResultFn, nil)
	client.On("Device", "404", context.Background()).Return(mockDeviceResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("Device", mock.Anything, context.Background()).Return(mockDeviceResultFn, fmt.Errorf("some error"))

	mockDeviceForNameResultFn := func(name string, ctx context.Context) contract.Device {
		device := contract.Device{Id: uuid.New().String(), Name: name, Protocols: protocols}

		return device
	}
	client.On("DeviceForName", testDeviceName, context.Background()).Return(mockDeviceForNameResultFn, nil)
	client.On("DeviceForName", "404", context.Background()).Return(mockDeviceForNameResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("DeviceForName", mock.Anything, context.Background()).Return(mockDeviceForNameResultFn,
		fmt.Errorf("some error"))

	return client
}

func getProtocols() map[string]contract.ProtocolProperties {
	p1 := make(map[string]string)
	p1["host"] = "localhost"
	p1["port"] = "1234"
	p1["unitID"] = "1"

	p2 := make(map[string]string)
	p2["serialPort"] = "/dev/USB0"
	p2["baudRate"] = "19200"
	p2["dataBits"] = "8"
	p2["stopBits"] = "1"
	p2["parity"] = "0"
	p2["unitID"] = "2"

	wrap := make(map[string]contract.ProtocolProperties)
	wrap["modbus-ip"] = p1
	wrap["modbus-rtu"] = p2

	return wrap
}
