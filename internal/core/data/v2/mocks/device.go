//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"fmt"
	"net/http"

	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/mocks"

	"github.com/edgexfoundry/go-mod-bootstrap/di"

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

	mockDeviceResultFn := func(ctx context.Context, id string) contract.Device {
		if bson.IsObjectIdHex(id) {
			return contract.Device{Id: id, Name: testDeviceName, Protocols: protocols}
		}
		return contract.Device{}
	}
	client.On("Device", context.Background(), "valid").Return(mockDeviceResultFn, nil)
	client.On("Device", context.Background(), "404").Return(mockDeviceResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("Device", context.Background(), mock.Anything).Return(mockDeviceResultFn, fmt.Errorf("some error"))

	mockDeviceForNameResultFn := func(ctx context.Context, name string) contract.Device {
		device := contract.Device{Id: uuid.New().String(), Name: name, Protocols: protocols}

		return device
	}
	client.On("DeviceForName", context.Background(), testDeviceName).Return(mockDeviceForNameResultFn, nil)
	client.On("DeviceForName", context.Background(), "404").Return(mockDeviceForNameResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("DeviceForName", context.Background(), mock.Anything).Return(mockDeviceForNameResultFn,
		fmt.Errorf("some error"))

	mockCheckForDeviceResultFn := func(ctx context.Context, name string) contract.Device {
		device := contract.Device{Id: uuid.New().String(), Name: name, Protocols: protocols}

		return device
	}
	client.On("CheckForDevice", context.Background(), testDeviceName).Return(mockCheckForDeviceResultFn, nil)
	client.On("CheckForDevice", context.Background(), "404").Return(mockCheckForDeviceResultFn,
		types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("CheckForDevice", context.Background(), mock.Anything).Return(mockCheckForDeviceResultFn,
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

func MetadataDeviceClientFrom(get di.Get) *mocks.DeviceClient {
	return get(v2DataContainer.MetadataDeviceClientName).(*mocks.DeviceClient)
}
