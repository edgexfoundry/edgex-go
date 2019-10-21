package command

import (
	"context"
	goErrors "errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	mdMocks "github.com/edgexfoundry/go-mod-core-contracts/clients/metadata/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/command/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/command/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

//commandByDeviceID
func TestExecuteGETCommandByDeviceIdAndCommandId(t *testing.T) {

	LoggingClient = logger.NewMockClient()
	mdc = newMockDeviceClient()
	dbClient = newCommandMock()

	tests := []struct {
		name        string
		deviceId    string
		commandId   string
		expectedErr error
	}{
		{
			"Device-InvalidObjectId",
			status400,
			TestCommandId,
			types.NewErrServiceClient(400, []byte("Invalid object ID")),
		},
		{
			"Device-NotFound",
			status404,
			TestCommandId,
			types.NewErrServiceClient(http.StatusNotFound, []byte{}),
		},
		{
			"Device-InternalServerErr",
			status500,
			TestCommandId,
			goErrors.New("unexpected error"),
		},
		{
			"Device-Locked",
			status423,
			TestCommandId,
			errors.NewErrDeviceLocked(""),
		},
		{
			"Command-NotFound",
			d200c404,
			status400,
			db.ErrNotFound,
		},
		{
			"Command-InternalServerErr",
			d200c500,
			status500,
			goErrors.New("unexpected error"),
		},
		{
			"Command-NotBelongToDevice",
			mismatch,
			mismatch,
			errors.NewErrCommandNotAssociatedWithDevice(TestCommandId, mismatch),
		},
		{
			"Command-500URLCannotBeParsed",
			"d200c200",
			"200",
			&url.Error{Op: "parse", URL: "://:0?", Err: goErrors.New("missing protocol scheme")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, actualErr := commandByDeviceID(tt.deviceId, TestCommandId, "", "", false, context.Background())
			if actualErr == nil {
				t.Fatal("expected error")
			}
			if tt.expectedErr.Error() != actualErr.Error() {
				t.Fatalf("error value mismatch -- expected %v got %v", tt.expectedErr, actualErr)
			}
		})
	}
}

func newMockDeviceClient() *mdMocks.DeviceClient {
	client := mdMocks.DeviceClient{}
	client.On("Device", status404, context.Background()).Return(contract.Device{}, types.NewErrServiceClient(http.StatusNotFound, []byte{}))
	client.On("Device", status400, context.Background()).Return(contract.Device{}, types.NewErrServiceClient(400, []byte("Invalid object ID")))
	client.On("Device", status500, context.Background()).Return(contract.Device{}, goErrors.New("unexpected error"))
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
	dbMock.On("GetCommandsByDeviceId", status500).Return(nil, goErrors.New("unexpected error"))
	dbMock.On("GetCommandsByDeviceId", mismatch).Return([]models.Command{contract.Command{Id: "dummy"}}, nil)

	dbMock.On("GetCommandsByDeviceId", TestDeviceId).Return([]models.Command{contract.Command{Id: TestCommandId}}, nil)

	return dbMock
}

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
