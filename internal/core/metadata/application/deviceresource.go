//
// Copyright (C) 2021-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// DeviceResourceByProfileNameAndResourceName query the device resource by profileName and resourceName
func DeviceResourceByProfileNameAndResourceName(profileName string, resourceName string, dic *di.Container) (resource dtos.DeviceResource, err errors.EdgeX) {
	if profileName == "" {
		return resource, errors.NewCommonEdgeX(errors.KindContractInvalid, "profile name is empty", nil)
	}
	if resourceName == "" {
		return resource, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	profile, err := dbClient.DeviceProfileByName(profileName)
	if err != nil {
		return resource, errors.NewCommonEdgeXWrapper(err)
	}
	r, err := resourceByName(profile.DeviceResources, resourceName)
	if err != nil {
		return resource, errors.NewCommonEdgeXWrapper(err)
	}

	resource = dtos.FromDeviceResourceModelToDTO(r)
	return resource, nil
}

func resourceByName(resources []models.DeviceResource, resourceName string) (models.DeviceResource, errors.EdgeX) {
	for _, r := range resources {
		if r.Name == resourceName {
			return r, nil
		}
	}
	return models.DeviceResource{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("resource %s not exists", resourceName), nil)
}

func AddDeviceProfileResource(profileName string, resource models.DeviceResource, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	profile, err := dbClient.DeviceProfileByName(profileName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	err = deviceResourceUoMValidation(resource, dic)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	profile.DeviceResources = append(profile.DeviceResources, resource)

	profileDTO := dtos.FromDeviceProfileModelToDTO(profile)
	validateErr := profileDTO.Validate()
	if validateErr != nil {
		return errors.NewCommonEdgeXWrapper(validateErr)
	}

	err = dbClient.UpdateDeviceProfile(profile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("DeviceProfile deviceResources added on DB successfully. Correlation-id: %s ", correlation.FromContext(ctx))
	go publishUpdateDeviceProfileSystemEvent(profileDTO, ctx, dic)

	return nil
}

func PatchDeviceProfileResource(profileName string, dto dtos.UpdateDeviceResource, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	profile, err := dbClient.DeviceProfileByName(profileName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Find matched deviceResource
	index := -1
	for i := range profile.DeviceResources {
		if profile.DeviceResources[i].Name == *dto.Name {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device resource not found", nil)
	}

	requests.ReplaceDeviceResourceModelFieldsWithDTO(&profile.DeviceResources[index], dto)

	err = dbClient.UpdateDeviceProfile(profile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("DeviceProfile deviceResources patched on DB successfully. Correlation-id: %s ", correlation.FromContext(ctx))
	profileDTO := dtos.FromDeviceProfileModelToDTO(profile)
	go publishUpdateDeviceProfileSystemEvent(profileDTO, ctx, dic)

	return nil
}

func DeleteDeviceResourceByName(profileName string, resourceName string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if profileName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "profile name is empty", nil)
	}
	if resourceName == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	strictProfileChanges := container.ConfigurationFrom(dic.Get).Writable.ProfileChange.StrictDeviceProfileChanges
	if strictProfileChanges {
		return errors.NewCommonEdgeX(errors.KindServiceLocked, "profile change is not allowed when StrictDeviceProfileChanges config is enabled", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	// Check the associated Device existence
	devices, err := dbClient.DevicesByProfileName(0, 1, profileName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	} else if len(devices) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to update the device profile when associated device exists", nil)
	}

	profile, err := dbClient.DeviceProfileByName(profileName)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	index := -1
	for i := range profile.DeviceResources {
		if profile.DeviceResources[i].Name == resourceName {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "device resource not found", nil)
	}

	profile.DeviceResources = append(profile.DeviceResources[:index], profile.DeviceResources[index+1:]...)
	profileDTO := dtos.FromDeviceProfileModelToDTO(profile)
	e := (&profileDTO).Validate()
	if e != nil {
		return errors.NewCommonEdgeXWrapper(e)
	}

	err = dbClient.UpdateDeviceProfile(profile)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	go publishUpdateDeviceProfileSystemEvent(profileDTO, ctx, dic)
	return nil
}

func deviceResourceUoMValidation(r models.DeviceResource, dic *di.Container) errors.EdgeX {
	if container.ConfigurationFrom(dic.Get).Writable.UoM.Validation {
		uom := container.UnitsOfMeasureFrom(dic.Get)
		if ok := uom.Validate(r.Properties.Units); !ok {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("DeviceResource %s units %s is invalid", r.Name, r.Properties.Units), nil)
		}
	}

	return nil
}
