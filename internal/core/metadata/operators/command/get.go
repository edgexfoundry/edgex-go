/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package command

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

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
	config   bootstrapConfig.ServiceInfo
	database CommandLoader
}

func NewCommandLoadAll(cfg bootstrapConfig.ServiceInfo, db CommandLoader) CommandAllExecutor {
	return commandLoadAll{config: cfg, database: db}
}

func (op commandLoadAll) Execute() (cmds []contract.Command, err error) {
	cmds, err = op.database.GetAllCommands()
	if err != nil {
		return
	}

	if len(cmds) > op.config.MaxResultCount {
		err = errors.NewErrLimitExceeded(op.config.MaxResultCount)
		return []contract.Command{}, err
	}
	return
}

//Get Command By ID
type CommandByIdExecutor interface {
	Execute() (contract.Command, error)
}

type commandByIdLoader struct {
	database CommandLoader
	cid      string
}

func NewCommandById(db CommandLoader, cid string) CommandByIdExecutor {
	return commandByIdLoader{database: db, cid: cid}
}

func (op commandByIdLoader) Execute() (cmd contract.Command, err error) {
	cmd, err = op.database.GetCommandById(op.cid)
	if err != nil && err == db.ErrNotFound {
		err = errors.NewErrItemNotFound(fmt.Sprintf("command with id %s not found", op.cid))
	}
	return
}

//Get Command By Name
type CommandsByNameExecutor interface {
	Execute() ([]contract.Command, error)
}

type commandsByNameLoader struct {
	database CommandLoader
	cname    string
}

func NewCommandsByName(db CommandLoader, cname string) CommandsByNameExecutor {
	return commandsByNameLoader{database: db, cname: cname}
}

func (op commandsByNameLoader) Execute() (cmd []contract.Command, err error) {
	return op.database.GetCommandsByName(op.cname)
}
