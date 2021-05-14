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

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// The AddInterval function accepts the new Interval model from the controller function
// and then invokes AddInterval function of infrastructure layer to add new Interval
func AddInterval(interval models.Interval, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	schedulerManager := v2SchedulerContainer.SchedulerManagerFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	addedInterval, err := dbClient.AddInterval(interval)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	err = schedulerManager.AddInterval(interval)
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
	schedulerManager := v2SchedulerContainer.SchedulerManagerFrom(dic.Get)

	actions, err := dbClient.IntervalActionsByIntervalName(0, 1, name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(actions) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to delete the interval when associated intervalAction exists", nil)
	}

	err = dbClient.DeleteIntervalByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = schedulerManager.DeleteIntervalByName(name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}

// PatchInterval executes the PATCH operation with the DTO to replace the old data
func PatchInterval(dto dtos.UpdateInterval, ctx context.Context, dic *di.Container) errors.EdgeX {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	schedulerManager := v2SchedulerContainer.SchedulerManagerFrom(dic.Get)
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

	actions, err := dbClient.IntervalActionsByIntervalName(0, 1, interval.Name)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if len(actions) > 0 {
		return errors.NewCommonEdgeX(errors.KindStatusConflict, "fail to patch the interval when associated intervalAction exists", nil)
	}

	requests.ReplaceIntervalModelFieldsWithDTO(&interval, dto)

	edgeXerr = dbClient.UpdateInterval(interval)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	edgeXerr = schedulerManager.UpdateInterval(interval)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	lc.Debugf(
		"Interval patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)
	return nil
}

// LoadIntervalToSchedulerManager loads intervals to SchedulerManager before running the interval job
func LoadIntervalToSchedulerManager(dic *di.Container) errors.EdgeX {
	dbClient := v2SchedulerContainer.DBClientFrom(dic.Get)
	schedulerManager := v2SchedulerContainer.SchedulerManagerFrom(dic.Get)
	// Load intervals from config to DB
	configuration := schedulerContainer.ConfigurationFrom(dic.Get)
	for i := range configuration.Intervals {
		interval := models.Interval{
			Name:     configuration.Intervals[i].Name,
			Start:    configuration.Intervals[i].Start,
			End:      configuration.Intervals[i].End,
			Interval: configuration.Intervals[i].Interval,
			RunOnce:  configuration.Intervals[i].RunOnce,
		}
		_, err := dbClient.IntervalByName(interval.Name)
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			_, err = dbClient.AddInterval(interval)
			if err != nil {
				return errors.NewCommonEdgeXWrapper(err)
			}
		} else if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	// Load intervals from DB to scheduler
	intervals, err := AllIntervals(0, -1, dic)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	for _, interval := range intervals {
		err = schedulerManager.AddInterval(dtos.ToIntervalModel(interval))
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	return nil
}
