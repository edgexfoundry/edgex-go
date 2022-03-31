//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
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

	return nil
}
