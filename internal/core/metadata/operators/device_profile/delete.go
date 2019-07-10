package device_profile

import (
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

// DeleteExecutor handles the deletion of a device profile.
// Returns ErrDeviceProfileNotFound if a device profile could not be found with a matching ID nor name
// Returns ErrDeviceProfileInvalidState if the device profile has one or more devices or provision watchers
// associated with it
type DeleteExecutor interface {
	Execute() error
}

type deleteProfileById struct {
	db  DeviceProfileDeleter
	did string
}

// Execute performs the deletion of the device profile.
func (dpbi deleteProfileById) Execute() error {
	// Check if the device profile exists
	dp, err := dpbi.db.GetDeviceProfileById(dpbi.did)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrDeviceProfileNotFound(dpbi.did, "")
		}

		return err
	}

	// Delete the device profile
	return deleteDeviceProfile(dpbi.db, dp)
}

// NewDeleteByIDExecutor creates a new DeleteExecutor which deletes a device profile based on a device profile name.
func NewDeleteByIDExecutor(db DeviceProfileDeleter, did string) DeleteExecutor {
	return deleteProfileById{
		db:  db,
		did: did,
	}
}

type deleteProfileByName struct {
	db DeviceProfileDeleter
	dn string
}

// Execute performs the deletion of the device profile.
func (dpbn deleteProfileByName) Execute() error {
	// Check if the device profile exists
	dp, err := dpbn.db.GetDeviceProfileByName(dpbn.dn)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrDeviceProfileNotFound("", dpbn.dn)
		}

		return err
	}

	// Delete the device profile
	return deleteDeviceProfile(dpbn.db, dp)
}

// NewDeleteByNameExecutor creates a new DeleteExecutor which deletes a device profile based on a device profile ID.
func NewDeleteByNameExecutor(db DeviceProfileDeleter, dn string) DeleteExecutor {
	return deleteProfileByName{
		db: db,
		dn: dn,
	}
}

// Delete the device profile
// Make sure there are no devices still using it
// Delete the associated commands
func deleteDeviceProfile(dpd DeviceProfileDeleter, dp contract.DeviceProfile) error {
	// Check if the device profile is still in use by devices
	d, err := dpd.GetDevicesByProfileId(dp.Id)
	if err != nil {
		return err
	}

	if len(d) > 0 {
		return errors.NewErrDeviceProfileInvalidState(dp.Id, dp.Name, "Can't delete device profile, the profile is still in use by a device")
	}

	// Check if the device profile is still in use by provision watchers
	pw, err := dpd.GetProvisionWatchersByProfileId(dp.Id)
	if err != nil {
		return err
	}

	if len(pw) > 0 {
		return errors.NewErrDeviceProfileInvalidState(dp.Id, dp.Name, "Cant delete device profile, the profile is still in use by a provision watcher")

	}
	// Delete the profile
	if err := dpd.DeleteDeviceProfileById(dp.Id); err != nil {
		return err
	}

	return nil
}
