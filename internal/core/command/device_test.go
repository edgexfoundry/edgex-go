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
	"errors"
	"net/http"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	mdMocks "github.com/edgexfoundry/go-mod-core-contracts/clients/metadata/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

var (
	status400 = "400"
	status404 = "404"
	status500 = "500"
	status423 = "423"

	mismatch = "d200-c200-mismatch"
	d200c404 = "d200-c404"
	d200c500 = "d200-c500"

	TestCommandId = "TestCommandID"
)

// commandByDeviceID
func TestExecuteGETCommandByDeviceIdAndCommandId(t *testing.T) {
	mdc = newMockDeviceClient()
	dbClient = newCommandMock()

	tests := []struct {
		name           string
		deviceId       string
		commandId      string
		expectedStatus int
	}{
		{
			"Device-InvalidObjectId",
			status400,
			TestCommandId,
			http.StatusBadRequest,
		},
		{
			"Device-NotFound",
			status404,
			TestCommandId,
			http.StatusNotFound,
		},
		{
			"Device-InternalServerErr",
			status500,
			TestCommandId,
			http.StatusInternalServerError,
		},
		{
			"Device-Locked",
			status423,
			TestCommandId,
			http.StatusLocked,
		},
		{
			"Command-NotFound",
			d200c404,
			status400,
			http.StatusNotFound,
		},
		{
			"Command-InternalServerErr",
			d200c500,
			status500,
			http.StatusInternalServerError,
		},
		{
			"Command-NotBelongToDevice",
			mismatch,
			mismatch,
			http.StatusNotFound,
		},
		{
			"Command-500URLCannotBeParsed",
			"d200c200",
			"200",
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, statusCode := commandByDeviceID(tt.deviceId, TestCommandId, "", "", false, context.Background(),
				logger.NewMockClient())
			if tt.expectedStatus != statusCode {
				t.Errorf("status code mismatch -- expected %v got %v", tt.expectedStatus, statusCode)
				return
			}
		})
	}
}

func newMockDeviceClient() *mdMocks.DeviceClient {
	client := mdMocks.DeviceClient{}
	client.On("Device", status404, context.Background()).Return(contract.Device{}, types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("Device", status400, context.Background()).Return(contract.Device{}, types.NewErrServiceClient(400, []byte("Invalid object ID")))
	client.On("Device", status500, context.Background()).Return(contract.Device{}, errors.New("unexpected error"))
	client.On("Device", status423, context.Background()).Return(contract.Device{AdminState: "LOCKED"}, nil)
	client.On("Device", d200c404, context.Background()).Return(contract.Device{Id: status400}, nil)
	client.On("Device", d200c500, context.Background()).Return(contract.Device{Id: status500}, nil)
	client.On("Device", mismatch, context.Background()).Return(contract.Device{Id: mismatch}, nil)

	client.On("Device", "d200c200", context.Background()).Return(contract.Device{Id: TestDeviceId}, nil)
	return &client
}

func newCommandMock() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", status400).Return(nil, db.ErrNotFound)
	dbMock.On("GetCommandsByDeviceId", status500).Return(nil, errors.New("unexpected error"))
	dbMock.On("GetCommandsByDeviceId", mismatch).Return([]models.Command{contract.Command{Id: "dummy"}}, nil)

	dbMock.On("GetCommandsByDeviceId", TestDeviceId).Return([]models.Command{contract.Command{Id: TestCommandId}}, nil)

	return dbMock
}
