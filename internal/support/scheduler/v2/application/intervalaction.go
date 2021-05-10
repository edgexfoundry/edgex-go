//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	schedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	v2SchedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// The AddIntervalAction function accepts the new IntervalAction model from the controller function
// and then invokes AddIntervalAction function of infrastructure layer to add new IntervalAction
func AddIntervalAction(action models.IntervalAction, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	schedulerManager := v2SchedulerContainer.SchedulerManagerFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	// checks the interval existence by name
	_, edgeXerr = dbClient.IntervalByName(action.IntervalName)
	if edgeXerr != nil {
		return id, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	addedAction, err := dbClient.AddIntervalAction(action)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	err = schedulerManager.AddIntervalAction(action)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	lc.Debugf("IntervalAction created on DB successfully. IntervalAction ID: %s, Correlation-ID: %s ",
		addedAction.Id,
		correlation.FromContext(ctx))

	return addedAction.Id, nil
}

// AllIntervalActions query the intervalActions with offset and limit
func AllIntervalActions(offset int, limit int, dic *di.Container) (intervalActionDTOs []dtos.IntervalAction, err errors.EdgeX) {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	intervalActions, err := dbClient.AllIntervalActions(offset, limit)
	if err != nil {
		return intervalActionDTOs, errors.NewCommonEdgeXWrapper(err)
	}
	intervalActionDTOs = make([]dtos.IntervalAction, len(intervalActions))
	for i, action := range intervalActions {
		dto := dtos.FromIntervalActionModelToDTO(action)
		intervalActionDTOs[i] = dto
	}
	return intervalActionDTOs, nil
}

// IntervalActionByName query the intervalAction by name
func IntervalActionByName(name string, ctx context.Context, dic *di.Container) (dto dtos.IntervalAction, edgeXerr errors.EdgeX) {
	if name == "" {
		return dto, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	action, edgeXerr := dbClient.IntervalActionByName(name)
	if edgeXerr != nil {
		return dto, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	dto = dtos.FromIntervalActionModelToDTO(action)
	return dto, nil
}

// DeleteIntervalActionByName delete the intervalAction by name
func DeleteIntervalActionByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	schedulerManager := v2SchedulerContainer.SchedulerManagerFrom(dic.Get)

	err := dbClient.DeleteIntervalActionByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = schedulerManager.DeleteIntervalActionByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// PatchIntervalAction executes the PATCH operation with the DTO to replace the old data
func PatchIntervalAction(dto dtos.UpdateIntervalAction, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	schedulerManager := v2SchedulerContainer.SchedulerManagerFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	var action models.IntervalAction
	var edgeXerr errors.EdgeX
	if dto.Id != nil {
		action, edgeXerr = dbClient.IntervalActionById(*dto.Id)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	} else {
		action, edgeXerr = dbClient.IntervalActionByName(*dto.Name)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}
	if dto.Name != nil && *dto.Name != action.Name {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("interval name '%s' not match the exsting '%s' ", *dto.Name, action.Name), nil)
	}

	// checks the interval existence by name
	if dto.IntervalName != nil {
		_, edgeXerr = dbClient.IntervalByName(*dto.IntervalName)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}

	requests.ReplaceIntervalActionModelFieldsWithDTO(&action, dto)

	edgeXerr = dbClient.UpdateIntervalAction(action)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	edgeXerr = schedulerManager.UpdateIntervalAction(action)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	lc.Debugf(
		"IntervalAction patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)
	return nil
}

// LoadIntervalActionToSchedulerManager loads intervalActions to SchedulerManager before running the interval job
func LoadIntervalActionToSchedulerManager(dic *di.Container) errors.EdgeX {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	schedulerManager := v2SchedulerContainer.SchedulerManagerFrom(dic.Get)
	// Load intervalActions from config to DB
	configuration := schedulerContainer.ConfigurationFrom(dic.Get)
	for i := range configuration.IntervalActions {
		action := models.IntervalAction{
			Name:         configuration.IntervalActions[i].Name,
			IntervalName: configuration.IntervalActions[i].Interval,
			Address: models.RESTAddress{
				BaseAddress: models.BaseAddress{
					Type: v2.REST,
					Host: configuration.IntervalActions[i].Host,
					Port: configuration.IntervalActions[i].Port,
				},
				Path:       configuration.IntervalActions[i].Path,
				HTTPMethod: configuration.IntervalActions[i].Method,
			},
		}
		_, err := dbClient.IntervalActionByName(action.Name)
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			_, err = dbClient.AddIntervalAction(action)
			if err != nil {
				return errors.NewCommonEdgeXWrapper(err)
			}
		} else if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	// Load intervalActions from DB to scheduler
	actions, err := AllIntervalActions(0, -1, dic)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	for _, action := range actions {
		err = schedulerManager.AddIntervalAction(dtos.ToIntervalActionModel(action))
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	return nil
}
