// addressable contains functionality for obtaining Addressable data.
package addressable

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

// AddressLoader provides functionality for obtaining Addressables.
type AddressLoader interface {
	GetAddressables() ([]contract.Addressable, error)
	GetAddressablesByAddress(address string) ([]contract.Addressable, error)
	GetAddressablesByPublisher(p string) ([]contract.Addressable, error)
	GetAddressablesByPort(p int) ([]contract.Addressable, error)
	GetAddressablesByTopic(t string) ([]contract.Addressable, error)
	GetAddressableByName(n string) (contract.Addressable, error)
	GetAddressableById(id string) (contract.Addressable, error)
}

// AddressWriter provides functionality for adding Addressables.
type AddressWriter interface {
	AddAddressable(addressable contract.Addressable) (string, error)
}

// AddressUpdater provides functionality for updating existing Addressables.
type AddressUpdater interface {
	UpdateAddressable(addressable contract.Addressable) error

	// This method is needed for testing whether addressables are still in use.
	GetDeviceServicesByAddressableId(id string) ([]contract.DeviceService, error)

	AddressLoader
	AddressWriter
}

// AddressDeleter provides functionality for removing existing Addressables.
type AddressDeleter interface {
	DeleteAddressableById(id string) error

	// This method is needed for testing whether addressables are still in use.
	GetDeviceServicesByAddressableId(id string) ([]contract.DeviceService, error)

	AddressLoader
}
