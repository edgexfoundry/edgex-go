// addressable contains functionality for obtaining Addressable data.
package addressable

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type AddressableAllExecutor interface {
	Execute() ([]contract.Addressable, error)
}

type addressableLoadAll struct {
	config   config.ServiceInfo
	database AddressLoader
	logger   logger.LoggingClient
}

func NewAddressableLoadAll(cfg config.ServiceInfo, db AddressLoader, log logger.LoggingClient) AddressableAllExecutor {
	return addressableLoadAll{config: cfg, database: db, logger: log}
}

func (op addressableLoadAll) Execute() (addressables []contract.Addressable, err error) {
	addressables, err = op.database.GetAddressables()
	if err != nil {
		op.logger.Error(err.Error())
		return nil, err
	}
	if len(addressables) > op.config.MaxResultCount {
		err = errors.NewErrLimitExceeded(op.config.MaxResultCount)
		op.logger.Error(err.Error())
		return nil, err
	}
	return
}

type IdExecutor interface {
	Execute() (contract.Addressable, error)
}

type addressLoadById struct {
	database AddressLoader
	id       string
}

func (op addressLoadById) Execute() (contract.Addressable, error) {
	return op.database.GetAddressableById(op.id)
}

func NewIdExecutor(db AddressLoader, id string) IdExecutor {
	return addressLoadById{
		database: db,
		id:       id,
	}
}

type NameExecutor interface {
	Execute() (contract.Addressable, error)
}

type addressLoadByName struct {
	database AddressLoader
	name     string
}

func (op addressLoadByName) Execute() (contract.Addressable, error) {
	return op.database.GetAddressableByName(op.name)
}

func NewNameExecutor(db AddressLoader, name string) NameExecutor {
	return addressLoadByName{
		database: db,
		name:     name,
	}
}

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

// TopicExecutor provides functionality for retrieving addresses from an underlying data-store.
type TopicExecutor interface {
	Execute() ([]contract.Addressable, error)
}

// addressByTopicLoader loads address by way of the operator pattern.
type addressByTopicLoader struct {
	database AddressLoader
	topic    string
}

// Execute retrieves the addressables from the underlying data-store.
func (n addressByTopicLoader) Execute() ([]contract.Addressable, error) {
	return n.database.GetAddressablesByTopic(n.topic)
}

// NewTopicExecutor creates a TopicExecutor.
func NewTopicExecutor(db AddressLoader, topic string) TopicExecutor {
	return addressByTopicLoader{
		database: db,
		topic:    topic}
}
