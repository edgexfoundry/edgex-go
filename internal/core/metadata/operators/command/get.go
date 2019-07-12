package command

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

//Get commands by device id
type DeviceIdExecutor interface {
	Execute() ([]contract.Command, error)
}

type commandLoadByDeviceId struct {
	database CommandLoader
	deviceId string
}

func NewDeviceIdExecutor(db CommandLoader, deviceId string) DeviceIdExecutor {
	return commandLoadByDeviceId{database: db, deviceId: deviceId}
}

func (op commandLoadByDeviceId) Execute() (commands []contract.Command, err error) {
	return op.database.GetCommandsByDeviceId(op.deviceId)
}

//Get All commands
type CommandAllExecutor interface {
	Execute() ([]contract.Command, error)
}

type commandLoadAll struct {
	config   config.ServiceInfo
	database CommandLoader
}

func NewCommandLoadAll(cfg config.ServiceInfo, db CommandLoader) CommandAllExecutor {
	return commandLoadAll{config: cfg, database: db}
}

func (op commandLoadAll) Execute() (cmds []contract.Command, err error) {
	cmds, err = op.database.GetAllCommands()
	if err != nil {
		return
	}

	if len(cmds) > op.config.MaxResultCount {
		err = errors.NewErrLimitExceeded(op.config.MaxResultCount)
		return cmds, err
	}
	return
}
