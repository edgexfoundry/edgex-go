// addressable contains functionality for obtaining Addressable data.
package addressable

import contract "github.com/edgexfoundry/go-mod-core-contracts/models"

// AddressExecutor provides functionality for retrieving addresses from an underlying data-store.
type AddressExecutor interface {
	Execute() ([]contract.Addressable, error)
}

// addressLoadByAddress loads address by way of the operator pattern.
type addressLoadByAddress struct {
	database AddressLoader
	address  string
}

// Execute retrieves the addressables from the underlying data-store.
func (n addressLoadByAddress) Execute() ([]contract.Addressable, error) {
	return n.database.GetAddressablesByAddress(n.address)
}

// NewAddressExecutor creates an AddressExecutor.
func NewAddressExecutor(db AddressLoader, address string) AddressExecutor {
	return addressLoadByAddress{
		database: db,
		address:  address}
}

// PublisherExecutor provides functionality for retrieving addresses from an underlying data-store.
type PublisherExecutor interface {
	Execute() ([]contract.Addressable, error)
}

// addressLoadByPublisher loads address by way of the operator pattern.
type addressLoadByPublisher struct {
	database  AddressLoader
	publisher string
}

// Execute retrieves the addressables from the underlying data-store.
func (p addressLoadByPublisher) Execute() ([]contract.Addressable, error) {
	return p.database.GetAddressablesByPublisher(p.publisher)
}

// NewPublisherExecutor creates a PublisherExecutor.
func NewPublisherExecutor(db AddressLoader, publisher string) PublisherExecutor {
	return addressLoadByPublisher{
		database:  db,
		publisher: publisher,
	}
}

// PortExecutor provides functionality for retrieving addresses from an underlying data-store.
type PortExecutor interface {
	Execute() ([]contract.Addressable, error)
}

// addressByPortLoader loads address by way of the operator pattern.
type addressByPortLoader struct {
	database AddressLoader
	port     int
}

// Execute retrieves the addressables from the underlying data-store.
func (n addressByPortLoader) Execute() ([]contract.Addressable, error) {
	return n.database.GetAddressablesByPort(n.port)
}

// NewPortExecutor creates a PortExecutor.
func NewPortExecutor(db AddressLoader, port int) PortExecutor {
	return addressByPortLoader{
		database: db,
		port:     port}
}
