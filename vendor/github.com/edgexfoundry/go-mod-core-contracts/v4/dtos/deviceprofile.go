//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"errors"
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type DeviceProfile struct {
	DeviceProfileBasicInfo `json:",inline" yaml:",inline"`
	DeviceResources        []DeviceResource `json:"deviceResources" yaml:"deviceResources" validate:"dive"`
	DeviceCommands         []DeviceCommand  `json:"deviceCommands" yaml:"deviceCommands" validate:"dive"`
	ApiVersion             string           `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

// Validate satisfies the Validator interface
func (dp *DeviceProfile) Validate() error {
	err := common.Validate(dp)
	if err != nil {
		// The DeviceProfileBasicInfo is the internal struct in Golang programming, not in the Profile model,
		// so it should be hidden from the error messages.
		err = errors.New(strings.ReplaceAll(err.Error(), ".DeviceProfileBasicInfo", ""))
		return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, "Invalid DeviceProfile.", err)
	}
	return ValidateDeviceProfileDTO(*dp)
}

// UnmarshalYAML implements the Unmarshaler interface for the DeviceProfile type
func (dp *DeviceProfile) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var alias struct {
		DeviceProfileBasicInfo `yaml:",inline"`
		DeviceResources        []DeviceResource `yaml:"deviceResources"`
		DeviceCommands         []DeviceCommand  `yaml:"deviceCommands"`
		ApiVersion             string           `yaml:"apiVersion"`
	}
	if err := unmarshal(&alias); err != nil {
		return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, "failed to unmarshal request body as YAML.", err)
	}
	*dp = DeviceProfile(alias)

	if err := dp.Validate(); err != nil {
		return edgexErrors.NewCommonEdgeXWrapper(err)
	}

	// Normalize resource's value type
	for i, resource := range dp.DeviceResources {
		valueType, err := common.NormalizeValueType(resource.Properties.ValueType)
		if err != nil {
			return edgexErrors.NewCommonEdgeXWrapper(err)
		}
		dp.DeviceResources[i].Properties.ValueType = valueType
	}
	return nil
}

// ToDeviceProfileModel transforms the DeviceProfile DTO to the DeviceProfile model
func ToDeviceProfileModel(deviceProfileDTO DeviceProfile) models.DeviceProfile {
	return models.DeviceProfile{
		DBTimestamp:     models.DBTimestamp(deviceProfileDTO.DBTimestamp),
		Id:              deviceProfileDTO.Id,
		Name:            deviceProfileDTO.Name,
		Description:     deviceProfileDTO.Description,
		Manufacturer:    deviceProfileDTO.Manufacturer,
		Model:           deviceProfileDTO.Model,
		Labels:          deviceProfileDTO.Labels,
		DeviceResources: ToDeviceResourceModels(deviceProfileDTO.DeviceResources),
		DeviceCommands:  ToDeviceCommandModels(deviceProfileDTO.DeviceCommands),
		ApiVersion:      deviceProfileDTO.ApiVersion,
	}
}

// FromDeviceProfileModelToDTO transforms the DeviceProfile Model to the DeviceProfile DTO
func FromDeviceProfileModelToDTO(deviceProfile models.DeviceProfile) DeviceProfile {
	if deviceProfile.ApiVersion == "" {
		deviceProfile.ApiVersion = common.ApiVersion
	}
	return DeviceProfile{
		DeviceProfileBasicInfo: DeviceProfileBasicInfo{
			DBTimestamp:  DBTimestamp(deviceProfile.DBTimestamp),
			Id:           deviceProfile.Id,
			Name:         deviceProfile.Name,
			Description:  deviceProfile.Description,
			Manufacturer: deviceProfile.Manufacturer,
			Model:        deviceProfile.Model,
			Labels:       deviceProfile.Labels,
		},
		DeviceResources: FromDeviceResourceModelsToDTOs(deviceProfile.DeviceResources),
		DeviceCommands:  FromDeviceCommandModelsToDTOs(deviceProfile.DeviceCommands),
		ApiVersion:      deviceProfile.ApiVersion,
	}
}

// FromDeviceProfileModelToBasicInfoDTO transforms the DeviceProfile Model to the DeviceProfileBasicInfo DTO
func FromDeviceProfileModelToBasicInfoDTO(deviceProfile models.DeviceProfile) DeviceProfileBasicInfo {
	return DeviceProfileBasicInfo{
		DBTimestamp:  DBTimestamp(deviceProfile.DBTimestamp),
		Id:           deviceProfile.Id,
		Name:         deviceProfile.Name,
		Description:  deviceProfile.Description,
		Manufacturer: deviceProfile.Manufacturer,
		Model:        deviceProfile.Model,
		Labels:       deviceProfile.Labels,
	}
}

func ValidateDeviceProfileDTO(profile DeviceProfile) error {
	// deviceResources validation
	dupCheck := make(map[string]bool)
	for _, resource := range profile.DeviceResources {
		if resource.Properties.ValueType == common.ValueTypeBinary &&
			strings.Contains(resource.Properties.ReadWrite, common.ReadWrite_W) {
			return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, fmt.Sprintf("write permission not support %s value type for resource '%s'", common.ValueTypeBinary, resource.Name), nil)
		}
		// deviceResource name should not duplicated
		if dupCheck[resource.Name] {
			return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, fmt.Sprintf("device resource %s is duplicated", resource.Name), nil)
		}
		dupCheck[resource.Name] = true
	}
	// deviceCommands validation
	dupCheck = make(map[string]bool)
	for _, command := range profile.DeviceCommands {
		// deviceCommand name should not duplicated
		if dupCheck[command.Name] {
			return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, fmt.Sprintf("device command %s is duplicated", command.Name), nil)
		}
		dupCheck[command.Name] = true

		resourceOperations := command.ResourceOperations
		for _, ro := range resourceOperations {
			// ResourceOperations referenced in deviceCommands must exist
			if !deviceResourcesContains(profile.DeviceResources, ro.DeviceResource) {
				return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, fmt.Sprintf("device command's resource %s doesn't match any device resource", ro.DeviceResource), nil)
			}
			// Check the ReadWrite whether is align to the deviceResource
			if !validReadWritePermission(profile.DeviceResources, ro.DeviceResource, command.ReadWrite) {
				return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, fmt.Sprintf("device command's ReadWrite permission '%s' doesn't align the device resource", command.ReadWrite), nil)
			}
		}
	}
	return nil
}

func deviceResourcesContains(resources []DeviceResource, name string) bool {
	contains := false
	for _, resource := range resources {
		if resource.Name == name {
			contains = true
			break
		}
	}
	return contains
}

func validReadWritePermission(resources []DeviceResource, name string, readWrite string) bool {
	valid := true
	for _, resource := range resources {
		if resource.Name == name {
			if resource.Properties.ReadWrite != common.ReadWrite_RW && resource.Properties.ReadWrite != common.ReadWrite_WR &&
				resource.Properties.ReadWrite != readWrite {
				valid = false
				break
			}
		}
	}
	return valid
}
