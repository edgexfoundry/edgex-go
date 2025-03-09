// Copyright (C) 2025 IOTech Ltd

package cache

import (
	"crypto/md5" // #nosec
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/postgres/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// DeviceInfoIdCache used to store device_info id
type DeviceInfoIdCache interface {
	Get(deviceInfo models.DeviceInfo) (int, bool)
	Contains(deviceInfo models.DeviceInfo) bool
	Add(deviceInfo models.DeviceInfo)
	Remove(deviceInfo models.DeviceInfo)
	RemoveAll()
}

type deviceInfoIdCache struct {
	lc                 logger.LoggingClient
	deviceInfoKeyIdMap map[string]int
	mutex              sync.RWMutex
}

// Get returns deviceInfoId from the cache
func (s *deviceInfoIdCache) Get(deviceInfo models.DeviceInfo) (int, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := s.cacheKey(deviceInfo)
	if id, ok := s.deviceInfoKeyIdMap[key]; ok {
		return id, true
	}
	return 0, false
}

func (s *deviceInfoIdCache) cacheKey(deviceInfo models.DeviceInfo) string {
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

// Contains check if device_info key is in the cache
func (s *deviceInfoIdCache) Contains(deviceInfo models.DeviceInfo) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	key := s.cacheKey(deviceInfo)
	if _, ok := s.deviceInfoKeyIdMap[key]; ok {
		return true
	}
	return false
}

// Add adds deviceInfoId into the cache
func (s *deviceInfoIdCache) Add(deviceInfo models.DeviceInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := s.cacheKey(deviceInfo)
	if _, ok := s.deviceInfoKeyIdMap[key]; !ok {
		s.lc.Debugf("adding %d deviceInfoId into cache...", deviceInfo.Id)
		s.deviceInfoKeyIdMap[key] = deviceInfo.Id
		return
	}
	s.lc.Infof("deviceInfoIdCache already contains %d. skip adding...", deviceInfo.Id)
}

// Remove removes deviceInfoId out of the cache.
func (s *deviceInfoIdCache) Remove(deviceInfo models.DeviceInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := s.cacheKey(deviceInfo)
	if _, ok := s.deviceInfoKeyIdMap[key]; !ok {
		s.lc.Infof("deviceInfoIdCache does not contain %s. skip removing...", deviceInfo.Id)
	} else {
		s.lc.Debugf("removing %d out of deviceInfoIdCache...", deviceInfo.Id)
		delete(s.deviceInfoKeyIdMap, key)
	}
}

// RemoveAll removes all deviceInfoId out of the cache.
func (s *deviceInfoIdCache) RemoveAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(s.deviceInfoKeyIdMap) == 0 {
		return
	}
	clear(s.deviceInfoKeyIdMap)
}

func NewDeviceInfoIdCache(lc logger.LoggingClient) DeviceInfoIdCache {
	return &deviceInfoIdCache{deviceInfoKeyIdMap: make(map[string]int), lc: lc}
}
