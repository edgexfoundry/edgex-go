// addressable contains functionality for obtaining Addressable data
package addressable

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

// AddressablesRetreiver provide functionality for retrieving addresses from an underlying data-store
type AddressablesRetreiver interface {
	Execute() ([]contract.Addressable, error)
}

// addressLoader loads address by way of the operator pattern
type addressLoader struct {
	database AddressLoader
	address  string
}

// Execute retrieves the addressables from the underlying data-store
func (n addressLoader) Execute() ([]contract.Addressable, error) {
	return n.database.GetAddressablesByAddress(n.address)
}

// NewAddressLoader creates an AddressablesRetreiver
func NewAddressLoader(db AddressLoader, address string) AddressablesRetreiver {
	return addressLoader{
		database: db,
		address:  address}
}
