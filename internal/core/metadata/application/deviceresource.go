//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
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
