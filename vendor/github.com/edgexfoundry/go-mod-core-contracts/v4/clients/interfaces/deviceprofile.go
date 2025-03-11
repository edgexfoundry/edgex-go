//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// DeviceProfileClient defines the interface for interactions with the DeviceProfile endpoint on the EdgeX Foundry core-metadata service.
type DeviceProfileClient interface {
	// Add adds new profiles
	Add(ctx context.Context, reqs []requests.DeviceProfileRequest) ([]common.BaseWithIdResponse, errors.EdgeX)
	// Update updates profiles
	Update(ctx context.Context, reqs []requests.DeviceProfileRequest) ([]common.BaseResponse, errors.EdgeX)
	// AddByYaml adds new profile by uploading a file in YAML format
	AddByYaml(ctx context.Context, yamlFilePath string) (common.BaseWithIdResponse, errors.EdgeX)
	// UpdateByYaml updates profile by uploading a file in YAML format
	UpdateByYaml(ctx context.Context, yamlFilePath string) (common.BaseResponse, errors.EdgeX)
	// DeleteByName deletes profile by name
	DeleteByName(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
	// DeviceProfileByName queries profile by name
	DeviceProfileByName(ctx context.Context, name string) (responses.DeviceProfileResponse, errors.EdgeX)
	// AllDeviceProfiles queries all profiles
	AllDeviceProfiles(ctx context.Context, labels []string, offset int, limit int) (responses.MultiDeviceProfilesResponse, errors.EdgeX)
	// AllDeviceProfileBasicInfos queries all profile basic infos
	AllDeviceProfileBasicInfos(ctx context.Context, labels []string, offset int, limit int) (responses.MultiDeviceProfileBasicInfoResponse, errors.EdgeX)
	// DeviceProfilesByModel queries profiles by model
	DeviceProfilesByModel(ctx context.Context, model string, offset int, limit int) (responses.MultiDeviceProfilesResponse, errors.EdgeX)
	// DeviceProfilesByManufacturer queries profiles by manufacturer
	DeviceProfilesByManufacturer(ctx context.Context, manufacturer string, offset int, limit int) (responses.MultiDeviceProfilesResponse, errors.EdgeX)
	// DeviceProfilesByManufacturerAndModel queries profiles by manufacturer and model
	DeviceProfilesByManufacturerAndModel(ctx context.Context, manufacturer string, model string, offset int, limit int) (responses.MultiDeviceProfilesResponse, errors.EdgeX)
	// DeviceResourceByProfileNameAndResourceName queries the device resource by profileName and resourceName
	DeviceResourceByProfileNameAndResourceName(ctx context.Context, profileName string, resourceName string) (responses.DeviceResourceResponse, errors.EdgeX)
	// UpdateDeviceProfileBasicInfo updates existing profile's basic info
	UpdateDeviceProfileBasicInfo(ctx context.Context, reqs []requests.DeviceProfileBasicInfoRequest) ([]common.BaseResponse, errors.EdgeX)
	// AddDeviceProfileResource adds new device resource to an existing profile
	AddDeviceProfileResource(ctx context.Context, reqs []requests.AddDeviceResourceRequest) ([]common.BaseResponse, errors.EdgeX)
	// UpdateDeviceProfileResource updates existing device resource
	UpdateDeviceProfileResource(ctx context.Context, reqs []requests.UpdateDeviceResourceRequest) ([]common.BaseResponse, errors.EdgeX)
	// DeleteDeviceResourceByName deletes device resource by name
	DeleteDeviceResourceByName(ctx context.Context, profileName string, resourceName string) (common.BaseResponse, errors.EdgeX)
	// AddDeviceProfileDeviceCommand adds new device command to an existing profile
	AddDeviceProfileDeviceCommand(ctx context.Context, reqs []requests.AddDeviceCommandRequest) ([]common.BaseResponse, errors.EdgeX)
	// UpdateDeviceProfileDeviceCommand updates existing device command
	UpdateDeviceProfileDeviceCommand(ctx context.Context, reqs []requests.UpdateDeviceCommandRequest) ([]common.BaseResponse, errors.EdgeX)
	// DeleteDeviceCommandByName deletes device command by name
	DeleteDeviceCommandByName(ctx context.Context, profileName string, commandName string) (common.BaseResponse, errors.EdgeX)
}
