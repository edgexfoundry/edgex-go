//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/models"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testDeviceName   = "test-device"
	testResourceName = "test-resource"
	testSourceName   = "testSource"
	testDeviceInfo   = models.DeviceInfo{
		Id:           1,
		DeviceName:   testDeviceName,
		SourceName:   testSourceName,
		ResourceName: testResourceName,
		ValueType:    common.ValueTypeString,
	}
)

func mockDic() *di.Container {
	mockMetricsManager := &mocks.MetricsManager{}
	mockMetricsManager.On("Register", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMetricsManager.On("Unregister", mock.Anything)
	return di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.MetricsManagerInterfaceName: func(get di.Get) interface{} {
			return mockMetricsManager
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func TestDeviceInfoCache_GetDeviceInfoId(t *testing.T) {
	dic := mockDic()
	cache := NewDeviceInfoCache(dic, []models.DeviceInfo{testDeviceInfo})

	id, exist := cache.GetDeviceInfoId(testDeviceInfo)
	assert.Equal(t, testDeviceInfo.Id, id)
	assert.True(t, exist)

	invalidDeviceInfo := testDeviceInfo
	invalidDeviceInfo.Id = 2
	invalidDeviceInfo.SourceName = "invalid"
	_, exist = cache.GetDeviceInfoId(invalidDeviceInfo)
	assert.False(t, exist)
}

func TestDeviceInfoCache_GetDeviceInfoMap(t *testing.T) {
	dic := mockDic()
	cache := NewDeviceInfoCache(dic, []models.DeviceInfo{testDeviceInfo})

	result := cache.CloneDeviceInfoMapWithSourceName()
	assert.Equal(t, 1, len(result))
	assert.Equal(t, testDeviceInfo, result[1])
}

func TestDeviceInfoCache_Add(t *testing.T) {
	dic := mockDic()
	cache := NewDeviceInfoCache(dic, []models.DeviceInfo{})

	cache.Add(testDeviceInfo)
	result := cache.CloneDeviceInfoMapWithSourceName()
	assert.Equal(t, 1, len(result))
	assert.Equal(t, testDeviceInfo, result[1])
}

func TestDeviceInfoCache_Remove(t *testing.T) {
	dic := mockDic()

	testDeviceInfo2 := testDeviceInfo
	testDeviceInfo2.Id = 2
	testDeviceInfo2.DeviceName = "testDevice2"
	cache := NewDeviceInfoCache(dic, []models.DeviceInfo{testDeviceInfo, testDeviceInfo2})

	cache.Remove(testDeviceInfo)
	result := cache.CloneDeviceInfoMapWithSourceName()
	assert.Equal(t, 1, len(result))
	_, exists := cache.GetDeviceInfoId(testDeviceInfo)
	assert.False(t, exists)

	cache.Remove(testDeviceInfo2)
	result = cache.CloneDeviceInfoMapWithSourceName()
	assert.Equal(t, 0, len(result))
	_, exists = cache.GetDeviceInfoId(testDeviceInfo)
	assert.False(t, exists)
}
