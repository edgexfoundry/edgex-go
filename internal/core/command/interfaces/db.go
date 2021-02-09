package interfaces

import (
	contract "github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type DBClient interface {
	CloseSession()
	GetAllCommands() ([]contract.Command, error)
	GetCommandById(id string) (contract.Command, error)
	GetCommandsByName(id string) ([]contract.Command, error)
	GetCommandsByDeviceId(id string) ([]contract.Command, error)
	GetCommandByNameAndDeviceId(cname string, did string) (contract.Command, error)
}
