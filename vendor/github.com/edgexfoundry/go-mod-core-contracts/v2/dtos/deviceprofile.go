//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// DeviceProfile and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-metadata/2.1.0#/DeviceProfile
type DeviceProfile struct {
	DBTimestamp     `json:",inline"`
	Id              string           `json:"id" validate:"omitempty,uuid"`
	Name            string           `json:"name" yaml:"name" validate:"required,edgex-dto-none-empty-string,edgex-dto-rfc3986-unreserved-chars"`
	Manufacturer    string           `json:"manufacturer" yaml:"manufacturer"`
	Description     string           `json:"description" yaml:"description"`
	Model           string           `json:"model" yaml:"model"`
	Labels          []string         `json:"labels" yaml:"labels,flow"`
	DeviceResources []DeviceResource `json:"deviceResources" yaml:"deviceResources" validate:"required,gt=0,dive"`
	DeviceCommands  []DeviceCommand  `json:"deviceCommands" yaml:"deviceCommands" validate:"dive"`
}

// Validate satisfies the Validator interface
func (dp *DeviceProfile) Validate() error {
	err := common.Validate(dp)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid DeviceProfile.", err)
	}
	return ValidateDeviceProfileDTO(*dp)
}

// UnmarshalYAML implements the Unmarshaler interface for the DeviceProfile type
func (dp *DeviceProfile) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var alias struct {
		DBTimestamp
		Id              string           `yaml:"id"`
		Name            string           `yaml:"name"`
		Manufacturer    string           `yaml:"manufacturer"`
		Description     string           `yaml:"description"`
		Model           string           `yaml:"model"`
		Labels          []string         `yaml:"labels"`
		DeviceResources []DeviceResource `yaml:"deviceResources"`
		DeviceCommands  []DeviceCommand  `yaml:"deviceCommands"`
	}
	if err := unmarshal(&alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to unmarshal request body as YAML.", err)
	}
	*dp = DeviceProfile(alias)

	if err := dp.Validate(); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Normalize resource's value type
	for i, resource := range dp.DeviceResources {
		valueType, err := common.NormalizeValueType(resource.Properties.ValueType)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
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
	}
}

// FromDeviceProfileModelToDTO transforms the DeviceProfile Model to the DeviceProfile DTO
func FromDeviceProfileModelToDTO(deviceProfile models.DeviceProfile) DeviceProfile {
	return DeviceProfile{
		DBTimestamp:     DBTimestamp(deviceProfile.DBTimestamp),
		Id:              deviceProfile.Id,
		Name:            deviceProfile.Name,
		Description:     deviceProfile.Description,
		Manufacturer:    deviceProfile.Manufacturer,
		Model:           deviceProfile.Model,
		Labels:          deviceProfile.Labels,
		DeviceResources: FromDeviceResourceModelsToDTOs(deviceProfile.DeviceResources),
		DeviceCommands:  FromDeviceCommandModelsToDTOs(deviceProfile.DeviceCommands),
	}
}

func ValidateDeviceProfileDTO(profile DeviceProfile) error {
	// deviceResources validation
	dupCheck := make(map[string]bool)
	for _, resource := range profile.DeviceResources {
		if resource.Properties.ValueType == common.ValueTypeBinary &&
			strings.Contains(resource.Properties.ReadWrite, common.ReadWrite_W) {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("write permission not support %s value type for resource '%s'", common.ValueTypeBinary, resource.Name), nil)
		}
		// deviceResource name should not duplicated
		if dupCheck[resource.Name] {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device resource %s is duplicated", resource.Name), nil)
		}
		dupCheck[resource.Name] = true
	}
	// deviceCommands validation
	dupCheck = make(map[string]bool)
	for _, command := range profile.DeviceCommands {
		// deviceCommand name should not duplicated
		if dupCheck[command.Name] {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device command %s is duplicated", command.Name), nil)
		}
		dupCheck[command.Name] = true

		resourceOperations := command.ResourceOperations
		for _, ro := range resourceOperations {
			// ResourceOperations referenced in deviceCommands must exist
			if !deviceResourcesContains(profile.DeviceResources, ro.DeviceResource) {
				return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device command's resource %s doesn't match any device resource", ro.DeviceResource), nil)
			}
			// Check the ReadWrite whether is align to the deviceResource
			if !validReadWritePermission(profile.DeviceResources, ro.DeviceResource, command.ReadWrite) {
				return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("device command's ReadWrite permission '%s' doesn't align the device resource", command.ReadWrite), nil)
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
			if resource.Properties.ReadWrite != common.ReadWrite_RW &&
				resource.Properties.ReadWrite != readWrite {
				valid = false
				break
			}
		}
	}
	return valid
}
