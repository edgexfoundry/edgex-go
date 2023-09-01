//
// Copyright (C) 2021-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
)

var asyncPurgeReadingOnce sync.Once

// ReadingTotalCount return the count of all of readings currently stored in the database and error if any
func ReadingTotalCount(dic *di.Container) (uint32, errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	count, err := dbClient.ReadingTotalCount()
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}

// AllReadings query events by offset, and limit
func AllReadings(offset int, limit int, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	readingModels, err := dbClient.AllReadings(offset, limit)
	if err == nil {
		readings, err = convertReadingModelsToDTOs(readingModels)
		if err == nil {
			totalCount, err = dbClient.ReadingTotalCount()
		}
	}

	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

// ReadingsByResourceName query readings with offset, limit, and resource name
func ReadingsByResourceName(offset int, limit int, resourceName string, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if resourceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "resourceName is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByResourceName(offset, limit, resourceName)
	if err == nil {
		readings, err = convertReadingModelsToDTOs(readingModels)
		if err == nil {
			totalCount, err = dbClient.ReadingCountByResourceName(resourceName)
		}
	}

	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

// ReadingsByDeviceName query readings with offset, limit, and device name
func ReadingsByDeviceName(offset int, limit int, name string, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if name == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByDeviceName(offset, limit, name)
	if err == nil {
		readings, err = convertReadingModelsToDTOs(readingModels)
		if err == nil {
			totalCount, err = dbClient.ReadingCountByDeviceName(name)
		}
	}

	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

// ReadingsByTimeRange query readings with offset, limit and time range
func ReadingsByTimeRange(start int, end int, offset int, limit int, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByTimeRange(start, end, offset, limit)
	if err == nil {
		readings, err = convertReadingModelsToDTOs(readingModels)
		if err == nil {
			totalCount, err = dbClient.ReadingCountByTimeRange(start, end)
		}
	}

	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

func convertReadingModelsToDTOs(readingModels []models.Reading) (readings []dtos.BaseReading, err errors.EdgeX) {
	readings = make([]dtos.BaseReading, len(readingModels))
	for i, r := range readingModels {
		readings[i] = dtos.FromReadingModelToDTO(r)
	}
	return readings, nil
}

// ReadingCountByDeviceName return the count of all of readings associated with given device and error if any
func ReadingCountByDeviceName(deviceName string, dic *di.Container) (uint32, errors.EdgeX) {
	if deviceName == "" {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	count, err := dbClient.ReadingCountByDeviceName(deviceName)
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}

// ReadingsByResourceNameAndTimeRange returns readings by resource name and specified time range. Readings are sorted in descending order of origin time.
func ReadingsByResourceNameAndTimeRange(resourceName string, start int, end int, offset int, limit int, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if resourceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "resourceName is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByResourceNameAndTimeRange(resourceName, start, end, offset, limit)
	if err == nil {
		readings, err = convertReadingModelsToDTOs(readingModels)
		if err == nil {
			totalCount, err = dbClient.ReadingCountByResourceNameAndTimeRange(resourceName, start, end)
		}
	}

	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

// ReadingsByDeviceNameAndResourceName query readings with offset, limit, device name and its associated resource name
func ReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, offset int, limit int, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if deviceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if resourceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByDeviceNameAndResourceName(deviceName, resourceName, offset, limit)
	if err == nil {
		readings, err = convertReadingModelsToDTOs(readingModels)
		if err == nil {
			totalCount, err = dbClient.ReadingCountByDeviceNameAndResourceName(deviceName, resourceName)
		}
	}

	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

// ReadingsByDeviceNameAndResourceNameAndTimeRange query readings with offset, limit, device name, its associated resource name and specified time range
func ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start, end, offset, limit int, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if deviceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if resourceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, start, end, offset, limit)
	if err == nil {
		readings, err = convertReadingModelsToDTOs(readingModels)
		if err == nil {
			totalCount, err = dbClient.ReadingCountByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, start, end)
		}
	}

	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

// ReadingsByDeviceNameAndResourceNamesAndTimeRange query readings with offset, limit, device name, its associated resource name and specified time range
func ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start, end, offset, limit int, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if deviceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	var readingModels []models.Reading
	if len(resourceNames) > 0 {
		readingModels, totalCount, err = dbClient.ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName, resourceNames, start, end, offset, limit)
	} else {
		readingModels, err = dbClient.ReadingsByDeviceNameAndTimeRange(deviceName, start, end, offset, limit)
		if err == nil {
			totalCount, err = dbClient.ReadingCountByDeviceNameAndTimeRange(deviceName, start, end)
		}
	}

	if err == nil {
		readings, err = convertReadingModelsToDTOs(readingModels)
	}
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

// AsyncPurgeReading purge readings and related events according to the retention capability.
func AsyncPurgeReading(interval time.Duration, ctx context.Context, dic *di.Container) {
	asyncPurgeReadingOnce.Do(func() {
		go func() {
			lc := bootstrapContainer.LoggingClientFrom(dic.Get)
			timer := time.NewTimer(interval)
			for {
				timer.Reset(interval) // since event deletion might take lots of time, restart the timer to recount the time
				select {
				case <-ctx.Done():
					lc.Info("Exiting reading retention")
					return
				case <-timer.C:
					err := purgeReading(dic)
					if err != nil {
						lc.Errorf("Failed to purge events and readings, %v", err)
						break
					}
				}
			}
		}()
	})
}

func purgeReading(dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	total, err := dbClient.ReadingTotalCount()
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "failed to query reading total count, %v", err)
	}
	if total >= config.Retention.MaxCap {
		lc.Debugf("Purging the reading amount %d to the minimum capacity %d", total, config.Retention.MinCap)
		// Using reading origin instead event origin to remove events and readings by age.
		// If we remove readings to MinCap and remove related events, some readings might lose the related event.
		reading, err := dbClient.LatestReadingByOffset(config.Retention.MinCap)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query reading with offset '%d'", config.Retention.MinCap), err)
		}
		age := time.Now().UnixNano() - reading.GetBaseReading().Origin
		err = dbClient.DeleteEventsByAge(age)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings by age '%d'", age), err)
		}
	}
	return nil
}
