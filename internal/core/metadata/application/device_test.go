//
// Copyright (C) 2024-2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/config"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/utils"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/stretchr/testify/mock"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	profile             = "test-profile"
	notFountProfileName = "notFoundProfile"
	source1             = "source1"
	source2             = "resource2"
	command1            = "command1"
	command2            = "command2"
	deviceProfile       = models.DeviceProfile{
		Name:            profile,
		DeviceResources: []models.DeviceResource{{Name: source1}, {Name: source2}},
		DeviceCommands:  []models.DeviceCommand{{Name: command1}, {Name: command2}},
	}
	testDeviceServiceName = "TestDeviceServiceName"
)

func TestValidateParentProfileAndAutoEvents(t *testing.T) {
	dic := di.NewContainer(di.ServiceConstructorMap{})
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", profile).Return(deviceProfile, nil)
	dbClientMock.On("DeviceProfileByName", notFountProfileName).Return(models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "not found", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name          string
		device        models.Device
		errorExpected bool
	}{
		{"empty profile",
			models.Device{},
			false,
		},
		{"not found profile",
			models.Device{
				ProfileName: notFountProfileName,
			},
			true,
		},
		{"no auto events",
			models.Device{
				ProfileName: profile,
			},
			false,
		},
		{"resource exist",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: source1, Interval: "1s"}},
			},
			false,
		},
		{"command exist",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: source1, Interval: "1s"}, {SourceName: command1, Interval: "1s"}},
			},
			false,
		},
		{"resource not exist",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: "notFoundSource", Interval: "1s"}},
			},
			true,
		},
		{"interval format not valid",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: source1, Interval: "1"}},
			},
			true,
		},
		{"no profile",
			models.Device{
				AutoEvents: []models.AutoEvent{{SourceName: source1, Interval: "1s"}},
			},
			false,
		},
		{"resource match regex",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: "res.*", Interval: "1s"}},
			},
			false,
		},
		{"is own parent",
			models.Device{
				ProfileName: profile,
				Parent:      "me",
				Name:        "me",
			},
			true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateParentProfileAndAutoEvent(dic, testCase.device)
			if testCase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateParentProfileAndAutoEventsDropInvalid(t *testing.T) {
	dic := di.NewContainer(di.ServiceConstructorMap{})
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", profile).Return(deviceProfile, nil)
	dbClientMock.On("DeviceProfileByName", notFountProfileName).Return(models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "not found", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	tests := []struct {
		name               string
		device             models.Device
		expectedAutoEvents int
		errorExpected      bool
	}{
		{"empty profile",
			models.Device{},
			0,
			false,
		},
		{"not found profile",
			models.Device{
				ProfileName: notFountProfileName,
			},
			0,
			true,
		},
		{"no auto events",
			models.Device{
				ProfileName: profile,
			},
			0,
			false,
		},
		{"is own parent",
			models.Device{
				ProfileName: profile,
				Parent:      "me",
				Name:        "me",
				AutoEvents:  []models.AutoEvent{{SourceName: source1, Interval: "1s"}},
			},
			0,
			true,
		},
		{"all valid",
			models.Device{
				ProfileName: profile,
				AutoEvents:  []models.AutoEvent{{SourceName: source1, Interval: "1s"}},
			},
			1,
			false,
		},
		{"one invalid",
			models.Device{
				ProfileName: profile,
				AutoEvents: []models.AutoEvent{
					{SourceName: source1, Interval: "1s"},
					{SourceName: "invalid", Interval: "1s"},
				},
			},
			1,
			false,
		},
		{"all invalid",
			models.Device{
				ProfileName: profile,
				AutoEvents: []models.AutoEvent{
					{SourceName: "invalid1", Interval: "1s"},
					{SourceName: "invalid2", Interval: "1s"},
				},
			},
			0,
			false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			d, err := validateParentProfileAndAutoEventDropInvalid(dic, testCase.device)
			if testCase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedAutoEvents, len(d.AutoEvents))
			}
		})
	}
}

func TestForceAddDevice(t *testing.T) {
	invalidDeviceName := "invalidDevice"
	validDeviceName := "validDevice"
	invalidDeviceName2 := "invalidDevice2"
	invalidDevice := models.Device{Name: invalidDeviceName}
	invalidDevice2 := models.Device{Name: invalidDeviceName2}
	returnedDevice := models.Device{Name: validDeviceName}

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
			}
		},
	})

	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceByName", invalidDeviceName).Return(models.Device{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query", nil))
	dbClientMock.On("DeviceByName", validDeviceName).Return(returnedDevice, nil)
	dbClientMock.On("DeviceByName", invalidDeviceName2).Return(invalidDevice2, nil)
	dbClientMock.On("UpdateDevice", returnedDevice).Return(nil)
	dbClientMock.On("UpdateDevice", invalidDevice2).Return(errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to update", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name          string
		device        models.Device
		errorExpected bool
	}{
		{"invalid - DeviceByName error", invalidDevice, true},
		{"valid", returnedDevice, false},
		{"invalid - UpdateDevice error", invalidDevice2, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, _ := correlation.FromContextOrNew(context.Background())
			result, err := updateDevice(testCase.device, ctx, dic)
			if testCase.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, returnedDevice.Id, result)
			}
		})
	}
}

func newDeviceWriteDBClientMock(deviceName string) *mocks.DBClient {
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceProfileByName", profile).Return(deviceProfile, nil)
	dbClientMock.On("DeviceServiceNameExists", testDeviceServiceName).Return(true, nil)
	dbClientMock.On("DeviceNameExists", deviceName).Return(false, nil)
	dbClientMock.On("DeviceByName", deviceName).Return(models.Device{
		Name: deviceName, ServiceName: testDeviceServiceName, ProfileName: profile,
	}, nil)
	// No device or provision watcher is associated with the profile, so the deletion is allowed
	dbClientMock.On("ProvisionWatchersByProfileName", 0, 1, profile).Return([]models.ProvisionWatcher{}, nil)
	dbClientMock.On("DeleteDeviceProfileByName", profile).Return(nil)
	return dbClientMock
}

func newDeviceWriteDIC(dbClientMock *mocks.DBClient) *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{}
		},
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.ProfileAssignmentLockName: func(get di.Get) interface{} {
			return utils.NewProfileAssignmentLock()
		},
	})
}

// assertProfileDeletionBlocked asserts that the device profile deletion waits for the device write
// in progress, otherwise the deletion cannot see the device which is not in the database yet.
// The device write blocks in the DBClient mock until the returned channel is closed.
func assertProfileDeletionBlocked(t *testing.T, dic *di.Container, writing <-chan struct{}, finishWrite func(), deviceWrite func()) {
	written := make(chan struct{})
	go func() {
		defer close(written)
		deviceWrite()
	}()
	<-writing

	deleted := make(chan errors.EdgeX, 1)
	go func() {
		ctx, _ := correlation.FromContextOrNew(context.Background())
		deleted <- DeleteDeviceProfileByName(profile, ctx, dic)
	}()

	select {
	case <-deleted:
		t.Fatal("the device profile deletion should be blocked by the in-flight device write")
	case <-time.After(100 * time.Millisecond):
	}

	finishWrite()
	<-written

	select {
	case err := <-deleted:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("the device profile deletion was not resumed after the device write finished")
	}
}

func TestAddDeviceBlocksProfileDeletion(t *testing.T) {
	deviceName := "testDevice"
	device := models.Device{Name: deviceName, ServiceName: testDeviceServiceName, ProfileName: profile}

	writing, release := make(chan struct{}), make(chan struct{})
	dbClientMock := newDeviceWriteDBClientMock(deviceName)
	dbClientMock.On("DevicesByProfileName", 0, 1, profile).Return([]models.Device{}, nil)
	dbClientMock.On("AddDevice", mock.Anything).Return(models.Device{Name: deviceName}, nil).Once().
		Run(func(mock.Arguments) {
			close(writing)
			<-release
		})
	dic := newDeviceWriteDIC(dbClientMock)

	assertProfileDeletionBlocked(t, dic, writing, func() { close(release) }, func() {
		ctx, _ := correlation.FromContextOrNew(context.Background())
		_, err := AddDevice(device, ctx, dic, true, false)
		require.NoError(t, err)
	})
}

func TestPatchDeviceBlocksProfileDeletion(t *testing.T) {
	deviceName := "testDevice"
	dto := dtos.UpdateDevice{Name: &deviceName}

	writing, release := make(chan struct{}), make(chan struct{})
	dbClientMock := newDeviceWriteDBClientMock(deviceName)
	dbClientMock.On("DevicesByProfileName", 0, 1, profile).Return([]models.Device{}, nil)
	dbClientMock.On("UpdateDevice", mock.Anything).Return(nil).Once().
		Run(func(mock.Arguments) {
			close(writing)
			<-release
		})
	dic := newDeviceWriteDIC(dbClientMock)

	assertProfileDeletionBlocked(t, dic, writing, func() { close(release) }, func() {
		ctx, _ := correlation.FromContextOrNew(context.Background())
		require.NoError(t, PatchDevice(dto, ctx, dic, true))
	})
}

// The device profile deletion waits for every concurrent device write, releasing only some of them
// must not let the deletion through.
func TestConcurrentAddDeviceBlocksProfileDeletion(t *testing.T) {
	const deviceCount = 3
	dbClientMock := newDeviceWriteDBClientMock("testDevice")
	dbClientMock.On("DevicesByProfileName", 0, 1, profile).Return([]models.Device{}, nil)

	writing := make(chan struct{}, deviceCount)
	releases := make([]chan struct{}, deviceCount)
	for i := range releases {
		releases[i] = make(chan struct{})
		deviceName := fmt.Sprintf("testDevice-%d", i)
		dbClientMock.On("DeviceNameExists", deviceName).Return(false, nil)
		dbClientMock.On("AddDevice", mock.MatchedBy(func(d models.Device) bool { return d.Name == deviceName })).
			Return(models.Device{Name: deviceName}, nil).Once().
			Run(func(mock.Arguments) {
				writing <- struct{}{}
				<-releases[i]
			})
	}
	dic := newDeviceWriteDIC(dbClientMock)

	written := make(chan struct{})
	go func() {
		defer close(written)
		var wg sync.WaitGroup
		for i := range releases {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, _ := correlation.FromContextOrNew(context.Background())
				_, err := AddDevice(models.Device{
					Name: fmt.Sprintf("testDevice-%d", i), ServiceName: testDeviceServiceName, ProfileName: profile,
				}, ctx, dic, true, false)
				assert.NoError(t, err)
			}()
		}
		wg.Wait()
	}()
	for range releases {
		<-writing
	}

	deleted := make(chan errors.EdgeX, 1)
	go func() {
		ctx, _ := correlation.FromContextOrNew(context.Background())
		deleted <- DeleteDeviceProfileByName(profile, ctx, dic)
	}()

	// Release the writes one by one, the deletion must stay blocked until the last one completes
	for i, release := range releases {
		select {
		case <-deleted:
			t.Fatalf("the device profile deletion should still be blocked by %d in-flight device writes", deviceCount-i)
		case <-time.After(50 * time.Millisecond):
		}
		close(release)
	}
	<-written

	select {
	case err := <-deleted:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("the device profile deletion was not resumed after all the device writes finished")
	}
}

// The device write waits for the device profile deletion in progress, otherwise a device could be
// written between the association check and the deletion of the profile it references.
func TestProfileDeletionBlocksAddDevice(t *testing.T) {
	deviceName := "testDevice"
	device := models.Device{Name: deviceName, ServiceName: testDeviceServiceName, ProfileName: profile}

	deleting, release := make(chan struct{}), make(chan struct{})
	dbClientMock := newDeviceWriteDBClientMock(deviceName)
	// Block the deletion in the middle of its association check, before the profile is removed
	dbClientMock.On("DevicesByProfileName", 0, 1, profile).Return([]models.Device{}, nil).Once().
		Run(func(mock.Arguments) {
			close(deleting)
			<-release
		})
	dbClientMock.On("AddDevice", mock.Anything).Return(models.Device{Name: deviceName}, nil)
	dic := newDeviceWriteDIC(dbClientMock)

	deleted := make(chan errors.EdgeX, 1)
	go func() {
		ctx, _ := correlation.FromContextOrNew(context.Background())
		deleted <- DeleteDeviceProfileByName(profile, ctx, dic)
	}()
	<-deleting

	added := make(chan errors.EdgeX, 1)
	go func() {
		ctx, _ := correlation.FromContextOrNew(context.Background())
		_, err := AddDevice(device, ctx, dic, true, false)
		added <- err
	}()

	select {
	case <-added:
		t.Fatal("the device write should be blocked by the device profile deletion in progress")
	case <-time.After(100 * time.Millisecond):
	}
	dbClientMock.AssertNotCalled(t, "AddDevice", mock.Anything)

	close(release)
	require.NoError(t, <-deleted)

	select {
	case err := <-added:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("the device write was not resumed after the device profile deletion finished")
	}
}

func TestDeviceById(t *testing.T) {
	validId := "82eb2e26-0f24-48aa-ae4c-de9dac3fb9bc"
	notFoundId := "00000000-0000-0000-0000-000000000000"
	returnedDevice := models.Device{Id: validId, Name: "testDevice"}

	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})

	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceById", validId).Return(returnedDevice, nil)
	dbClientMock.On("DeviceById", notFoundId).Return(models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device doesn't exist in the database", nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	tests := []struct {
		name          string
		id            string
		errorExpected bool
	}{
		{"Valid - find device by id", validId, false},
		{"Invalid - empty id", "", true},
		{"Invalid - device not found by id", notFoundId, true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			device, err := DeviceById(testCase.id, dic)
			if testCase.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, validId, device.Id)
				assert.Equal(t, returnedDevice.Name, device.Name)
			}
		})
	}
}

func TestUpdateDevicePropagatesDeviceByNameError(t *testing.T) {
	missingName := "deviceDeletedBetweenChecks"
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					LogLevel: "DEBUG",
				},
			}
		},
	})
	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("DeviceByName", missingName).Return(models.Device{},
		errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device with name '%s' does not exist", missingName), nil))
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	ctx, _ := correlation.FromContextOrNew(context.Background())
	id, err := updateDevice(models.Device{Name: missingName}, ctx, dic)
	require.Error(t, err)
	require.Empty(t, id)
	// the force-add path must keep the kind and the reason of the lookup failure, otherwise the
	// caller of POST /device?force=true gets a message-less 500 for a device that no longer exists
	require.Equal(t, errors.KindEntityDoesNotExist, errors.Kind(err))
	require.Equal(t, http.StatusNotFound, err.Code())
	require.Contains(t, err.Message(), missingName)
}
