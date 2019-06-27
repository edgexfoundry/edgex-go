// addressable contains functionality for obtaining Addressable data.
package addressable

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

// AddressLoader provides functionality for obtaining Addressables.
type AddressLoader interface {
	GetAddressablesByAddress(address string) ([]contract.Addressable, error)
	GetAddressablesByPublisher(p string) ([]contract.Addressable, error)
	GetAddressablesByPort(p int) ([]contract.Addressable, error)
	GetAddressablesByTopic(t string) ([]contract.Addressable, error)
}
