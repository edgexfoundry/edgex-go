//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	dbModels "github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/models"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// AsyncPurgeEvent purge events and related readings according to the retention capability.
func (a *CoreDataApp) AsyncPurgeEvent(ctx context.Context, dic *di.Container) errors.EdgeX {
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
	a.asyncPurgeEventOnce.Do(func() {
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
					}
				}
			}
		}()
	})

	return nil
}

func purgeEvents(dic *di.Container) errors.EdgeX {
	deviceInfoCache := container.DeviceInfoCacheFrom(dic.Get)
	deviceInfoMap := deviceInfoCache.CloneDeviceInfoMapWithSourceName()

	ep := newEventPurgeExecutor(deviceInfoMap)

	// Purge events matched conditions defined in the retention policies of auto events
	if err := ep.purgeEventsByAutoEvent(dic); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Purge the remaining device & none-device events not covered by retention policies and auto events
	if err := ep.purgeEventsByDeviceInfo(dic); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

type eventPurgeExecutor struct {
	deviceInfoMap map[int]dbModels.DeviceInfo
	mutex         sync.Mutex
}

func newEventPurgeExecutor(deviceInfoMap map[int]dbModels.DeviceInfo) *eventPurgeExecutor {
	return &eventPurgeExecutor{deviceInfoMap: deviceInfoMap}
}

// cullMatchingDeviceInfo evaluates all entries in the deviceInfoMap from cache against the provided deviceName and sourceName regular expression patterns.
// If both deviceName and sourceName patterns match, the entry is removed from deviceInfoMap.
func (ep *eventPurgeExecutor) cullMatchingDeviceInfo(deviceName, sourceName string, lc logger.LoggingClient) {
	ep.mutex.Lock()
	defer ep.mutex.Unlock()

	for id, deviceInfo := range ep.deviceInfoMap {
		lc.Debugf("Checking deviceInfo id '%d' with device '%s', source '%s' from the purge event list", id, deviceName, deviceInfo.SourceName)

		deviceMatched, _ := regexp.MatchString(deviceName, deviceInfo.DeviceName)
		if deviceMatched {
			sourceMatched, _ := regexp.MatchString(sourceName, deviceInfo.SourceName)
			if sourceMatched {
				lc.Debugf("Removing deviceInfo id '%d' with device '%s', source '%s' from the purge event list", id, deviceName, deviceInfo.SourceName)
				delete(ep.deviceInfoMap, id)
			}
		}
	}
}

// purgeEventsByAutoEvent iterates over the auto events defined for each device and removes any events that match the retention policy conditions specified in the auto event.
func (ep *eventPurgeExecutor) purgeEventsByAutoEvent(dic *di.Container) errors.EdgeX {
	config := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	devices := container.DeviceStoreFrom(dic.Get).Devices()

	lc.Debug("Starting purge events by auto events ......")

	for _, device := range devices {
		for _, autoEvent := range device.AutoEvents {
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

			ep.cullMatchingDeviceInfo(device.Name, autoEvent.SourceName, lc)

			err := handleEventRetention(
				device.Name,
				autoEvent.SourceName,
				autoEvent.Retention.MaxCap,
				autoEvent.Retention.MinCap,
				autoEvent.Retention.Duration,
				dic,
			)
			if err != nil {
				lc.Errorf("failed to execute event retention for device '%s', source '%s', maxCap: '%d', minCap: '%d', duration '%s': %v",
					device.Name, autoEvent.SourceName, autoEvent.Retention.MaxCap, autoEvent.Retention.MinCap, autoEvent.Retention.Duration, err)
				continue
			}
		}
	}
	return nil
}

// purgeEventsByDeviceInfo traverses the device info list and applies the retention settings from config
// based on the device name and source name in each device info entry.
func (ep *eventPurgeExecutor) purgeEventsByDeviceInfo(dic *di.Container) errors.EdgeX {
	config := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	lc.Debug("Starting purge events by retention policies on default config level ......")

	for _, deviceInfo := range ep.deviceInfoMap {
		err := handleEventRetention(
			deviceInfo.DeviceName,
			deviceInfo.SourceName,
			config.Retention.DefaultMaxCap,
			config.Retention.DefaultMinCap,
			config.Retention.DefaultDuration,
			dic,
		)
		if err != nil {
			lc.Errorf(
				"failed to execute event retention for device info with device name: '%s', source name: '%s', error: %v",
				deviceInfo.DeviceName,
				deviceInfo.SourceName,
				err,
			)
			continue
		}
	}
	return nil
}

func handleEventRetention(
	deviceName string,
	sourceName string,
	maxCap int64,
	minCap int64,
	durationStr string,
	dic *di.Container,
) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	if maxCap > 0 {
		isReach, err := isReachMaxCap(deviceName, sourceName, maxCap, dic)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		if !isReach {
			lc.Debugf(
				"Skip the event retention for the device with source name: '%s' because max capacity '%d' not reached",
				sourceName,
				maxCap,
			)
			return nil
		}
	}

	validDuration, duration := common.ParseDurationWithDay(durationStr)
	if !validDuration {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("retention duration '%s' parse failed", durationStr), nil)
	}

	lc.Debugf("Starting execute event retention for device name: '%s', source name: '%s', maxCap: %d, minCap: %d, duration: %s",
		deviceName, sourceName, maxCap, minCap, durationStr)

	if duration > 0 {
		return timeBasedEventRetention(deviceName, sourceName, minCap, duration, dic)
	}
	return countBasedEventRetention(deviceName, sourceName, minCap, dic)
}

func isReachMaxCap(deviceName, sourceName string, maxCap int64, dic *di.Container) (bool, errors.EdgeX) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)
	// Check high watermark, count events by limit instead of count all to improve performance
	count, err := dbClient.EventCountByDeviceNameAndSourceNameAndLimit(deviceName, sourceName, int(maxCap))
	if err != nil {
		return false, errors.NewCommonEdgeXWrapper(err)
	}
	if count < uint32(maxCap) {
		lc.Debugf(
			"Skip the event retention for the source name `%s` of device `%s`, the event number `%d` is less than the max capacity `%d`.",
			sourceName, deviceName, count, maxCap,
		)
		return false, nil
	}
	return true, nil
}

func timeBasedEventRetention(deviceName, sourceName string, minCap int64, duration time.Duration, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)

	if minCap <= 0 {
		lc.Debugf("MinCap is disabled, purge events by duration '%d' and deviceName '%s', and sourceName '%s'", duration, deviceName, sourceName)
		err := dbClient.DeleteEventsByAgeAndDeviceNameAndSourceName(duration.Nanoseconds(), deviceName, sourceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err),
				fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and duration '%s'",
					deviceName, sourceName, duration), err)
		}
	} else {
		lc.Debugf("MinCap is '%d', purge events by duration '%d' and deviceName '%s', and sourceName '%s' to meet the minCap",
			minCap, duration, deviceName, sourceName)
		// Find the event that the age is within the duration and use offset as minCap to keep data
		// SELECT * FROM core_data.event WHERE event.origin <= $1 and devicename=$2 and sourcename=$3 ORDER BY origin desc offset $4;
		event, err := dbClient.LatestEventByDeviceNameAndSourceNameAndAgeAndOffset(deviceName, sourceName, duration.Nanoseconds(), uint32(minCap))
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			lc.Debugf("Skip the event retention for the event source '%s' of device '%s', the event number might equal or less than the minCap", sourceName, deviceName)
			return nil
		} else if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err),
				fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', duration '%s', and minCap '%d'",
					deviceName, sourceName, duration, minCap), err)
		}

		age := time.Now().UnixNano() - event.Origin
		err = dbClient.DeleteEventsByAgeAndDeviceNameAndSourceName(age, deviceName, sourceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and minCap '%d'",
				deviceName, sourceName, minCap), err)
		}
	}
	return nil
}

func countBasedEventRetention(deviceName, sourceName string, minCap int64, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)

	if minCap <= 0 {
		lc.Debugf("MinCap is disabled, purge events by deviceName '%s' and sourceName '%s'", deviceName, sourceName)
		err := dbClient.DeleteEventsByDeviceNameAndSourceName(deviceName, sourceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and minCap '%d'",
				deviceName, sourceName, minCap), err)
		}
	} else {
		lc.Debugf("MinCap is '%d', purge events by deviceName '%s' and sourceName '%s' to meet the minCap",
			minCap, deviceName, sourceName)
		event, err := dbClient.LatestEventByDeviceNameAndSourceNameAndOffset(deviceName, sourceName, uint32(minCap))
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			lc.Debugf("Skip the event retention for the auto event source '%s' of device '%s', the event number might equal or less than the minCap", sourceName, deviceName)
			return nil
		} else if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and minCap '%d'",
				deviceName, sourceName, minCap), err)
		}
		age := time.Now().UnixNano() - event.Origin
		err = dbClient.DeleteEventsByAgeAndDeviceNameAndSourceName(age, deviceName, sourceName)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to delete events and readings with specific deviceName '%s', sourceName '%s', and minCap '%d'",
				deviceName, sourceName, minCap), err)
		}
	}
	return nil
}
