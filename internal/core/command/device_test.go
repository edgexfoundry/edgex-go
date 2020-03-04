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
	goErrors "errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/command/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces/mocks"
	mdMocks "github.com/edgexfoundry/edgex-go/internal/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	ExistingDeviceID                       = "existing device id"
	DeviceIDWithAssociatedInvalidObjectID  = "device id with associated invalid object id"
	NonExistentDeviceID                    = "non existent device id"
	DeviceIDResultingInInternalServerError = "device id resulting in internal server error"
	DeviceIDForLockedResource              = "device id for locked resource"
	MismatchedDeviceID                     = "mismatched device id"
	DeviceIDd200c404                       = "device id d200-c404"
	DeviceIDd200c500                       = "device id d200-c500"
	DeviceIDd200c200                       = "device id d200-c200"
	TestCommandID                          = "test command id"
	TestDeviceID                           = "test device id"
)

func createTestDeviceWithPathUrl(commandID string, deviceID string) models.Device {

	return models.Device{
		AdminState: models.Unlocked,
		Service: models.DeviceService{
			Addressable: models.Addressable{
				Path: "/api/v1/device/" + deviceID + "/command/" + commandID}}}
}

func createDeviceRequestWithPathParameters(
	sampleDevice models.Device,
	params map[string]string) *http.Request {

	req, _ := http.NewRequest(http.MethodGet, cmdURI, nil)
	req.URL.Path = sampleDevice.Service.Addressable.Path

	return mux.SetURLVars(req, params)
}

// commandByDeviceID
func TestExecuteGETCommandByDeviceIDAndCommandID(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		request     *http.Request
		expectedErr error
	}{
		{
			"device with invalid object id",
			DeviceIDWithAssociatedInvalidObjectID,
			createDeviceRequestWithPathParameters(
				createTestDeviceWithPathUrl(TestCommandID, DeviceIDWithAssociatedInvalidObjectID),
				map[string]string{ID: DeviceIDWithAssociatedInvalidObjectID, COMMANDID: TestCommandID}),
			types.NewErrServiceClient(400, []byte("Invalid object ID")),
		},
		{
			"device not found",
			NonExistentDeviceID,
			createDeviceRequestWithPathParameters(
				createTestDeviceWithPathUrl(TestCommandID, NonExistentDeviceID),
				map[string]string{ID: NonExistentDeviceID, COMMANDID: TestCommandID}),
			types.NewErrServiceClient(http.StatusNotFound, []byte{}),
		},
		{
			"device resulting in internal server error",
			DeviceIDResultingInInternalServerError,
			createDeviceRequestWithPathParameters(
				createTestDeviceWithPathUrl(TestCommandID, DeviceIDResultingInInternalServerError),
				map[string]string{ID: DeviceIDResultingInInternalServerError, COMMANDID: TestCommandID}),
			goErrors.New("unexpected error"),
		},
		{
			"device was locked",
			DeviceIDForLockedResource,
			createDeviceRequestWithPathParameters(
				createTestDeviceWithPathUrl(TestCommandID, DeviceIDForLockedResource),
				map[string]string{ID: DeviceIDForLockedResource, COMMANDID: TestCommandID}),
			errors.NewErrDeviceLocked(""),
		},
		{
			"command was not found",
			DeviceIDd200c404,
			createDeviceRequestWithPathParameters(
				createTestDeviceWithPathUrl(TestCommandID, DeviceIDd200c404),
				map[string]string{ID: DeviceIDd200c404, COMMANDID: DeviceIDWithAssociatedInvalidObjectID}),
			db.ErrNotFound,
		},
		{
			"command could not be handled",
			DeviceIDd200c500,
			createDeviceRequestWithPathParameters(
				createTestDeviceWithPathUrl(TestCommandID, DeviceIDd200c500),
				map[string]string{ID: DeviceIDd200c500, COMMANDID: DeviceIDResultingInInternalServerError}),
			goErrors.New("unexpected error"),
		},
		{
			"command did not belong to device",
			MismatchedDeviceID,
			createDeviceRequestWithPathParameters(
				createTestDeviceWithPathUrl(TestCommandID, MismatchedDeviceID),
				map[string]string{ID: MismatchedDeviceID, COMMANDID: TestCommandID}),
			errors.NewErrCommandNotAssociatedWithDevice(TestCommandID, MismatchedDeviceID),
		}, {
			"command could not be parsed",
			DeviceIDd200c200,
			createDeviceRequestWithPathParameters(
				createTestDeviceWithPathUrl(ExistingDeviceID, DeviceIDd200c200),
				map[string]string{ID: DeviceIDd200c200, COMMANDID: ExistingDeviceID}),
			&url.Error{Op: "parse", URL: "://:0?", Err: goErrors.New("missing protocol scheme")},
		},
	}
	httpCaller := createMockHttpCaller()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, actualErr := executeCommandByDeviceID(
				tt.request,
				"",
				logger.NewMockClient(),
				newMockDBClient(),
				newMockDeviceClient(),
				httpCaller)
			if actualErr == nil {
				t.Fatal("expected error")
			}
			require.Equal(t, tt.expectedErr.Error(), actualErr.Error())
		})
	}
}

func newMockDeviceClient() *mdMocks.DeviceClient {
	client := mdMocks.DeviceClient{}
	client.On("Device", mock.Anything, DeviceIDWithAssociatedInvalidObjectID).Return(contract.Device{}, types.NewErrServiceClient(400, []byte("Invalid object ID")))
	client.On("Device", mock.Anything, NonExistentDeviceID).Return(contract.Device{}, types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("Device", mock.Anything, DeviceIDResultingInInternalServerError).Return(contract.Device{}, goErrors.New("unexpected error"))
	client.On("Device", mock.Anything, DeviceIDForLockedResource).Return(contract.Device{Id: DeviceIDForLockedResource, AdminState: "LOCKED"}, nil)
	client.On("Device", mock.Anything, DeviceIDd200c404).Return(contract.Device{Id: DeviceIDWithAssociatedInvalidObjectID}, nil)
	client.On("Device", mock.Anything, DeviceIDd200c500).Return(contract.Device{Id: DeviceIDResultingInInternalServerError}, nil)
	client.On("Device", mock.Anything, MismatchedDeviceID).Return(contract.Device{Id: MismatchedDeviceID}, nil)
	client.On("Device", mock.Anything, TestDeviceID).Return(contract.Device{Id: TestDeviceID}, nil)
	client.On("Device", mock.Anything, DeviceIDd200c200).Return(contract.Device{Id: DeviceIDd200c200}, nil)
	return &client
}

func newMockDBClient() interfaces.DBClient {
	dbMock := &mocks.DBClient{}
	dbMock.On("GetCommandsByDeviceId", DeviceIDWithAssociatedInvalidObjectID).Return(nil, db.ErrNotFound)
	dbMock.On("GetCommandsByDeviceId", DeviceIDResultingInInternalServerError).Return(nil, goErrors.New("unexpected error"))
	dbMock.On("GetCommandsByDeviceId", MismatchedDeviceID).Return([]models.Command{{Id: "dummy"}}, nil)
	dbMock.On("GetCommandsByDeviceId", TestDeviceID).Return([]models.Command{{Id: TestCommandID}}, nil)
	dbMock.On("GetCommandsByDeviceId", ExistingDeviceID).Return([]models.Command{{Id: ExistingDeviceID}}, nil)
	dbMock.On("GetCommandsByDeviceId", DeviceIDd200c200).Return([]models.Command{{Id: ExistingDeviceID}}, nil)
	return dbMock
}
