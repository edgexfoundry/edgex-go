package command

import (
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type CommandLoader interface {
	GetCommandsByDeviceId(did string) ([]contract.Command, error)
	GetAllCommands() ([]contract.Command, error)
	GetCommandById(id string) (contract.Command, error)
	GetCommandsByName(id string) ([]contract.Command, error)
}
