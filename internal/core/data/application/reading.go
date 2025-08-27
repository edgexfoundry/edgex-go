//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/query"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
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
func AllReadings(parms query.Parameters, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.AllReadings(parms.Offset, parms.Limit)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	processNumericReadings(parms.Numeric, readings)

	if parms.Offset < 0 {
		return readings, 0, err // skip total count
	}
	totalCount, err = dbClient.ReadingTotalCount()
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.BaseReading{}, totalCount, err
	}
	return readings, totalCount, err
}

// ReadingsByResourceName query readings with offset, limit, and resource name
func ReadingsByResourceName(parms query.Parameters, resourceName string, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if resourceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "resourceName is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsByResourceName(parms.Offset, parms.Limit, resourceName)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	processNumericReadings(parms.Numeric, readings)

	if parms.Offset < 0 {
		return readings, 0, err // skip total count
	}
	totalCount, err = dbClient.ReadingCountByResourceName(resourceName)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.BaseReading{}, totalCount, err
	}
	return readings, totalCount, err
}

// ReadingsByDeviceName query readings with offset, limit, and device name
func ReadingsByDeviceName(parms query.Parameters, name string, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if name == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsByDeviceName(parms.Offset, parms.Limit, name)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	processNumericReadings(parms.Numeric, readings)

	if parms.Offset < 0 {
		return readings, 0, err // skip total count
	}
	totalCount, err = dbClient.ReadingCountByDeviceName(name)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.BaseReading{}, totalCount, err
	}
	return readings, totalCount, err
}

// ReadingsByTimeRange query readings with offset, limit and time range
func ReadingsByTimeRange(parms query.Parameters, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsByTimeRange(parms.Start, parms.End, parms.Offset, parms.Limit)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	processNumericReadings(parms.Numeric, readings)

	if parms.Offset < 0 {
		return readings, 0, err // skip total count
	}
	totalCount, err = dbClient.ReadingCountByTimeRange(parms.Start, parms.End)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.BaseReading{}, totalCount, err
	}
	return readings, totalCount, err
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
func ReadingsByResourceNameAndTimeRange(resourceName string, parms query.Parameters, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if resourceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "resourceName is empty", nil)
	}
	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsByResourceNameAndTimeRange(resourceName, parms.Start, parms.End, parms.Offset, parms.Limit)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	processNumericReadings(parms.Numeric, readings)

	if parms.Offset < 0 {
		return readings, 0, err // skip total count
	}
	totalCount, err = dbClient.ReadingCountByResourceNameAndTimeRange(resourceName, parms.Start, parms.End)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.BaseReading{}, totalCount, err
	}
	return readings, totalCount, err
}

// ReadingsByDeviceNameAndResourceName query readings with offset, limit, device name and its associated resource name
func ReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, parms query.Parameters, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if deviceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if resourceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsByDeviceNameAndResourceName(deviceName, resourceName, parms.Offset, parms.Limit)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	processNumericReadings(parms.Numeric, readings)

	if parms.Offset < 0 {
		return readings, 0, err // skip total count
	}
	totalCount, err = dbClient.ReadingCountByDeviceNameAndResourceName(deviceName, resourceName)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.BaseReading{}, totalCount, err
	}
	return readings, totalCount, err
}

// ReadingsByDeviceNameAndResourceNameAndTimeRange query readings with offset, limit, device name, its associated resource name and specified time range
func ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, parms query.Parameters, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if deviceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if resourceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "resource name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)

	readingModels, err := dbClient.ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, parms.Start, parms.End, parms.Offset, parms.Limit)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	processNumericReadings(parms.Numeric, readings)

	if parms.Offset < 0 {
		return readings, 0, err // skip total count
	}
	totalCount, err = dbClient.ReadingCountByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, parms.Start, parms.End)
	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit)
	if !cont {
		return []dtos.BaseReading{}, totalCount, err
	}
	return readings, totalCount, err
}

// ReadingsByDeviceNameAndResourceNamesAndTimeRange query readings with offset, limit, device name, its associated resource name and specified time range
func ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, parms query.Parameters, dic *di.Container) (readings []dtos.BaseReading, totalCount uint32, err errors.EdgeX) {
	if deviceName == "" {
		return readings, totalCount, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	var readingModels []models.Reading
	if len(resourceNames) > 0 {
		if parms.Offset >= 0 {
			totalCount, err = dbClient.ReadingCountByDeviceNameAndResourceNamesAndTimeRange(deviceName, resourceNames, parms.Start, parms.End)
			if err != nil {
				return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
			}
			if cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit); !cont {
				return []dtos.BaseReading{}, totalCount, err
			}
		}
		readingModels, err = dbClient.ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName, resourceNames, parms.Start, parms.End, parms.Offset, parms.Limit)
	} else {
		if parms.Offset >= 0 {
			totalCount, err = dbClient.ReadingCountByDeviceNameAndTimeRange(deviceName, parms.Start, parms.End)
			if err != nil {
				return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
			}
			if cont, err := utils.CheckCountRange(totalCount, parms.Offset, parms.Limit); !cont {
				return []dtos.BaseReading{}, totalCount, err
			}
		}
		readingModels, err = dbClient.ReadingsByDeviceNameAndTimeRange(deviceName, parms.Start, parms.End, parms.Offset, parms.Limit)
	}

	if err != nil {
		return readings, totalCount, errors.NewCommonEdgeXWrapper(err)
	}
	readings, err = convertReadingModelsToDTOs(readingModels)
	processNumericReadings(parms.Numeric, readings)
	return readings, totalCount, err
}

// AsyncPurgeEvent purge events and related readings according to the retention capability.
func AsyncPurgeEvent(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	interval, err := time.ParseDuration(config.Retention.Interval)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "retention interval parse failed", err)
	}
	if interval <= 0 {
		lc.Infof("Event retention is disabled because the retention interval is `%s`.", interval)
		return nil
	}

	// purge events by auto event
	asyncPurgeReadingOnce.Do(func() {
		go func() {
			timer := time.NewTimer(interval)
			for {
				timer.Reset(interval) // since event deletion might take lots of time, restart the timer to recount the time
				select {
				case <-ctx.Done():
					lc.Info("Exiting event retention")
					return
				case <-timer.C:
					err = purgeEvents(dic)
					if err != nil {
						lc.Errorf("Failed to purge events and readings, %v", err)
						break
					}
				}
			}
		}()
	})

	return nil
}

func purgeEvents(dic *di.Container) errors.EdgeX {
	devices := container.DeviceStoreFrom(dic.Get).Devices()
	for _, device := range devices {
		for _, e := range device.AutoEvents {
			if err := purgeEventsByAutoEvent(device.Name, e, dic); err != nil {
				return errors.NewCommonEdgeXWrapper(err)
			}
		}

		// purge events that are not in auto events
		if err := purgeNoneAutoEventsByDevice(device, dic); err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("purge device '%s' events not coming from auto events", device.Name), err)
		}
	}
	return nil
}

func purgeEventsByAutoEvent(deviceName string, autoEvent models.AutoEvent, dic *di.Container) errors.EdgeX {
	config := container.ConfigurationFrom(dic.Get)
	// apply the default retention policy, when maxCap/minCap/duration are not specified or equal to zero/empty string
	// which are mentioned in the documentation
	if autoEvent.Retention.MaxCap == 0 {
		autoEvent.Retention.MaxCap = config.Retention.DefaultMaxCap
	}
	if autoEvent.Retention.MinCap == 0 {
		autoEvent.Retention.MinCap = config.Retention.DefaultMinCap
	}
	if autoEvent.Retention.Duration == "" {
		autoEvent.Retention.Duration = config.Retention.DefaultDuration
	}

	if autoEvent.Retention.MaxCap > 0 {
		isReach, err := isReachMaxCap(deviceName, autoEvent, dic)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		if !isReach {
			return nil
		}
	}

	duration, err := time.ParseDuration(autoEvent.Retention.Duration)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "auto event retention duration parse failed", err)
	}
	if duration > 0 {
		// Do time-based event retention when duration is greater than 0
		if err = timeBasedEventRetention(deviceName, autoEvent, duration, dic); err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		// otherwise do count-based retention
		if err = countBasedEventRetention(deviceName, autoEvent, dic); err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	return nil
}

func isReachMaxCap(deviceName string, autoEvent models.AutoEvent, dic *di.Container) (bool, errors.EdgeX) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)
	// Check high watermark, count events by limit instead of count all to improve performance
	count, err := dbClient.EventCountByDeviceNameAndSourceNameAndLimit(deviceName, autoEvent.SourceName, int(autoEvent.Retention.MaxCap))
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	if count < uint32(autoEvent.Retention.MaxCap) {
		lc.Debugf(
			"Skip the event retention for the auto event source `%s` of device `%s`, the event number `%d` is less than the max capacity `%d`.",
			autoEvent.SourceName, deviceName, count, autoEvent.Retention.MaxCap,
		)
		return false, nil
	}
	return true, nil
}

func timeBasedEventRetention(deviceName string, autoEvent models.AutoEvent, duration time.Duration, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)

	if autoEvent.Retention.MinCap <= 0 {
		lc.Debugf("MinCap is disabled, purge events by duration '%d' and deviceName '%s', and sourceName '%s'", duration, deviceName, autoEvent.SourceName)
		err := dbClient.DeleteEventsByAgeAndDeviceNameAndSourceName(duration.Nanoseconds(), deviceName, autoEvent.SourceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err),
				fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and duration '%s'",
					deviceName, autoEvent.SourceName, autoEvent.Retention.Duration), err)
		}
	} else {
		lc.Debugf("MinCap is '%d', purge events by duration '%d' and deviceName '%s', and sourceName '%s' to meet the minCap",
			autoEvent.Retention.MinCap, duration, deviceName, autoEvent.SourceName)
		// Find the event that the age is within the duration and use offset as minCap to keep data
		// SELECT * FROM core_data.event WHERE event.origin <= $1 and devicename=$2 and sourcename=$3 ORDER BY origin desc offset $4;
		event, err := dbClient.LatestEventByDeviceNameAndSourceNameAndAgeAndOffset(deviceName, autoEvent.SourceName, duration.Nanoseconds(), uint32(autoEvent.Retention.MinCap))
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			lc.Debugf("Skip the event retention for the auto event source '%s' of device '%s', the event number might equal or less than the minCap", deviceName, autoEvent.SourceName)
			return nil
		} else if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err),
				fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', duration '%s', and minCap '%d'",
					deviceName, autoEvent.SourceName, autoEvent.Retention.Duration, autoEvent.Retention.MinCap), err)
		}

		age := time.Now().UnixNano() - event.Origin
		err = dbClient.DeleteEventsByAgeAndDeviceNameAndSourceName(age, deviceName, autoEvent.SourceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and minCap '%d'",
				deviceName, autoEvent.SourceName, autoEvent.Retention.MinCap), err)
		}
	}
	return nil
}

func countBasedEventRetention(deviceName string, autoEvent models.AutoEvent, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)

	if autoEvent.Retention.MinCap <= 0 {
		lc.Debugf("MinCap is disabled, purge events by deviceName '%s' and sourceName '%s'", deviceName, autoEvent.SourceName)
		err := dbClient.DeleteEventsByDeviceNameAndSourceName(deviceName, autoEvent.SourceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and minCap '%d'",
				deviceName, autoEvent.SourceName, autoEvent.Retention.MinCap), err)
		}
	} else {
		lc.Debugf("MinCap is '%d', purge events by deviceName '%s' and sourceName '%s' to meet the minCap",
			autoEvent.Retention.MinCap, deviceName, autoEvent.SourceName)
		event, err := dbClient.LatestEventByDeviceNameAndSourceNameAndOffset(deviceName, autoEvent.SourceName, uint32(autoEvent.Retention.MinCap))
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			lc.Debugf("Skip the event retention for the auto event source '%s' of device '%s', the event number might equal or less than the minCap", deviceName, autoEvent.SourceName)
			return nil
		} else if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and minCap '%d'",
				deviceName, autoEvent.SourceName, autoEvent.Retention.MinCap), err)
		}
		age := time.Now().UnixNano() - event.Origin
		err = dbClient.DeleteEventsByAgeAndDeviceNameAndSourceName(age, deviceName, autoEvent.SourceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and minCap '%d'",
				deviceName, autoEvent.SourceName, autoEvent.Retention.MinCap), err)
		}
	}
	return nil
}

func purgeNoneAutoEventsByDevice(device models.Device, dic *di.Container) errors.EdgeX {
	if len(device.ProfileName) == 0 {
		return nil
	}
	sources, err := noneAutoEventSourcesByDevice(device, dic)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	for _, source := range sources {
		autoEvent := models.AutoEvent{SourceName: source} // Empty AutoEvent.Retention will apply default retention policy
		err = purgeEventsByAutoEvent(device.Name, autoEvent, dic)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	return nil
}

// noneAutoEventSourcesByDevice returns sources that are not defined in the device's auto events
func noneAutoEventSourcesByDevice(device models.Device, dic *di.Container) ([]string, errors.EdgeX) {
	profileClient := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	profile, err := profileClient.DeviceProfileByName(context.Background(), device.ProfileName)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("retrieve event sources not in auto events from the device '%s'", device.Name), err)
	}
	var sources []string
	for _, r := range profile.Profile.DeviceResources {
		// skip the device resource that is associated with any AutoEvent definitions.
		if contained := slices.ContainsFunc(device.AutoEvents, func(event models.AutoEvent) bool {
			return event.SourceName == r.Name
		}); contained {
			continue
		}
		// skip write-only device resource that do not generate events
		if r.Properties.ReadWrite == common.ReadWrite_W {
			continue
		}
		sources = append(sources, r.Name)
	}
	for _, c := range profile.Profile.DeviceCommands {
		// skip the device command that is associated with any AutoEvent definitions.
		if contained := slices.ContainsFunc(device.AutoEvents, func(event models.AutoEvent) bool {
			return event.SourceName == c.Name
		}); contained {
			continue
		}
		// skip write-only device command that do not generate events
		if c.ReadWrite == common.ReadWrite_W {
			continue
		}
		sources = append(sources, c.Name)
	}
	return sources, nil
}

func processNumericReadings(isNumeric bool, readings []dtos.BaseReading) {
	for i := range readings {
		if isNumeric {
			convertToNumeric(&readings[i])
		} else {
			convertToString(&readings[i])
		}
	}
}

// convertToNumeric converts SimpleReadding value to NumericReading value
func convertToNumeric(reading *dtos.BaseReading) {
	if len(reading.Value) == 0 {
		return
	}
	switch reading.ValueType {
	case common.ValueTypeInt8, common.ValueTypeInt16, common.ValueTypeInt32, common.ValueTypeInt64:
		reading.NumericReading.NumericValue = cast.ToInt64(reading.Value)
	case common.ValueTypeUint8, common.ValueTypeUint16, common.ValueTypeUint32, common.ValueTypeUint64:
		reading.NumericReading.NumericValue = cast.ToUint64(reading.Value)
	case common.ValueTypeFloat32, common.ValueTypeFloat64:
		reading.NumericReading.NumericValue = cast.ToFloat64(reading.Value)
	case common.ValueTypeInt8Array, common.ValueTypeInt16Array, common.ValueTypeInt32Array, common.ValueTypeInt64Array:
		arrayValue, err := parseArray(reading.Value, cast.ToInt64)
		if err != nil {
			return
		}
		reading.NumericReading.NumericValue = arrayValue
	case common.ValueTypeUint8Array, common.ValueTypeUint16Array, common.ValueTypeUint32Array, common.ValueTypeUint64Array:
		arrayValue, err := parseArray(reading.Value, cast.ToUint64)
		if err != nil {
			return
		}
		reading.NumericReading.NumericValue = arrayValue
	case common.ValueTypeFloat32Array, common.ValueTypeFloat64Array:
		arrayValue, err := parseArray(reading.Value, cast.ToFloat64)
		if err != nil {
			return
		}
		reading.NumericReading.NumericValue = arrayValue
	default:
		// skip not matched data type like bool, string
		return
	}
	reading.Value = ""
}

// convertToNumeric converts NumericReading value to SimpleReadding value
func convertToString(reading *dtos.BaseReading) {
	if reading.NumericReading.NumericValue == nil {
		return
	}
	switch reading.ValueType {
	case common.ValueTypeFloat32:
		reading.Value = strconv.FormatFloat(cast.ToFloat64(reading.NumericValue), 'e', -1, 32)
	case common.ValueTypeFloat64:
		reading.Value = strconv.FormatFloat(cast.ToFloat64(reading.NumericValue), 'e', -1, 64)
	default:
		reading.Value = fmt.Sprintf("%v", reading.NumericReading.NumericValue)
	}
	reading.NumericReading.NumericValue = nil
}

func parseArray[T any](val string, conv func(any) T) ([]T, error) {
	if len(val) <= 2 {
		return nil, fmt.Errorf("invalid array string: length must be greater than 2, got %d", len(val))
	}

	parts := strings.Split(val[1:len(val)-1], ",") // trim "[" and "]"
	result := make([]T, len(parts))
	for i, p := range parts {
		result[i] = conv(strings.TrimSpace(p))
	}

	return result, nil
}
