//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"
)

// AddProvisionWatcher function accepts the new provision watcher model from the controller function
// and then invokes AddProvisionWatcher function of infrastructure layer to add new device service
func AddProvisionWatcher(pw models.ProvisionWatcher, ctx context.Context, dic *di.Container) (id string, err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	addProvisionWatcher, err := dbClient.AddProvisionWatcher(pw)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("ProvisionWatcher created on DB successfully. ProvisionWatcher ID: %s, Correlation-ID: %s ",
		addProvisionWatcher.Id,
		correlationId,
	)

	return addProvisionWatcher.Id, nil
}

// ProvisionWatcherByName query the provision watcher by name
func ProvisionWatcherByName(name string, dic *di.Container) (provisionWatcher dtos.ProvisionWatcher, err errors.EdgeX) {
	if name == "" {
		return provisionWatcher, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	pw, err := dbClient.ProvisionWatcherByName(name)
	if err != nil {
		return provisionWatcher, errors.NewCommonEdgeXWrapper(err)
	}
	provisionWatcher = dtos.FromProvisionWatcherModelToDTO(pw)

	return
}

// ProvisionWatchersByServiceName query provision watchers with offset, limit and service name
func ProvisionWatchersByServiceName(offset int, limit int, name string, dic *di.Container) (provisionWatchers []dtos.ProvisionWatcher, err errors.EdgeX) {
	if name == "" {
		return provisionWatchers, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	pwModels, err := dbClient.ProvisionWatchersByServiceName(offset, limit, name)
	if err != nil {
		return provisionWatchers, errors.NewCommonEdgeXWrapper(err)
	}

	provisionWatchers = make([]dtos.ProvisionWatcher, len(pwModels))
	for i, pw := range pwModels {
		provisionWatchers[i] = dtos.FromProvisionWatcherModelToDTO(pw)
	}

	return
}

// ProvisionWatchersByProfileName query provision watchers with offset, limit and profile name
func ProvisionWatchersByProfileName(offset int, limit int, name string, dic *di.Container) (provisionWatchers []dtos.ProvisionWatcher, err errors.EdgeX) {
	if name == "" {
		return provisionWatchers, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	pwModels, err := dbClient.ProvisionWatchersByProfileName(offset, limit, name)
	if err != nil {
		return provisionWatchers, errors.NewCommonEdgeXWrapper(err)
	}

	provisionWatchers = make([]dtos.ProvisionWatcher, len(pwModels))
	for i, pw := range pwModels {
		provisionWatchers[i] = dtos.FromProvisionWatcherModelToDTO(pw)
	}

	return
}

// AllProvisionWatchers query the provision watchers with offset, limit and labels
func AllProvisionWatchers(offset int, limit int, labels []string, dic *di.Container) (provisionWatchers []dtos.ProvisionWatcher, err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	pwModels, err := dbClient.AllProvisionWatchers(offset, limit, labels)
	if err != nil {
		return provisionWatchers, errors.NewCommonEdgeXWrapper(err)
	}

	provisionWatchers = make([]dtos.ProvisionWatcher, len(pwModels))
	for i, pw := range pwModels {
		provisionWatchers[i] = dtos.FromProvisionWatcherModelToDTO(pw)
	}

	return
}

// DeleteProvisionWatcherByName deletes the provision watcher by name
func DeleteProvisionWatcherByName(name string, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	err := dbClient.DeleteProvisionWatcherByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// PatchProvisionWatcher executes the PATCH operation with the provisionWatcher DTO to replace the old data
func PatchProvisionWatcher(ctx context.Context, dto dtos.UpdateProvisionWatcher, dic *di.Container) errors.EdgeX {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	var provisionWatcher models.ProvisionWatcher
	var edgexErr errors.EdgeX
	if dto.Name != nil {
		if *dto.Name == "" {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
		}
		provisionWatcher, edgexErr = dbClient.ProvisionWatcherByName(*dto.Name)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
	} else {
		if *dto.Id == "" {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "id is empty", nil)
		}
		_, err := uuid.Parse(*dto.Id)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindInvalidId, "failed to parse id as an UUID", err)
		}
		provisionWatcher, edgexErr = dbClient.ProvisionWatcherById(*dto.Id)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
	}
	if dto.Name != nil && *dto.Name != provisionWatcher.Name {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("provision watcher name '%s' not match the existing '%s' ", *dto.Name, provisionWatcher.Name), nil)
	}

	requests.ReplaceProvisionWatcherModelFieldsWithDTO(&provisionWatcher, dto)
	exists, edgeXerr := dbClient.DeviceServiceNameExists(provisionWatcher.ServiceName)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("device service '%s' existence check failed", provisionWatcher.ServiceName), edgeXerr)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device service '%s' does not exist", provisionWatcher.ServiceName), nil)
	}
	exists, edgeXerr = dbClient.DeviceProfileNameExists(provisionWatcher.ProfileName)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("device profile '%s' existence check failed", provisionWatcher.ProfileName), edgeXerr)
	} else if !exists {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exist", provisionWatcher.ProfileName), nil)
	}

	edgexErr = dbClient.DeleteProvisionWatcherByName(provisionWatcher.Name)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	_, edgexErr = dbClient.AddProvisionWatcher(provisionWatcher)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	lc.Debugf("ProvisionWatcher patched on DB successfully. Correlation-ID: %s ", correlation.FromContext(ctx))
	return nil
}
