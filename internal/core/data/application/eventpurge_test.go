//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"testing"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/core/data/cache"
	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/infrastructure/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/core/data/mocks"
	pkgCache "github.com/edgexfoundry/edgex-go/internal/pkg/cache"
	dbModels "github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/models"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	lcMocks "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewEventPurge(t *testing.T) {
	deviceInfoMap := map[int]dbModels.DeviceInfo{1: {DeviceName: testDeviceName}}
	ep := newEventPurgeExecutor(deviceInfoMap)
	assert.Equal(t, testDeviceName, ep.deviceInfoMap[1].DeviceName)
}

func TestPurgeMatchingDeviceInfo(t *testing.T) {
	lc := logger.NewMockClient()
	tests := []struct {
		name                string
		deviceName          string
		sourceName          string
		expectedResultCount int
	}{
		{"Test cullMatchingDeviceInfo using device/source defined by matched regular expressions", "^Test+", "^Source+", 0},
		{"Test cullMatchingDeviceInfo using device/source defined by unmatched regular expressions", "^Not+", "^Not+", 1},
		{"Test cullMatchingDeviceInfo using device/source defined by exact matched names", "Test1", "Source1", 0},
		{"Test cullMatchingDeviceInfo using device/source defined by unmatched names", "Test2", "Source2", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deviceInfoMap := map[int]dbModels.DeviceInfo{1: {DeviceName: "Test1", SourceName: "Source1"}}
			ep := newEventPurgeExecutor(deviceInfoMap)
			ep.cullMatchingDeviceInfo(tt.deviceName, tt.sourceName, lc)
			assert.Equal(t, tt.expectedResultCount, len(ep.deviceInfoMap))
		})
	}

}

func TestPurgeEventsByAutoEvent(t *testing.T) {
	dic := mocks.NewMockDIC()
	deviceName := "testDevice"
	sourceName := "testSource"
	coreDataConfig := container.ConfigurationFrom(dic.Get)
	coreDataConfig.Retention = config.EventRetention{
		Interval:        "10m",
		DefaultMaxCap:   -1,
		DefaultMinCap:   1,
		DefaultDuration: "30m",
	}
	//readingCache := cache.NewReadingCache(dic)

	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return coreDataConfig
		},
		//container.ReadingCacheInterfaceName: func(get di.Get) interface{} { return readingCache },
	})

	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("LatestEventByDeviceNameAndSourceNameAndAgeAndOffset", deviceName, sourceName, mock.Anything, mock.Anything).Return(models.Event{}, nil)
	dbClientMock.On("DeleteEventsByAgeAndDeviceNameAndSourceName", mock.Anything, deviceName, sourceName).Return(nil)
	dbClientMock.On("DeleteEventsByDeviceNameAndSourceName", deviceName, sourceName).Return(nil)
	dbClientMock.On("LatestEventByDeviceNameAndSourceNameAndOffset", deviceName, sourceName, mock.Anything).Return(models.Event{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	deviceInfoMap := map[int]dbModels.DeviceInfo{}
	ep := newEventPurgeExecutor(deviceInfoMap)

	tests := []struct {
		name      string
		autoEvent models.AutoEvent
	}{
		{"time-based event retention",
			models.AutoEvent{
				Interval:          "",
				OnChange:          false,
				OnChangeThreshold: 0,
				SourceName:        sourceName,
				Retention: models.Retention{
					MaxCap:   -1,
					MinCap:   -1,
					Duration: "10m",
				},
			},
		},
		{"time-based event retention with miniCap",
			models.AutoEvent{
				Interval:          "",
				OnChange:          false,
				OnChangeThreshold: 0,
				SourceName:        sourceName,
				Retention: models.Retention{
					MaxCap:   -1,
					MinCap:   1,
					Duration: "10m",
				},
			},
		},
		{"count-based event retention",
			models.AutoEvent{
				Interval:          "",
				OnChange:          false,
				OnChangeThreshold: 0,
				SourceName:        sourceName,
				Retention: models.Retention{
					MaxCap:   -1,
					MinCap:   -1,
					Duration: "0s",
				},
			},
		},
		{"count-based event retention with miniCap",
			models.AutoEvent{
				Interval:          "",
				OnChange:          false,
				OnChangeThreshold: 0,
				SourceName:        sourceName,
				Retention: models.Retention{
					MaxCap:   -1,
					MinCap:   1,
					Duration: "0s",
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			deviceStore := pkgCache.DeviceStore(dic)
			device := models.Device{
				Name:       deviceName,
				AutoEvents: []models.AutoEvent{testCase.autoEvent},
			}
			deviceStore.Add(device)
			dic.Update(di.ServiceConstructorMap{
				container.DeviceStoreInterfaceName: func(get di.Get) interface{} {
					return deviceStore
				},
			})

			err := ep.purgeEventsByAutoEvent(dic)
			require.NoError(t, err)
			duration, parseErr := time.ParseDuration(testCase.autoEvent.Retention.Duration)
			require.NoError(t, parseErr)
			if duration > 0 && testCase.autoEvent.Retention.MinCap <= 0 {
				// time-based retention
				dbClientMock.AssertCalled(t, "DeleteEventsByAgeAndDeviceNameAndSourceName", mock.Anything, deviceName, testCase.autoEvent.SourceName)
			} else if duration > 0 && testCase.autoEvent.Retention.MinCap > 0 {
				// time-based retention with miniCap
				dbClientMock.AssertCalled(t, "LatestEventByDeviceNameAndSourceNameAndAgeAndOffset", deviceName, sourceName, duration.Nanoseconds(), testCase.autoEvent.Retention.MinCap)
				dbClientMock.AssertCalled(t, "DeleteEventsByAgeAndDeviceNameAndSourceName", mock.Anything, deviceName, testCase.autoEvent.SourceName)
			} else if testCase.autoEvent.Retention.MinCap <= 0 {
				// count-based retention
				dbClientMock.AssertCalled(t, "DeleteEventsByDeviceNameAndSourceName", deviceName, testCase.autoEvent.SourceName)
			} else {
				// count-based retention with miniCap
				dbClientMock.AssertCalled(t, "LatestEventByDeviceNameAndSourceNameAndOffset", deviceName, sourceName, testCase.autoEvent.Retention.MinCap)
				dbClientMock.AssertCalled(t, "DeleteEventsByAgeAndDeviceNameAndSourceName", mock.Anything, deviceName, testCase.autoEvent.SourceName)
			}

		})
	}
}

func TestPurgeEventsByDeviceInfo(t *testing.T) {
	testResourceName := "test-resource"
	testDeviceInfo := dbModels.DeviceInfo{
		Id:           1,
		DeviceName:   testDeviceName,
		SourceName:   testSourceName,
		ResourceName: testResourceName,
		ValueType:    common.ValueTypeString,
	}

	failedDeviceName := "failedDevice"
	failedDeviceInfo := testDeviceInfo
	failedDeviceInfo.DeviceName = failedDeviceName
	testDuration := "3d10h"

	isValid, duration := common.ParseDurationWithDay(testDuration)
	assert.True(t, isValid)

	dic := di.NewContainer(di.ServiceConstructorMap{})

	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteEventsByAgeAndDeviceNameAndSourceName", duration.Nanoseconds(), testDeviceName, testSourceName).Return(nil)
	dbClientMock.On("DeleteEventsByAgeAndDeviceNameAndSourceName", duration.Nanoseconds(), failedDeviceName, testSourceName).Return(errors.NewCommonEdgeX(errors.KindServerError, "failed to delete event with device name", nil))
	dbClientMock.On("EventCountByDeviceNameAndSourceNameAndLimit", testDeviceName, testSourceName, 1000).Return(int64(0), errors.NewCommonEdgeX(errors.KindServerError, "failed to get event count", nil))
	dbClientMock.On("EventCountByDeviceNameAndSourceNameAndLimit", failedDeviceName, testSourceName, 1000).Return(int64(0), errors.NewCommonEdgeX(errors.KindServerError, "failed to get event count", nil))

	mockLogger := &lcMocks.LoggingClient{}
	mockLogger.On("Errorf", "failed to execute event retention for device info with device name: '%s', source name: '%s', error: %v", failedDeviceName, testSourceName, mock.Anything)
	mockLogger.On("Errorf", "failed to execute event retention for device info with device name: '%s', source name: '%s', config DefaultMinCap: '%d', config DefaultDuration: %s, error: %v",
		failedDeviceName, testSourceName, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockLogger.On("Debugf", "Starting execute event retention for device name: '%s', source name: '%s', maxCap: %d, minCap: %d, duration: %s", testDeviceName, testSourceName, int64(-1), int64(0), testDuration).Return(nil)
	mockLogger.On("Debugf", "Starting execute event retention for device name: '%s', source name: '%s', maxCap: %d, minCap: %d, duration: %s", failedDeviceName, testSourceName, int64(-1), int64(0), testDuration).Return(nil)
	mockLogger.On("Debugf", "MinCap is disabled, purge events by duration '%d' and deviceName '%s', and sourceName '%s'", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Debug", "Starting purge events by retention policies on default config level ......").Return(nil)

	validConfig := config.EventRetention{
		Interval:        "10m",
		DefaultMaxCap:   -1,
		DefaultMinCap:   0,
		DefaultDuration: testDuration,
	}
	invalidDurationConfig := validConfig
	invalidDurationConfig.DefaultDuration = "abc"
	maxCapConfig := validConfig
	maxCapConfig.DefaultMaxCap = 1000

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return mockLogger
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					PersistData: true,
				},
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 20,
				},
			}
		},
	})

	tests := []struct {
		name                string
		deviceInfos         []dbModels.DeviceInfo
		retentionConfig     config.EventRetention
		expectParseError    bool
		expectLogError      bool
		expectLogErrorCount int
	}{
		{"Valid - successfully purge events by device info", []dbModels.DeviceInfo{testDeviceInfo}, validConfig, false, false, 0},
		{"Valid - no device info", []dbModels.DeviceInfo{}, validConfig, false, false, 0},
		{"Invalid - invalid duration", []dbModels.DeviceInfo{failedDeviceInfo}, invalidDurationConfig, false, true, 1},
		{"Invalid - failed to delete events with maxCap", []dbModels.DeviceInfo{failedDeviceInfo}, maxCapConfig, false, true, 2},
		{"Invalid - failed to delete events with device name", []dbModels.DeviceInfo{failedDeviceInfo}, validConfig, false, true, 3},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			deviceInfoCache := cache.NewDeviceInfoCache(dic, testCase.deviceInfos)
			coreDataConfig := container.ConfigurationFrom(dic.Get)
			coreDataConfig.Retention = testCase.retentionConfig
			dic.Update(di.ServiceConstructorMap{
				container.DeviceInfoCacheInterfaceName: func(get di.Get) interface{} {
					return deviceInfoCache
				},
				container.ConfigurationName: func(get di.Get) interface{} { return coreDataConfig },
			})

			ep := newEventPurgeExecutor(deviceInfoCache.CloneDeviceInfoMapWithSourceName())
			err := ep.purgeEventsByDeviceInfo(dic)

			if testCase.expectParseError {
				assert.Error(t, err, "expected error when purge events by device info")
			} else {
				assert.NoError(t, err, "expected no error when purge events by device info")
			}
			if testCase.expectLogError {
				mockLogger.AssertNumberOfCalls(t, "Errorf", testCase.expectLogErrorCount)
			}
		})
	}
}

func TestHandlePurgeEvents(t *testing.T) {
	testResourceName := "test-resource"
	testDeviceInfo := dbModels.DeviceInfo{
		Id:           1,
		DeviceName:   testDeviceName,
		SourceName:   testSourceName,
		ResourceName: testResourceName,
		ValueType:    common.ValueTypeString,
	}

	testDuration := "3d10h"

	isValid, duration := common.ParseDurationWithDay(testDuration)
	assert.True(t, isValid)

	dic := di.NewContainer(di.ServiceConstructorMap{})

	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteEventsByAgeAndDeviceNameAndSourceName", duration.Nanoseconds(), testDeviceName, testSourceName).Return(nil)
	dbClientMock.On("EventCountByDeviceNameAndSourceNameAndLimit", testDeviceName, testSourceName, 1000).Return(int64(0), errors.NewCommonEdgeX(errors.KindServerError, "failed to get event count", nil))

	mockLogger := &lcMocks.LoggingClient{}
	mockLogger.On("Debugf", "Starting execute event retention for device name: '%s', source name: '%s', maxCap: %d, minCap: %d, duration: %s", testDeviceName, testSourceName, int64(0), int64(0), testDuration).Return(nil)
	mockLogger.On("Debugf", "MinCap is disabled, purge events by duration '%d' and deviceName '%s', and sourceName '%s'", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	validConfig := config.EventRetention{
		Interval:        "10m",
		DefaultMaxCap:   -1,
		DefaultMinCap:   0,
		DefaultDuration: testDuration,
	}

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return mockLogger
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Writable: config.WritableInfo{
					PersistData: true,
				},
				Service: bootstrapConfig.ServiceInfo{
					MaxResultCount: 20,
				},
			}
		},
	})

	deviceInfoCache := cache.NewDeviceInfoCache(dic, []dbModels.DeviceInfo{testDeviceInfo})
	coreDataConfig := container.ConfigurationFrom(dic.Get)
	coreDataConfig.Retention = validConfig
	dic.Update(di.ServiceConstructorMap{
		container.DeviceInfoCacheInterfaceName: func(get di.Get) interface{} {
			return deviceInfoCache
		},
		container.ConfigurationName: func(get di.Get) interface{} { return coreDataConfig },
	})

	err := handleEventRetention(testDeviceName, testSourceName, 0, 0, testDuration, dic)
	assert.NoError(t, err, "expected no error when purge events by device info")
}
