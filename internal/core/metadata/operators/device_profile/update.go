package device_profile

import (
	"strconv"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
)

// UpdateDeviceProfileExecutor provides functionality for updating and persisting device profiles.
// Returns ErrDeviceProfileNotFound if a device profile could not be found with a matching ID nor name
// Returns ErrDeviceProfileInvalidState if the device profile has one or more devices or provision watchers
// associated with it
type UpdateDeviceProfileExecutor interface {
	Execute() (contract.DeviceProfile, error)
}

// updateDeviceProfile updates device profiles by way of the operator pattern.
type updateDeviceProfile struct {
	database DeviceProfileUpdater
	dp       contract.DeviceProfile
}

// Execute updates and persists the device profile.
func (n updateDeviceProfile) Execute() (contract.DeviceProfile, error) {
	// Check if the Device Profile exists
	var existingDeviceProfile contract.DeviceProfile
	// First try with ID
	existingDeviceProfile, err := n.database.GetDeviceProfileById(n.dp.Id)
	if err != nil {
		// Try with name
		existingDeviceProfile, err = n.database.GetDeviceProfileByName(n.dp.Name)
		if err != nil {
			return contract.DeviceProfile{}, errors.NewErrDeviceProfileNotFound(n.dp.Id, n.dp.Name)
		}
	}

	d, err := n.database.GetDevicesByProfileId(existingDeviceProfile.Id)
	if err != nil {
		return contract.DeviceProfile{}, err
	}

	if len(d) > 0 {
		return contract.DeviceProfile{}, errors.NewErrDeviceProfileInvalidState(n.dp.Id, n.dp.Name, strconv.Itoa(len(d))+" devices are associated with the device profile")
	}

	p, err := n.database.GetProvisionWatchersByProfileId(existingDeviceProfile.Id)
	if err != nil {
		return contract.DeviceProfile{}, err
	}

	if len(p) > 0 {

		return contract.DeviceProfile{}, errors.NewErrDeviceProfileInvalidState(n.dp.Id, n.dp.Name, strconv.Itoa(len(p))+" provision watchers are associated with the device profile")
	}

	n.dp.Id = existingDeviceProfile.Id
	if err := n.database.UpdateDeviceProfile(n.dp); err != nil {
		return contract.DeviceProfile{}, err
	}

	return n.dp, nil
}

// NewUpdateDeviceProfileExecutor creates an UpdateDeviceProfileExecutor.
func NewUpdateDeviceProfileExecutor(db DeviceProfileUpdater, dp contract.DeviceProfile) UpdateDeviceProfileExecutor {
	return updateDeviceProfile{
		database: db,
		dp:       dp}
}
