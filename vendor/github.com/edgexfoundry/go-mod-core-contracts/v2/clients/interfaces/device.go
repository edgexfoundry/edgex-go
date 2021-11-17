//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// DeviceClient defines the interface for interactions with the Device endpoint on the EdgeX Foundry core-metadata service.
type DeviceClient interface {
	// Add adds new devices.
	Add(ctx context.Context, reqs []requests.AddDeviceRequest) ([]common.BaseWithIdResponse, errors.EdgeX)
	// Update updates devices.
	Update(ctx context.Context, reqs []requests.UpdateDeviceRequest) ([]common.BaseResponse, errors.EdgeX)
	// AllDevices returns all devices. Devices can also be filtered by labels.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	AllDevices(ctx context.Context, labels []string, offset int, limit int) (responses.MultiDevicesResponse, errors.EdgeX)
	// DeviceNameExists checks whether the device exists.
	DeviceNameExists(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
	// DeviceByName returns a device by device name.
	DeviceByName(ctx context.Context, name string) (responses.DeviceResponse, errors.EdgeX)
	// DeleteByName deletes a device by device name.
	DeleteDeviceByName(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
	// DevicesByProfileName returns devices associated with the specified device profile.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	DevicesByProfileName(ctx context.Context, name string, offset int, limit int) (responses.MultiDevicesResponse, errors.EdgeX)
	// DevicesByServiceName returns devices associated with the specified device service.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	DevicesByServiceName(ctx context.Context, name string, offset int, limit int) (responses.MultiDevicesResponse, errors.EdgeX)
}
