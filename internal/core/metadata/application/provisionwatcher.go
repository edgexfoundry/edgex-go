//
// Copyright (C) 2021-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// AddProvisionWatcher function accepts the new provision watcher model from the controller function
// and then invokes AddProvisionWatcher function of infrastructure layer to add new device service
func AddProvisionWatcher(pw models.ProvisionWatcher, ctx context.Context, dic *di.Container) (id string, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	// check the associated ProfileName existence
	if pw.DiscoveredDevice.ProfileName != "" {
		exists, err := dbClient.DeviceProfileNameExists(pw.DiscoveredDevice.ProfileName)
		if err != nil {
			return "", errors.NewCommonEdgeXWrapper(err)
		} else if !exists {
			return "", errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device profile '%s' does not exists", pw.DiscoveredDevice.ProfileName), err)
		}
	}

	addProvisionWatcher, err := dbClient.AddProvisionWatcher(pw)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("ProvisionWatcher created on DB successfully. ProvisionWatcher ID: %s, Correlation-ID: %s ",
		addProvisionWatcher.Id,
		correlationId,
	)
	go publishSystemEvent(common.ProvisionWatcherSystemEventType, common.SystemEventActionAdd, pw.ServiceName, dtos.FromProvisionWatcherModelToDTO(pw), ctx, dic)
	return addProvisionWatcher.Id, nil
}

// ProvisionWatcherByName query the provision watcher by name
func ProvisionWatcherByName(name string, dic *di.Container) (provisionWatcher dtos.ProvisionWatcher, err errors.EdgeX) {
	if name == "" {
		return provisionWatcher, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	pw, err := dbClient.ProvisionWatcherByName(name)
	if err != nil {
		return provisionWatcher, errors.NewCommonEdgeXWrapper(err)
	}
	provisionWatcher = dtos.FromProvisionWatcherModelToDTO(pw)

	return
}

// ProvisionWatchersByServiceName query provision watchers with offset, limit and service name
func ProvisionWatchersByServiceName(offset int, limit int, name string, dic *di.Container) (provisionWatchers []dtos.ProvisionWatcher, totalCount int64, err errors.EdgeX) {
	if name == "" {
		return provisionWatchers, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	totalCount, err = dbClient.ProvisionWatcherCountByServiceName(name)
	if err != nil {
		return provisionWatchers, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.ProvisionWatcher{}, totalCount, err
	}

	pwModels, err := dbClient.ProvisionWatchersByServiceName(offset, limit, name)
	if err != nil {
		return provisionWatchers, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	provisionWatchers = make([]dtos.ProvisionWatcher, len(pwModels))
	for i, pw := range pwModels {
		provisionWatchers[i] = dtos.FromProvisionWatcherModelToDTO(pw)
	}
	return provisionWatchers, totalCount, nil
}

// ProvisionWatchersByProfileName query provision watchers with offset, limit and profile name
func ProvisionWatchersByProfileName(offset int, limit int, name string, dic *di.Container) (provisionWatchers []dtos.ProvisionWatcher, totalCount int64, err errors.EdgeX) {
	if name == "" {
		return provisionWatchers, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	totalCount, err = dbClient.ProvisionWatcherCountByProfileName(name)
	if err != nil {
		return provisionWatchers, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.ProvisionWatcher{}, totalCount, err
	}

	pwModels, err := dbClient.ProvisionWatchersByProfileName(offset, limit, name)
	if err != nil {
		return provisionWatchers, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	provisionWatchers = make([]dtos.ProvisionWatcher, len(pwModels))
	for i, pw := range pwModels {
		provisionWatchers[i] = dtos.FromProvisionWatcherModelToDTO(pw)
	}
	return provisionWatchers, totalCount, nil
}

// AllProvisionWatchers query the provision watchers with offset, limit and labels
func AllProvisionWatchers(offset int, limit int, labels []string, dic *di.Container) (provisionWatchers []dtos.ProvisionWatcher, totalCount int64, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	totalCount, err = dbClient.ProvisionWatcherCountByLabels(labels)
	if err != nil {
		return provisionWatchers, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, offset, limit)
	if !cont {
		return []dtos.ProvisionWatcher{}, totalCount, err
	}

	pwModels, err := dbClient.AllProvisionWatchers(offset, limit, labels)
	if err != nil {
		return provisionWatchers, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	provisionWatchers = make([]dtos.ProvisionWatcher, len(pwModels))
	for i, pw := range pwModels {
		provisionWatchers[i] = dtos.FromProvisionWatcherModelToDTO(pw)
	}
	return provisionWatchers, totalCount, nil
}

// DeleteProvisionWatcherByName deletes the provision watcher by name
func DeleteProvisionWatcherByName(ctx context.Context, name string, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	pw, err := dbClient.ProvisionWatcherByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = dbClient.DeleteProvisionWatcherByName(pw.Name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	go publishSystemEvent(common.ProvisionWatcherSystemEventType, common.SystemEventActionDelete, pw.ServiceName, dtos.FromProvisionWatcherModelToDTO(pw), ctx, dic)
	return nil
}

// PatchProvisionWatcher executes the PATCH operation with the provisionWatcher DTO to replace the old data
func PatchProvisionWatcher(ctx context.Context, dto dtos.UpdateProvisionWatcher, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	pw, err := provisionWatcherByDTO(dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Old service name is used for invoking callback
	var oldServiceName string
	if dto.ServiceName != nil && *dto.ServiceName != pw.ServiceName {
		oldServiceName = pw.ServiceName
	}

	requests.ReplaceProvisionWatcherModelFieldsWithDTO(&pw, dto)

	err = dbClient.UpdateProvisionWatcher(pw)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("ProvisionWatcher patched on DB successfully. Correlation-ID: %s ", correlation.FromContext(ctx))

	if oldServiceName != "" {
		go publishSystemEvent(common.ProvisionWatcherSystemEventType, common.SystemEventActionUpdate, oldServiceName, dtos.FromProvisionWatcherModelToDTO(pw), ctx, dic)
	}
	go publishSystemEvent(common.ProvisionWatcherSystemEventType, common.SystemEventActionUpdate, pw.ServiceName, dtos.FromProvisionWatcherModelToDTO(pw), ctx, dic)
	return nil
}

func provisionWatcherByDTO(dbClient interfaces.DBClient, dto dtos.UpdateProvisionWatcher) (pw models.ProvisionWatcher, edgexErr errors.EdgeX) {
	// The ID or Name is required by DTO and the DTO also accepts empty string ID if the Name is provided
	if dto.Id != nil && *dto.Id != "" {
		pw, edgexErr = dbClient.ProvisionWatcherById(*dto.Id)
		if edgexErr != nil {
			return pw, errors.NewCommonEdgeXWrapper(edgexErr)
		}
	} else {
		pw, edgexErr = dbClient.ProvisionWatcherByName(*dto.Name)
		if edgexErr != nil {
			return pw, errors.NewCommonEdgeXWrapper(edgexErr)
		}
	}
	if dto.Name != nil && *dto.Name != pw.Name {
		return pw, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("provision watcher name '%s' not match the existing '%s' ", *dto.Name, pw.Name), nil)
	}
	return pw, nil
}
