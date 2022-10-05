//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/container"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/infrastructure/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// The AddIntervalAction function accepts the new IntervalAction model from the controller function
// and then invokes AddIntervalAction function of infrastructure layer to add new IntervalAction
func AddIntervalAction(action models.IntervalAction, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

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
func AllIntervalActions(offset int, limit int, dic *di.Container) (intervalActionDTOs []dtos.IntervalAction, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	intervalActions, err := dbClient.AllIntervalActions(offset, limit)
	if err == nil {
		totalCount, err = dbClient.IntervalActionTotalCount()
	}
	if err != nil {
		return intervalActionDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	intervalActionDTOs = make([]dtos.IntervalAction, len(intervalActions))
	for i, action := range intervalActions {
		dto := dtos.FromIntervalActionModelToDTO(action)
		intervalActionDTOs[i] = dto
	}
	return intervalActionDTOs, totalCount, nil
}

// IntervalActionByName query the intervalAction by name
func IntervalActionByName(name string, ctx context.Context, dic *di.Container) (dto dtos.IntervalAction, edgeXerr errors.EdgeX) {
	if name == "" {
		return dto, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
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
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)

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
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	action, err := intervalActionByDTO(dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	requests.ReplaceIntervalActionModelFieldsWithDTO(&action, dto)

	err = dbClient.UpdateIntervalAction(action)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = schedulerManager.UpdateIntervalAction(action)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	lc.Debugf(
		"IntervalAction patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)
	return nil
}

func intervalActionByDTO(dbClient interfaces.DBClient, dto dtos.UpdateIntervalAction) (action models.IntervalAction, err errors.EdgeX) {
	// The ID or Name is required by DTO and the DTO also accepts empty string ID if the Name is provided
	if dto.Id != nil && *dto.Id != "" {
		action, err = dbClient.IntervalActionById(*dto.Id)
		if err != nil {
			return action, errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		action, err = dbClient.IntervalActionByName(*dto.Name)
		if err != nil {
			return action, errors.NewCommonEdgeXWrapper(err)
		}
	}
	if dto.Name != nil && *dto.Name != action.Name {
		return action, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("intervalAction name '%s' not match the exsting '%s' ", *dto.Name, action.Name), nil)
	}
	return action, nil
}

// LoadIntervalActionToSchedulerManager loads intervalActions to SchedulerManager before running the interval job
func LoadIntervalActionToSchedulerManager(dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	// Load intervalActions from config to DB
	configuration := container.ConfigurationFrom(dic.Get)
	for i := range configuration.IntervalActions {
		dto := dtos.IntervalAction{
			Name:         configuration.IntervalActions[i].Name,
			IntervalName: configuration.IntervalActions[i].Interval,
			Address: dtos.Address{
				Type: common.REST,
				Host: configuration.IntervalActions[i].Host,
				Port: configuration.IntervalActions[i].Port,
				RESTAddress: dtos.RESTAddress{
					Path:       configuration.IntervalActions[i].Path,
					HTTPMethod: configuration.IntervalActions[i].Method,
				},
			},
			Content:     configuration.IntervalActions[i].Content,
			ContentType: configuration.IntervalActions[i].ContentType,
			AdminState:  configuration.IntervalActions[i].AdminState,
		}
		validateErr := common.Validate(dto)
		if validateErr != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("validate pre-defined IntervalAction %s from configuration failed", dto.Name), validateErr)
		}
		_, err := dbClient.IntervalByName(dto.IntervalName)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		action := dtos.ToIntervalActionModel(dto)
		_, err = dbClient.IntervalActionByName(action.Name)
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
	actions, _, err := AllIntervalActions(0, -1, dic)
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
