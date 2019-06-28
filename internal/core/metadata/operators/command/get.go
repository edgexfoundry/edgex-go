package command

import (
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
