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
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestNewPutCommandWithCorrelationId(t *testing.T) {
	expectedCorrelationIDHeaderValue := "Testing"
	testContext := context.WithValue(context.Background(), clients.CorrelationHeader, expectedCorrelationIDHeaderValue)
	putCommand, _ := NewPutCommand(testDevice, testCommand, "Test body", testContext, nil, logger.NewMockClient())
	actualCorrelationIDHeaderValue := putCommand.(serviceCommand).Request.Header.Get(clients.CorrelationHeader)
	if actualCorrelationIDHeaderValue == "" {
		t.Errorf("The populated PutCommand's request should contain a correlation ID header value")
	}

	if actualCorrelationIDHeaderValue != expectedCorrelationIDHeaderValue {
		t.Errorf("The populated PutCommand's request should contain the correct correlation ID")
	}
}
func TestNewPutCommandNoCorrelationIDInContext(t *testing.T) {
	putCommand, _ := NewPutCommand(testDevice, testCommand, "Test Body", context.Background(), nil, logger.NewMockClient())
	actualCorrelationIDHeaderValue := putCommand.(serviceCommand).Request.Header.Get(clients.CorrelationHeader)
	if actualCorrelationIDHeaderValue != "" {
		t.Errorf("No correlation ID should be specified")
	}
}

func TestNewPutCommandInvalidBaseUrl(t *testing.T) {
	device := testDevice
	device.Service.Addressable.Address = "!@#$"

	_, err := NewPutCommand(device, testCommand, "Test body", context.Background(), nil, logger.NewMockClient())
	if err != nil {
		t.Errorf("The invalid URL error was not properly propegated to the caller")
	}
}

func TestNewPutCommandBody(t *testing.T) {
	expectedRequestBody := "Test Request Body"
	expectedRequestBodySize := len(expectedRequestBody)
	putCommand, err := NewPutCommand(testDevice, testCommand, expectedRequestBody, context.Background(), nil, logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpectedly failed while creating a PutCommand")
	}

	actualBodyBytes, _ := ioutil.ReadAll(putCommand.(serviceCommand).Body)
	if expectedRequestBodySize != len(actualBodyBytes) {
		t.Errorf("Failed to verify the request body size")
	}

	actualRequestBody := string(actualBodyBytes)
	if expectedRequestBody != actualRequestBody {
		t.Error("Failed to verify the request body contents")
	}
}
