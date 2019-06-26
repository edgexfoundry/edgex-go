package device

import (
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type DeviceAdder interface {
	AddDevice(d contract.Device) (string, error)
	DeviceProfileLoader
	DeviceServiceLoader
}

type DeviceServiceLoader interface {
	GetDeviceServiceById(id string) (contract.DeviceService, error)
	GetDeviceServiceByName(n string) (contract.DeviceService, error)
}

type DeviceProfileLoader interface {
	GetDeviceProfileById(id string) (contract.DeviceProfile, error)
	GetDeviceProfileByName(n string) (contract.DeviceProfile, error)
}

type DeviceLoader interface {
	GetAllDevices() ([]contract.Device, error)
}
