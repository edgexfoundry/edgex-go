package device_profile

import (
	"strconv"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/mitchellh/copystructure"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
)

// UpdateDeviceProfileExecutor provides functionality for updating and persisting device profiles.
// Returns NewErrDuplicateName
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
	var to contract.DeviceProfile
	// First try with ID
	to, err := n.database.GetDeviceProfileById(n.dp.Id)
	if err != nil {
		// Try with name
		to, err = n.database.GetDeviceProfileByName(n.dp.Name)
		if err != nil {
			return contract.DeviceProfile{}, errors.NewErrDeviceProfileNotFound(n.dp.Id, n.dp.Name)
		}
	}

	d, err := n.database.GetDevicesByProfileId(to.Id)
	if err != nil {
		return contract.DeviceProfile{}, err
	}

	if len(d) > 0 {
		return contract.DeviceProfile{}, errors.NewErrDeviceProfileInvalidState(n.dp.Id, n.dp.Name, strconv.Itoa(len(d))+" devices are associated with the device profile")
	}

	p, err := n.database.GetProvisionWatchersByProfileId(to.Id)
	if err != nil {
		return contract.DeviceProfile{}, err
	}

	if len(p) > 0 {

		return contract.DeviceProfile{}, errors.NewErrDeviceProfileInvalidState(n.dp.Id, n.dp.Name, strconv.Itoa(len(p))+" provision watchers are associated with the device profile")
	}

	// Names must be unique for each device profile
	if err := n.checkDuplicateProfileNames(to); err != nil {
		return contract.DeviceProfile{}, err
	}

	// Update the device profile fields based on the passed JSON
	dp, err := n.updateDeviceProfileFields(n.dp, to)
	if err != nil {
		return contract.DeviceProfile{}, err
	}

	if err := n.database.UpdateDeviceProfile(dp); err != nil {
		return contract.DeviceProfile{}, err
	}

	return dp, nil
}

// NewUpdateDeviceProfileExecutor creates an UpdateDeviceProfileExecutor.
func NewUpdateDeviceProfileExecutor(db DeviceProfileUpdater, dp contract.DeviceProfile) UpdateDeviceProfileExecutor {
	return updateDeviceProfile{
		database: db,
		dp:       dp}
}

// Update the fields of the device profile and returns a new copy of DeviceProfile with the updated fields.
// to - the device profile that was already in Mongo (whose fields we're updating)
// from - the device profile that was passed in with the request
func (n updateDeviceProfile) updateDeviceProfileFields(from contract.DeviceProfile, to contract.DeviceProfile) (contract.DeviceProfile, error) {
	cpy, err := copystructure.Copy(to)
	if err != nil {
		return contract.DeviceProfile{}, err
	}

	var dp contract.DeviceProfile
	dp = cpy.(contract.DeviceProfile)
	if from.Description != "" {
		dp.Description = from.Description
	}

	if from.Labels != nil {
		dp.Labels = from.Labels
	}

	if from.Manufacturer != "" {
		dp.Manufacturer = from.Manufacturer
	}

	if from.Model != "" {
		dp.Model = from.Model
	}

	if from.Origin != 0 {
		dp.Origin = from.Origin
	}

	if from.Name != "" {
		dp.Name = from.Name
	}

	if from.DeviceResources != nil {
		dp.DeviceResources = from.DeviceResources
	}

	if from.DeviceCommands != nil {
		dp.DeviceCommands = from.DeviceCommands
	}

	if from.CoreCommands != nil {
		dp.CoreCommands = from.CoreCommands
	}

	return dp, nil
}

// Check for duplicate names in device profiles
func (n updateDeviceProfile) checkDuplicateProfileNames(dp contract.DeviceProfile) error {
	profiles, err := n.database.GetAllDeviceProfiles()
	if err != nil {
		return err
	}

	for _, p := range profiles {
		if p.Name == dp.Name && p.Id != n.dp.Id {
			return errors.NewErrDuplicateName("Duplicate device profile name within the device profile")
		}
	}

	return nil
}
