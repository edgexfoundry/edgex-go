// Copyright (C) 2025 IOTech Ltd

package cache

import (
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// ActiveDeviceStore used to store active devices
type ActiveDeviceStore interface {
	Contains(deviceName string) bool
	Add(device models.Device)
	Remove(deviceName string)
	RemoveAll()
	Devices() map[string]models.Device
}

type activeDeviceStore struct {
	lc              logger.LoggingClient
	activeDeviceMap map[string]models.Device
	mutex           sync.RWMutex
}

// Contains check if device is in the active device list
func (s *activeDeviceStore) Contains(deviceName string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if _, ok := s.activeDeviceMap[deviceName]; ok {
		return true
	}
	return false
}

// Add adds device into the active device list
func (s *activeDeviceStore) Add(device models.Device) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.activeDeviceMap[device.Name]; !ok {
		s.lc.Debugf("adding %s into activeDeviceStore...", device.Name)
		s.activeDeviceMap[device.Name] = device
		return
	}
	s.lc.Infof("activeDeviceStore already contains %s. skip adding...", device.Name)
}

// Remove removes device out of the active device list.
func (s *activeDeviceStore) Remove(deviceName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.activeDeviceMap[deviceName]; !ok {
		s.lc.Infof("activeDeviceStore does not contain %s. skip removing...", deviceName)
	} else {
		s.lc.Debugf("removing %s out of activeDeviceStore...", deviceName)
		delete(s.activeDeviceMap, deviceName)
	}
}

// RemoveAll removes all device out of the active device list.
func (s *activeDeviceStore) RemoveAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(s.activeDeviceMap) == 0 {
		return
	}
	clear(s.activeDeviceMap)
}

func (s *activeDeviceStore) Devices() map[string]models.Device {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	// since the device map might change, return the copy of devices to prevent unexpected result
	clonedDevices := make(map[string]models.Device, len(s.activeDeviceMap))
	for deviceName, device := range s.activeDeviceMap {
		clonedDevices[deviceName] = device.Clone()
	}
	return clonedDevices
}

func DeviceStore(dic *di.Container) ActiveDeviceStore {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	return &activeDeviceStore{activeDeviceMap: make(map[string]models.Device), lc: lc}
}
