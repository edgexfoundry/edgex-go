package device

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
)

func TestNotification(t *testing.T) {
	pass := DeviceEvent{DeviceId: uuid.New().String(), DeviceName: testDeviceName, HttpMethod: http.MethodPost, ServiceId: testDeviceServiceId}
	fail := DeviceEvent{Error: db.ErrNotFound}
	failInvalidMethod := pass
	failInvalidMethod.HttpMethod = "{}"

	tests := []struct {
		name        string
		dbMock      DeviceServiceLoader
		event       DeviceEvent
		expectError bool
	}{
		{"DeviceAdded", createNotifyDeviceServiceDbMock(), pass, false},
		{"ErrorOccurred", createNotifyDeviceServiceDbMock(), fail, true},
		{"NoDeviceService", createNotifyDeviceServiceDbMockFail(), pass, true},
		{"NoAddressable", createNotifyEmptyAddressableDbMockFail(), pass, true},
		{"InvalidRequest", createNotifyDeviceServiceDbMock(), failInvalidMethod, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			nc := mockNotificationClientOK{}
			requester, err := NewRequester(Mock, logger.MockLogger{}, context.Background())
			if err != nil {
				t.Error(err.Error())
				return
			}

			ch := make(chan DeviceEvent)
			defer close(ch)

			wg.Add(1)
			op := NewNotifier(ch, nc, testNotificationInfo, tt.dbMock, requester, newMockNotifyLogger(tt.expectError, t), context.Background())
			go func(wg *sync.WaitGroup) {
				op.Execute()
				wg.Done()
			}(&wg)

			ch <- tt.event
			wg.Wait()
		})
	}
}

func createNotifyDeviceServiceDbMock() DeviceServiceLoader {
	dbMock := &mocks.DeviceServiceLoader{}
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(testDeviceService, nil)
	dbMock.On("GetDeviceServiceByName", testDeviceServiceName).Return(testDeviceService, nil)
	return dbMock
}

func createNotifyDeviceServiceDbMockFail() DeviceServiceLoader {
	dbMock := &mocks.DeviceServiceLoader{}
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(models.DeviceService{}, db.ErrNotFound)
	return dbMock
}

func createNotifyEmptyAddressableDbMockFail() DeviceServiceLoader {
	ds := testDeviceService
	ds.Addressable = models.Addressable{}
	dbMock := &mocks.DeviceServiceLoader{}
	dbMock.On("GetDeviceServiceById", testDeviceServiceId).Return(ds, nil)
	return dbMock
}

var testNotificationInfo = config.NotificationInfo{PostDeviceChanges: true, Slug: "device-change-", Content: "Device update: ", Sender: "core-metadata",
	Description: "Metadata device notice", Label: "metadata"}

/*
  I really struggled with the question of whether to create these structs or whether to use Mockery as in the case of
  the new DB interfaces. Using mockery would require that I create a new interfaces in either this package or the
  metadata/interfaces package. The new interface would be local to metadata but have the same signature as the current
  NotificationsClient. I decided to create the types below because it's less of a conceptual shift but also because the
  interface in question has only one method. If it were a more complicated interface, thus requiring more flexible
  mocking of behavior, I may have chosen to define a local interface.
*/
type mockNotificationClientOK struct{}

func (m mockNotificationClientOK) SendNotification(n notifications.Notification, ctx context.Context) error {
	return nil
}

type mockNotificationClientFail struct{}

func (m mockNotificationClientFail) SendNotification(n notifications.Notification, ctx context.Context) error {
	return errors.NewErrBadRequest("simulated bad request 400 response")
}

// The NotifyLogger mock is used to register unexpected errors with the testing framework.
// See the implementation of the Error() method.
type mockNotifyLogger struct {
	expectError bool
	t           *testing.T
}

func newMockNotifyLogger(expectError bool, t *testing.T) logger.LoggingClient {
	return mockNotifyLogger{expectError: expectError, t: t}
}

// SetLogLevel simulates setting a log severity level
func (lc mockNotifyLogger) SetLogLevel(loglevel string) error {
	return nil
}

// Info simulates logging an entry at the INFO severity level
func (lc mockNotifyLogger) Info(msg string, args ...interface{}) {
}

// Debug simulates logging an entry at the DEBUG severity level
func (lc mockNotifyLogger) Debug(msg string, args ...interface{}) {
}

// Error simulates logging an entry at the ERROR severity level
func (lc mockNotifyLogger) Error(msg string, args ...interface{}) {
	if !lc.expectError {
		lc.t.Error(msg)
		lc.t.Fail()
	}
}

// Trace simulates logging an entry at the TRACE severity level
func (lc mockNotifyLogger) Trace(msg string, args ...interface{}) {
}

// Warn simulates logging an entry at the WARN severity level
func (lc mockNotifyLogger) Warn(msg string, args ...interface{}) {
}
