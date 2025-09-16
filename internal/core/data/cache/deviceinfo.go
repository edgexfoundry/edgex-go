//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"crypto/md5" // #nosec
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/models"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// DeviceInfoCache used to store device_info id
type DeviceInfoCache interface {
	GetDeviceInfoId(deviceInfo models.DeviceInfo) (int, bool)
	CloneDeviceInfoMapWithSourceName() map[int]models.DeviceInfo
	Add(deviceInfo models.DeviceInfo)
	Remove(deviceInfo models.DeviceInfo)
}

type deviceInfoCache struct {
	lc                 logger.LoggingClient
	deviceInfoKeyIdMap map[string]int
	deviceInfoMap      map[int]models.DeviceInfo
	mutex              sync.RWMutex
}

func NewDeviceInfoCache(dic *di.Container, deviceInfos []models.DeviceInfo) DeviceInfoCache {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	cache := &deviceInfoCache{
		lc:                 lc,
		deviceInfoMap:      make(map[int]models.DeviceInfo, len(deviceInfos)),
		deviceInfoKeyIdMap: make(map[string]int, len(deviceInfos)),
	}

	for _, deviceInfo := range deviceInfos {
		// Add id/deviceInfo to deviceInfoMap
		if _, exists := cache.deviceInfoMap[deviceInfo.Id]; exists {
			lc.Warnf("duplicate DeviceInfo ID %d exists in deviceInfoMap, skip adding the deviceInfo id to cache", deviceInfo.Id)
		} else {
			cache.deviceInfoMap[deviceInfo.Id] = deviceInfo
		}

		// Add key/deviceInfoId to deviceInfoKeyIdMap
		key := cache.cacheKey(deviceInfo)
		if cachedId, exists := cache.deviceInfoKeyIdMap[key]; exists {
			lc.Warnf("duplicate cache key %q detected for both DeviceInfo ID %d (exists) and %d (new), skip adding the new deviceInfo id to cache", key, cachedId, deviceInfo.Id)
		} else {
			cache.deviceInfoKeyIdMap[key] = deviceInfo.Id
		}
	}

	return cache
}

// GetDeviceInfoId returns deviceInfoId from the cache
func (s *deviceInfoCache) GetDeviceInfoId(deviceInfo models.DeviceInfo) (int, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := s.cacheKey(deviceInfo)
	if id, ok := s.deviceInfoKeyIdMap[key]; ok {
		return id, true
	}
	return 0, false
}

// CloneDeviceInfoMapWithSourceName returns device info map with source name from the cache
func (s *deviceInfoCache) CloneDeviceInfoMapWithSourceName() map[int]models.DeviceInfo {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	newDeviceInfoMap := make(map[int]models.DeviceInfo)
	for id, deviceInfo := range s.deviceInfoMap {
		// Only copy the deviceInfo with SourceName into the new map
		if deviceInfo.SourceName != "" {
			newDeviceInfoMap[id] = deviceInfo
		}
	}
	return newDeviceInfoMap
}

func (s *deviceInfoCache) cacheKey(deviceInfo models.DeviceInfo) string {
	data := [][]byte{
		[]byte(deviceInfo.DeviceName), []byte(deviceInfo.ProfileName), []byte(deviceInfo.SourceName),
		[]byte(deviceInfo.ResourceName), []byte(deviceInfo.ValueType), []byte(deviceInfo.Units),
		[]byte(deviceInfo.MediaType), []byte(fmt.Sprintf("%v", deviceInfo.Tags)),
	}
	var byteData []byte
	for _, b := range data {
		byteData = append(byteData, b...)
	}
	return fmt.Sprintf("%x", md5.Sum(byteData)) // #nosec
}

// Add adds deviceInfoId and DeviceInfo data into the cache
func (s *deviceInfoCache) Add(deviceInfo models.DeviceInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := s.cacheKey(deviceInfo)
	if _, ok := s.deviceInfoKeyIdMap[key]; !ok {
		s.lc.Debugf("adding deviceInfoId with id '%d' into cache...", deviceInfo.Id)
		s.deviceInfoKeyIdMap[key] = deviceInfo.Id
	} else {
		s.lc.Infof("deviceInfoIdCache already contains %d. skip adding...", deviceInfo.Id)
	}

	if _, ok := s.deviceInfoMap[deviceInfo.Id]; !ok {
		s.lc.Debugf("adding deviceInfo with id '%d' into deviceInfoMap cache...", deviceInfo.Id)
		s.deviceInfoMap[deviceInfo.Id] = deviceInfo
	} else {
		s.lc.Infof("deviceInfoMap cache already contains deviceInfo id '%d'. skip adding...", deviceInfo.Id)
	}
}

// Remove removes deviceInfoId and DeviceInfo data out of the cache.
func (s *deviceInfoCache) Remove(deviceInfo models.DeviceInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.cacheKey(deviceInfo)
	if _, ok := s.deviceInfoKeyIdMap[key]; !ok {
		s.lc.Infof("deviceInfoIdCache does not contain %d. skip removing...", deviceInfo.Id)
	} else {
		s.lc.Debugf("removing deviceInfo id '%d' out of deviceInfo cache...", deviceInfo.Id)
		delete(s.deviceInfoKeyIdMap, key)
	}

	if _, ok := s.deviceInfoMap[deviceInfo.Id]; !ok {
		s.lc.Infof("deviceInfo with id '%d' not exists in deviceInfoMap cache, skip removing.", deviceInfo.Id)
		s.deviceInfoMap[deviceInfo.Id] = deviceInfo
	} else {
		s.lc.Debugf("Removing deviceInfo with id '%d' from deviceInfoMap cache ...", deviceInfo.Id)
		delete(s.deviceInfoMap, deviceInfo.Id)
	}
}
