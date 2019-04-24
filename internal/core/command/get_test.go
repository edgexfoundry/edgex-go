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
	"context"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"testing"
)

const (
	TestProtocol = "http"
	TestDeviceId = "TestDeviceID"
	TestAddress  = "example.com"
	TestPort     = 8080
)

// Device which can be used as a basis for test setup. By default this is constructed for happy path testing.
var testDevice = contract.Device{
	Id:         TestDeviceId,
	AdminState: contract.Unlocked,
	Service: contract.DeviceService{
		Service: contract.Service{
			Addressable: contract.Addressable{
				Protocol: TestProtocol,
				Address:  TestAddress,
				Port:     TestPort,
			},
		},
	},
}

// Command which can be used as a basis for test setup. By default this is constructed for happy path testing.
var testCommand = contract.Command{
	Get: &contract.Get{
		Action: contract.Action{
			Path: "/some/uri",
		},
	},
	Put: &contract.Put{
		Action: contract.Action{
			Path: "/another/uri",
		},
	},
}

func TestNewGetCommandWithCorrelationId(t *testing.T) {
	expectedCorrelationIDHeaderValue := "Testing"
	testContext := context.WithValue(context.Background(), clients.CorrelationHeader, expectedCorrelationIDHeaderValue)
	getCommand, _ := NewGetCommand(testDevice, testCommand, testContext, nil)
	actualCorrelationIDHeaderValue := getCommand.(serviceCommand).Request.Header.Get(clients.CorrelationHeader)
	if actualCorrelationIDHeaderValue == "" {
		t.Errorf("The populated GetCommand's request should contain a correlation ID header value")
	}

	if actualCorrelationIDHeaderValue != expectedCorrelationIDHeaderValue {
		t.Errorf("The populated GetCommand's request should contain the correct correlation ID")
	}
}

func TestNewGetCommandNoCorrelationIDInContext(t *testing.T) {
	getCommand, _ := NewGetCommand(testDevice, testCommand, context.Background(), nil)
	actualCorrelationIDHeaderValue := getCommand.(serviceCommand).Request.Header.Get(clients.CorrelationHeader)
	if actualCorrelationIDHeaderValue != "" {
		t.Errorf("No correlation ID should be specified")
	}
}

func TestNewGetCommandInvalidBaseUrl(t *testing.T) {
	device := testDevice
	device.Service.Addressable.Address = "!@#$"
	_, err := NewGetCommand(device, testCommand, context.Background(), nil)
	if err != nil {
		t.Errorf("The invalid URL error was not properly propegated to the caller")
	}
}
