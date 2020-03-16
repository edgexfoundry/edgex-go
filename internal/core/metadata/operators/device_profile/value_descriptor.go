package device_profile

import (
	"context"
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	dataErrors "github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
)

// ValueDescriptorAdder provides the necessary functionality for creating a ValueDescriptor.
type ValueDescriptorAdder interface {
	Add(ctx context.Context, vdr *contract.ValueDescriptor) (string, error)
}

// ValueDescriptorAdder provides the necessary functionality for updating a ValueDescriptor.
type ValueDescriptorUpdater interface {
	ValueDescriptorsUsage(ctx context.Context, names []string) (map[string]bool, error)
	Add(ctx context.Context, vdr *contract.ValueDescriptor) (string, error)
	Update(ctx context.Context, vdr *contract.ValueDescriptor) error
	DeleteByName(ctx context.Context, name string) error
	ValueDescriptorForName(ctx context.Context, name string) (contract.ValueDescriptor, error)
}

// ValueDescriptorAddExecutor creates ValueDescriptor(s) via the operator pattern.
type ValueDescriptorAddExecutor interface {
	Execute() error
}

type addValueDescriptor struct {
	ctx    context.Context
	drs    []contract.DeviceResource
	client ValueDescriptorAdder
	logger logger.LoggingClient
}

// Execute creates the necessary ValueDescriptors for a set of DeviceResources.
func (a addValueDescriptor) Execute() error {
	for _, dr := range a.drs {

		desc := contract.From(dr)
		id, err := a.client.Add(a.ctx, &desc)
		if err != nil {
			a.logger.Error(fmt.Sprintf("Unable to create value descriptor: %s", err.Error()))
			return err
		}
		a.logger.Debug(fmt.Sprintf("Created Value Descriptor id: %s", id))
	}
	return nil
}

// NewAddValueDescriptorExecutor creates a new ValueDescriptorAddExecutor.
func NewAddValueDescriptorExecutor(
	ctx context.Context,
	client ValueDescriptorAdder,
	lc logger.LoggingClient,
	drs ...contract.DeviceResource) ValueDescriptorAddExecutor {

	return addValueDescriptor{
		ctx:    ctx,
		drs:    drs,
		client: client,
		logger: lc,
	}
}

// updateValueDescriptor encapsulates the data needed to update a value descriptor.
type updateValueDescriptor struct {
	ctx    context.Context
	dp     contract.DeviceProfile
	loader DeviceProfileUpdater
	client ValueDescriptorUpdater
	logger logger.LoggingClient
}

// UpdateValueDescriptorExecutor updates a value descriptor.
type UpdateValueDescriptorExecutor interface {
	Execute() error
}

// Execute updates a value descriptor with the provided information.
func (u updateValueDescriptor) Execute() error {
	// Get pre-existing device profile so we can determine what to do with the device resources provided in the update.
	// For example, update/create/delete.
	persistedDeviceProfile, err := u.loader.GetDeviceProfileByName(u.dp.Name)
	if err != nil {
		return err
	}

	devices, err := u.loader.GetDevicesByProfileId(persistedDeviceProfile.Id)
	if err != nil {
		return err
	}

	// Verify the associated DeviceProfile is in an upgradeable state, which means that no devices are associated with
	// it.
	if len(devices) > 0 {
		var associatedDeviceNames []string
		for _, d := range devices {
			associatedDeviceNames = append(associatedDeviceNames, d.Name)
		}

		return errors.NewErrDeviceProfileInvalidState(
			persistedDeviceProfile.Id,
			persistedDeviceProfile.Name,
			fmt.Sprintf("The DeviceProfile is in use by Device(s):[%s]", strings.Join(associatedDeviceNames, ",")))
	}

	// Get names of all the device resources so we can check the valueDescriptorUsage with one call to Core-Data.
	var persistedDeviceResourceNames []string
	for _, persistedDeviceResource := range persistedDeviceProfile.DeviceResources {
		persistedDeviceResourceNames = append(persistedDeviceResourceNames, persistedDeviceResource.Name)
	}

	// Check if any of the ValueDescriptors associated with the DeviceResources are in use.
	// If so return an error stating all the ValueDescriptors which are in use.
	valueDescriptorUsage, err := u.client.ValueDescriptorsUsage(u.ctx, persistedDeviceResourceNames)
	if err != nil {
		return err
	}

	var inUseValueDescriptors []string
	for name, inUse := range valueDescriptorUsage {
		if inUse {
			inUseValueDescriptors = append(inUseValueDescriptors, name)
		}
	}

	if len(inUseValueDescriptors) > 0 {
		return dataErrors.NewErrValueDescriptorsInUse(inUseValueDescriptors)
	}

	// Based on the DeviceProfile as it is before the update, determine which operation needs to be applied to each
	// ValueDescriptor to get it in the desired state which is the information passed to the update command.
	create, update, deleted := determineValueDescriptor(persistedDeviceProfile, u.dp)

	// Execute the necessary operations to get the DeviceProfile to the desired state.
	for _, d := range deleted {
		err = u.client.DeleteByName(u.ctx, d.Name)
		if err != nil {
			return err
		}

	}

	for _, up := range update {
		v, err := u.client.ValueDescriptorForName(u.ctx, up.Name)
		if err != nil {
			return err
		}

		up.Id = v.Id
		err = u.client.Update(u.ctx, &up)
		if err != nil {
			return err
		}
	}

	for _, c := range create {
		_, err = u.client.Add(u.ctx, &c)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewUpdateValueDescriptorExecutor creates a UpdateValueDescriptorExecutor which will update ValueDescriptors.
func NewUpdateValueDescriptorExecutor(
	ctx context.Context,
	dp contract.DeviceProfile,
	loader DeviceProfileUpdater,
	client ValueDescriptorUpdater,
	logger logger.LoggingClient) UpdateValueDescriptorExecutor {

	return updateValueDescriptor{
		dp:     dp,
		loader: loader,
		client: client,
		logger: logger,
		ctx:    ctx,
	}
}

// determineValueDescriptor creates and partitions the ValueDescriptors which need to be changed given the
// existingDeviceProfile state and the desired updatedDeviceProfile state.
//
//  Returns created - a slice of ValueDescriptors which need to be created.
//	Returns update - a slice of ValueDescriptors which need to be updated.
//	Returns deleted - a slice of ValueDescriptors which need to be deleted.
func determineValueDescriptor(
	existingDeviceProfile,
	updatedDeviceProfile contract.DeviceProfile) (create, update, deleted []contract.ValueDescriptor) {

	existingValueDescriptors := map[string]contract.ValueDescriptor{}
	updatedValueDescriptors := map[string]contract.ValueDescriptor{}

	var vd contract.ValueDescriptor
	// Extract the names from the DeviceResources.
	for _, dr := range existingDeviceProfile.DeviceResources {
		vd = contract.From(dr)
		existingValueDescriptors[vd.Name] = vd
	}

	for _, dr := range updatedDeviceProfile.DeviceResources {
		vd = contract.From(dr)
		updatedValueDescriptors[vd.Name] = vd
	}

	// Determine which ValueDescriptors need to be update, created or deleted.
	for k, v := range updatedValueDescriptors {
		// If updatedDeviceProfile dr's are in existingDeviceProfile then update
		if _, ok := existingValueDescriptors[k]; ok {
			update = append(update, v)
		} else {
			create = append(create, v)
		}
	}

	for k, v := range existingValueDescriptors {
		if _, ok := updatedValueDescriptors[k]; !ok {
			deleted = append(deleted, v)
		}
	}

	return
}
