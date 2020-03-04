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
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestNewPutCommandWithCorrelationId(t *testing.T) {
	expectedCorrelationIDHeaderValue := "Testing"
	testContext := context.WithValue(context.Background(), clients.CorrelationHeader, expectedCorrelationIDHeaderValue)
	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeCBOR}, http.MethodPut)
	putCommand, _ := NewPutCommand(testDevice, testCommand, "Test body", testContext, nil, logger.NewMockClient(), req)
	actualCorrelationIDHeaderValue := putCommand.(serviceCommand).Request.Header.Get(clients.CorrelationHeader)
	if actualCorrelationIDHeaderValue == "" {
		t.Errorf("The populated PutCommand's request should contain a correlation ID header value")
	}

	if actualCorrelationIDHeaderValue != expectedCorrelationIDHeaderValue {
		t.Errorf("The populated PutCommand's request should contain the correct correlation ID")
	}
}

func TestNewPutCommandNoCorrelationIDInContext(t *testing.T) {
	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeJSON}, http.MethodPut)
	putCommand, _ := NewPutCommand(testDevice, testCommand, "Test Body", context.Background(), nil, logger.NewMockClient(), req)
	actualCorrelationIDHeaderValue := putCommand.(serviceCommand).Request.Header.Get(clients.CorrelationHeader)
	if actualCorrelationIDHeaderValue != "" {
		t.Errorf("No correlation ID should be specified")
	}
}

func TestNewPutCommandInvalidBaseUrl(t *testing.T) {
	device := testDevice
	device.Service.Addressable.Address = "!@#$"

	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeCBOR}, http.MethodPut)
	_, err := NewPutCommand(device, testCommand, "Test body", context.Background(), nil, logger.NewMockClient(), req)
	if err != nil {
		t.Errorf("The invalid URL error was not properly propagated to the caller")
	}
}

func TestNewPutCommandBody(t *testing.T) {
	expectedRequestBody := "Test Request Body"
	req := newRequestWithHeaders(map[string]string{clients.ContentType: clients.ContentTypeJSON}, http.MethodPut)
	putCommand, err := NewPutCommand(testDevice, testCommand, expectedRequestBody, context.Background(), nil, logger.NewMockClient(), req)

	if err != nil {
		t.Errorf("Unexpectedly failed while creating a PutCommand")
	}

	expectedRequestBodySize := len(expectedRequestBody)
	actualBodyBytes, _ := ioutil.ReadAll(putCommand.(serviceCommand).Body)
	if expectedRequestBodySize != len(actualBodyBytes) {
		t.Errorf("Failed to verify the request body size")
	}

	actualRequestBody := string(actualBodyBytes)
	if expectedRequestBody != actualRequestBody {
		t.Error("Failed to verify the request body contents")
	}
}

func TestNewPutCommandContentType(t *testing.T) {
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
			putCommand, _ := NewPutCommand(
				testDevice,
				testCommand,
				"Test Body",
				ctx,
				nil,
				loggerMock,
				proxiedRequest)
			actualHeaders := map[string]string{}
			for headerName, headerValues := range putCommand.(serviceCommand).Request.Header {
				// Extract the first element only from slice.
				actualHeaders[headerName] = headerValues[0]
			}
			if !reflect.DeepEqual(actualHeaders, tt.expectedHeaders) {
				t.Errorf("expected %s does not match the observed %s", tt.expectedHeaders, actualHeaders)
			}
		})
	}
}
