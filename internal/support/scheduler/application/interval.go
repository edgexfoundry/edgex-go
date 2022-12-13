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

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// The AddInterval function accepts the new Interval model from the controller function
// and then invokes AddInterval function of infrastructure layer to add new Interval
func AddInterval(interval models.Interval, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

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
	dbClient := container.DBClientFrom(dic.Get)
	interval, err := dbClient.IntervalByName(name)
	if err != nil {
		return dto, errors.NewCommonEdgeXWrapper(err)
	}
	dto = dtos.FromIntervalModelToDTO(interval)
	return dto, nil
}

// AllIntervals query the intervals with offset and limit
func AllIntervals(offset int, limit int, dic *di.Container) (intervalDTOs []dtos.Interval, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	intervals, err := dbClient.AllIntervals(offset, limit)
	if err == nil {
		totalCount, err = dbClient.IntervalTotalCount()
	}
	if err != nil {
		return intervalDTOs, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	intervalDTOs = make([]dtos.Interval, len(intervals))
	for i, interval := range intervals {
		dto := dtos.FromIntervalModelToDTO(interval)
		intervalDTOs[i] = dto
	}
	return intervalDTOs, totalCount, nil
}

// DeleteIntervalByName delete the interval by name
func DeleteIntervalByName(name string, ctx context.Context, dic *di.Container) errors.EdgeX {
	if name == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	err := dbClient.DeleteIntervalByName(name)
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
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	interval, err := intervalByDTO(dbClient, dto)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	requests.ReplaceIntervalModelFieldsWithDTO(&interval, dto)

	err = dbClient.UpdateInterval(interval)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	err = schedulerManager.UpdateInterval(interval)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf(
		"Interval patched on DB successfully. Correlation-ID: %s ",
		correlation.FromContext(ctx),
	)
	return nil
}

func intervalByDTO(dbClient interfaces.DBClient, dto dtos.UpdateInterval) (interval models.Interval, err errors.EdgeX) {
	// The ID or Name is required by DTO and the DTO also accepts empty string ID if the Name is provided
	if dto.Id != nil && *dto.Id != "" {
		interval, err = dbClient.IntervalById(*dto.Id)
		if err != nil {
			return interval, errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		interval, err = dbClient.IntervalByName(*dto.Name)
		if err != nil {
			return interval, errors.NewCommonEdgeXWrapper(err)
		}
	}
	if dto.Name != nil && *dto.Name != interval.Name {
		return interval, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("interval name '%s' not match the exsting '%s' ", *dto.Name, interval.Name), nil)
	}
	return interval, nil
}

// LoadIntervalToSchedulerManager loads intervals to SchedulerManager before running the interval job
func LoadIntervalToSchedulerManager(dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	schedulerManager := container.SchedulerManagerFrom(dic.Get)
	// Load intervals from config to DB
	configuration := container.ConfigurationFrom(dic.Get)
	for i := range configuration.Intervals {
		dto := dtos.Interval{
			Name:     configuration.Intervals[i].Name,
			Start:    configuration.Intervals[i].Start,
			End:      configuration.Intervals[i].End,
			Interval: configuration.Intervals[i].Interval,
		}
		validateErr := common.Validate(dto)
		if validateErr != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("validate pre-defined Interval %s from configuration failed", dto.Name), validateErr)
		}
		interval := dtos.ToIntervalModel(dto)
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
	intervals, _, err := AllIntervals(0, -1, dic)
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
