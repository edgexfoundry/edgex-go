//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	v2SchedulerContainer "github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// The AddInterval function accepts the new Interval model from the controller function
// and then invokes AddInterval function of infrastructure layer to add new Interval
func AddInterval(d models.Interval, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	addedInterval, err := dbClient.AddInterval(d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Interval created on DB successfully. Interval ID: %s, Correlation-ID: %s ",
		addedInterval.Id,
		correlation.FromContext(ctx))

	return addedInterval.Id, nil
}

// IntervalByName query the interval by name
func IntervalByName(name string, ctx context.Context, dic *di.Container) (dto dtos.Interval, edgeXerr errors.EdgeX) {
	if name == "" {
		return dto, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	interval, err := dbClient.IntervalByName(name)
	if err != nil {
		return dto, errors.NewCommonEdgeXWrapper(err)
	}
	dto = dtos.FromIntervalModelToDTO(interval)
	return dto, nil
}

// AllIntervals query the intervals with offset and limit
func AllIntervals(offset int, limit int, dic *di.Container) (intervalDTOs []dtos.Interval, err errors.EdgeX) {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	intervals, err := dbClient.AllIntervals(offset, limit)
	if err != nil {
		return intervalDTOs, errors.NewCommonEdgeXWrapper(err)
	}
	intervalDTOs = make([]dtos.Interval, len(intervals))
	for i, interval := range intervals {
		dto := dtos.FromIntervalModelToDTO(interval)
		intervalDTOs[i] = dto
	}
	return intervalDTOs, nil
}

// DeleteIntervalByName delete the interval by name
func DeleteIntervalByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)

	// TODO Check the associated intervalAction existence
	//actions, err := dbClient.IntervalActionsByIntervalName(0, 1, name)
	//if err != nil {
	//	return errors.NewCommonEdgeXWrapper(err)
	//}
	//if len(actions) > 0 {
	//	return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the interval when associated intervalAction exists", nil)
	//}

	// TODO Remove interval from SchedulerQueue
	//err = sqDeleter.RemoveIntervalInQueue(interval.ID)

	err := dbClient.DeleteIntervalByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// PatchInterval executes the PATCH operation with the DTO to replace the old data
func PatchInterval(dto dtos.UpdateInterval, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	var interval models.Interval
	var edgeXerr errors.EdgeX
	if dto.Id != nil {
		interval, edgeXerr = dbClient.IntervalById(*dto.Id)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	} else {
		interval, edgeXerr = dbClient.IntervalByName(*dto.Name)
		if edgeXerr != nil {
			return errors.NewCommonEdgeXWrapper(edgeXerr)
		}
	}
	if dto.Name != nil && *dto.Name != interval.Name {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("interval name '%s' not match the exsting '%s' ", *dto.Name, interval.Name), nil)
	}

	// TODO Check if the interval still has attached interval actions
	//stillInUse, err := op.isIntervalStillInUse(to)
	//if err != nil {
	//	return err
	//}
	//if stillInUse {
	//	return errors.NewErrIntervalStillInUse(to.Name)
	//}
	// TODO Update the Scheduler Queue
	//err = op.scClient.UpdateIntervalInQueue(op.interval)
	//if err != nil {
	//	return err
	//}

	requests.ReplaceIntervalModelFieldsWithDTO(&interval, dto)

	edgeXerr = dbClient.UpdateInterval(interval)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	lc.Debugf(
		"Interval patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)
	return nil
}
