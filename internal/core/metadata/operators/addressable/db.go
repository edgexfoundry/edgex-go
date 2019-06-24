// addressable contains functionality for obtaining Addressable data
package addressable

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

// AddressLoader provides functionality for obtaining Addressables given an address
type AddressLoader interface {
	GetAddressablesByAddress(address string) ([]contract.Addressable, error)
}
