//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/core/command/config"
	"github.com/edgexfoundry/edgex-go/internal/core/command/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtosCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
)

const (
	testDeviceName = "testDeviceName"
	testServiceUrl = "http://localhost:59882"
	resource1      = "str1_R"
	resource2      = "str2_W"
	resource3      = "str3_RW"
	resource4      = "str4_RW"
	resource5      = "str5_RW"
	resource6      = "str6_RW"
	command1       = "cmd1"
	command2       = "cmd2"
)

func TestBuildCoreCommands(t *testing.T) {
	profile := dtos.DeviceProfile{
		DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{Name: "testProfile"},
		DeviceResources: []dtos.DeviceResource{
			{Name: resource1, Properties: dtos.ResourceProperties{ValueType: common.ValueTypeString, ReadWrite: common.ReadWrite_R}},
			{Name: resource2, Properties: dtos.ResourceProperties{ValueType: common.ValueTypeInt16, ReadWrite: common.ReadWrite_W}},
			{Name: resource3, Properties: dtos.ResourceProperties{ValueType: common.ValueTypeBool, ReadWrite: common.ReadWrite_RW}},
			{Name: resource4, Properties: dtos.ResourceProperties{ValueType: common.ValueTypeString, ReadWrite: common.ReadWrite_RW}, IsHidden: true},
			{Name: resource5, Properties: dtos.ResourceProperties{ValueType: common.ValueTypeInt16, ReadWrite: common.ReadWrite_RW}},
			{Name: resource6, Properties: dtos.ResourceProperties{ValueType: common.ValueTypeBool, ReadWrite: common.ReadWrite_RW}},
		},
		DeviceCommands: []dtos.DeviceCommand{
			{
				Name: command1, ReadWrite: common.ReadWrite_R,
				ResourceOperations: []dtos.ResourceOperation{
					{DeviceResource: resource1}, {DeviceResource: resource2}, {DeviceResource: resource3},
				},
			},
			{
				Name: command2, ReadWrite: common.ReadWrite_W, IsHidden: true,
				ResourceOperations: []dtos.ResourceOperation{
					{DeviceResource: resource4}, {DeviceResource: resource5},
				},
			},
			{
				Name: resource6, ReadWrite: common.ReadWrite_RW,
				ResourceOperations: []dtos.ResourceOperation{
					{DeviceResource: resource6},
				},
			},
		},
	}
	expectedCoreCommand := []dtos.CoreCommand{
		{
			Name: command1, Get: true, Path: commandPath(testDeviceName, command1), Url: testServiceUrl,
			Parameters: []dtos.CoreCommandParameter{
				{ResourceName: resource1, ValueType: common.ValueTypeString},
				{ResourceName: resource2, ValueType: common.ValueTypeInt16},
				{ResourceName: resource3, ValueType: common.ValueTypeBool},
			},
		},
		{
			Name: resource6, Get: true, Set: true, Path: commandPath(testDeviceName, resource6), Url: testServiceUrl,
			Parameters: []dtos.CoreCommandParameter{{ResourceName: resource6, ValueType: common.ValueTypeBool}},
		},
		{
			Name: resource1, Get: true, Path: commandPath(testDeviceName, resource1), Url: testServiceUrl,
			Parameters: []dtos.CoreCommandParameter{{ResourceName: resource1, ValueType: common.ValueTypeString}},
		},
		{
			Name: resource2, Set: true, Path: commandPath(testDeviceName, resource2), Url: testServiceUrl,
			Parameters: []dtos.CoreCommandParameter{{ResourceName: resource2, ValueType: common.ValueTypeInt16}},
		},
		{
			Name: resource3, Get: true, Set: true, Path: commandPath(testDeviceName, resource3), Url: testServiceUrl,
			Parameters: []dtos.CoreCommandParameter{{ResourceName: resource3, ValueType: common.ValueTypeBool}},
		},
		{
			Name: resource5, Get: true, Set: true, Path: commandPath(testDeviceName, resource5), Url: testServiceUrl,
			Parameters: []dtos.CoreCommandParameter{{ResourceName: resource5, ValueType: common.ValueTypeInt16}},
		},
	}

	result, err := buildCoreCommands(testDeviceName, testServiceUrl, profile)
	require.NoError(t, err)

	assert.ElementsMatch(t, expectedCoreCommand, result)
}

func TestAllCommands(t *testing.T) {
	ctx := context.Background()
	device1 := dtos.Device{Name: "test-device-1", ProfileName: "test-profile-1"}
	device2 := dtos.Device{Name: "test-device-2", ProfileName: "test-profile-2"}
	device3 := dtos.Device{Name: "test-device-3", ProfileName: ""}

	dic := di.NewContainer(di.ServiceConstructorMap{})
	dc := &mocks.DeviceClient{}
	dpc := &mocks.DeviceProfileClient{}
	dpc.On("DeviceProfileByName", ctx, device1.ProfileName).Return(responses.DeviceProfileResponse{}, nil)
	dpc.On("DeviceProfileByName", ctx, device2.ProfileName).Return(responses.DeviceProfileResponse{}, nil)
	dic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				Service: bootstrapConfig.ServiceInfo{
					Host: "localhost",
					Port: 59882,
				},
			}
		},
		bootstrapContainer.DeviceClientName: func(get di.Get) interface{} {
			return dc
		},
		bootstrapContainer.DeviceProfileClientName: func(get di.Get) interface{} {
			return dpc
		},
	})

	tests := []struct {
		name                 string
		multiDevicesResponse responses.MultiDevicesResponse
		expectedTotalCount   uint32
	}{
		{
			name: "query the commands",
			multiDevicesResponse: responses.MultiDevicesResponse{
				BaseWithTotalCountResponse: dtosCommon.BaseWithTotalCountResponse{TotalCount: 2},
				Devices:                    []dtos.Device{device1, device2},
			},
			expectedTotalCount: 2,
		},
		{
			name: "device has empty profile",
			multiDevicesResponse: responses.MultiDevicesResponse{
				BaseWithTotalCountResponse: dtosCommon.BaseWithTotalCountResponse{TotalCount: 2},
				Devices:                    []dtos.Device{device1, device3},
			},
			expectedTotalCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var labels []string
			dc.On("AllDevices", ctx, labels, 0, 0).Return(tt.multiDevicesResponse, nil).Once()
			_, count, err := AllCommands(0, 0, dic)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedTotalCount, count)
		})
	}
}
