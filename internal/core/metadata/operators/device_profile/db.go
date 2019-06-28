package device_profile

import (
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/operators/device"
)

type DeviceProfileUpdater interface {
	GetAllDeviceProfiles() ([]contract.DeviceProfile, error)
	GetProvisionWatchersByProfileId(id string) ([]contract.ProvisionWatcher, error)
	UpdateDeviceProfile(dp contract.DeviceProfile) error
	device.DeviceLoader
	device.DeviceProfileLoader
}
