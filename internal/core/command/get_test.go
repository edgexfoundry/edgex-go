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
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	TestProtocol        = "http"
	TestDeviceId        = "TestDeviceID"
	TestAddress         = "example.com"
	TestPort            = 8080
	NonPropagatedHeader = "NonPropagatedHeader"
)

// Device which can be used as a basis for test setup. By default this is constructed for happy path testing.
var testDevice = contract.Device{
	Id:         TestDeviceId,
	AdminState: contract.Unlocked,
	Service: contract.DeviceService{
		Addressable: contract.Addressable{
			Protocol: TestProtocol,
			Address:  TestAddress,
			Port:     TestPort,
		},
	},
}

// Command which can be used as a basis for test setup. By default this is constructed for happy path testing.
var testCommand = contract.Command{
	Get: contract.Get{
		Action: contract.Action{
			Path: "/some/uri",
		},
	},
	Put: contract.Put{
		Action: contract.Action{
			Path: "/another/uri",
		},
	},
}

func TestNewGetCommandWithCorrelationId(t *testing.T) {
	expectedCorrelationIDHeaderValue := "Testing"
	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeCBOR}, http.MethodGet)
	testContext := context.WithValue(context.Background(), clients.CorrelationHeader, expectedCorrelationIDHeaderValue)
	getCommand, _ := NewGetCommand(testDevice, testCommand, testContext, nil, logger.NewMockClient(), req)
	actualCorrelationIDHeaderValue := getCommand.(serviceCommand).Request.Header.Get(clients.CorrelationHeader)

	if actualCorrelationIDHeaderValue == "" {
		t.Errorf("The populated GetCommand's request should contain a correlation ID header value")
	}

	if actualCorrelationIDHeaderValue != expectedCorrelationIDHeaderValue {
		t.Errorf("The populated GetCommand's request should contain the correct correlation ID")
	}
}

func TestNewGetCommandWithQueryParams(t *testing.T) {
	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeCBOR}, http.MethodGet)
	queryParams := "test=value1&test2=value2"
	req.URL.RawQuery = queryParams
	getCommand, _ := NewGetCommand(testDevice, testCommand, context.Background(), nil, logger.NewMockClient(), req)
	r := getCommand.(serviceCommand).Request.URL
	if r.Scheme != TestProtocol {
		t.Errorf("Unexpected protocol")
	}
	expectedHost := TestAddress + ":" + strconv.Itoa(TestPort)
	if r.Host != expectedHost {
		t.Errorf("Unexpected host address and port")
	}
	if r.Path != testCommand.Get.Action.Path {
		t.Errorf("Unexpected path")
	}
	if r.RawQuery != queryParams {
		t.Errorf("Unexpected Raw Query Value")
	}
}
func TestNewGetCommandWithMalformedQueryParams(t *testing.T) {
	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeJSON}, http.MethodGet)
	queryParams := "!@#$%"
	req.URL.RawQuery = queryParams
	_, err := NewGetCommand(testDevice, testCommand, context.Background(), nil, logger.NewMockClient(), req)
	if err == nil {
		t.Errorf("Expected error for malformed query parameters")
	}
}
func TestNewGetCommandNoCorrelationIDInContext(t *testing.T) {
	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeJSON}, http.MethodGet)
	getCommand, _ := NewGetCommand(testDevice, testCommand, context.Background(), nil, logger.NewMockClient(), req)
	actualCorrelationIDHeaderValue := getCommand.(serviceCommand).Request.Header.Get(clients.CorrelationHeader)
	if actualCorrelationIDHeaderValue != "" {
		t.Errorf("No correlation ID should be specified")
	}
}

func TestNewGetCommandInvalidBaseUrl(t *testing.T) {
	device := testDevice
	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeCBOR}, http.MethodGet)
	device.Service.Addressable.Address = "!@#$"
	_, err := NewGetCommand(device, testCommand, context.Background(), nil, logger.NewMockClient(), req)
	if err == nil {
		t.Errorf("The invalid URL error was not properly propagated to the caller")
	}
}

func TestNewGetCommandContentType(t *testing.T) {
	tests := []struct {
		name            string
		originalHeaders map[string]string
		expectedHeaders map[string]string
	}{
		{
			name:            "cbor content type header propagated",
			originalHeaders: map[string]string{clients.ContentType: clients.ContentTypeCBOR},
			expectedHeaders: map[string]string{clients.ContentType: clients.ContentTypeCBOR},
		},
		{
			name:            "json content type header propagated",
			originalHeaders: map[string]string{clients.ContentType: clients.ContentTypeJSON},
			expectedHeaders: map[string]string{clients.ContentType: clients.ContentTypeJSON},
		},
		{
			name:            "no content type header provided",
			originalHeaders: map[string]string{clients.ContentType: ""},
			expectedHeaders: map[string]string{},
		},
		{
			name:            "cbor content type propagated, random header not propagated",
			originalHeaders: map[string]string{clients.ContentType: clients.ContentTypeCBOR, NonPropagatedHeader: "NonPropagatedHeader"},
			expectedHeaders: map[string]string{clients.ContentType: clients.ContentTypeCBOR},
		},
		{
			name:            "json content type propagated, random header not propagated",
			originalHeaders: map[string]string{clients.ContentType: clients.ContentTypeJSON, NonPropagatedHeader: "NonPropagatedHeader"},
			expectedHeaders: map[string]string{clients.ContentType: clients.ContentTypeJSON},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loggerMock = logger.NewMockClient()
			ctx := context.Background()
			proxiedRequest := newRequestWithHeaders(tt.originalHeaders, http.MethodGet)
			getCommand, _ := NewGetCommand(
				testDevice,
				testCommand,
				ctx,
				nil,
				loggerMock,
				proxiedRequest)
			actualHeaders := map[string]string{}
			for headerName, headerValues := range getCommand.(serviceCommand).Request.Header {
				// Extract the first element only from slice.
				actualHeaders[headerName] = headerValues[0]
			}
			if !reflect.DeepEqual(actualHeaders, tt.expectedHeaders) {
				t.Errorf("expected %s does not match the observed %s", tt.expectedHeaders, actualHeaders)
			}
		})
	}
}
