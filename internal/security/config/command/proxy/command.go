//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package proxy

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/security/config/command/proxy/adduser"
	"github.com/edgexfoundry/edgex-go/internal/security/config/command/proxy/deluser"
	"github.com/edgexfoundry/edgex-go/internal/security/config/command/proxy/tls"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	CommandName = "proxy"
)

func NewCommand(
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	var command interfaces.Command
	var err error

	if len(args) < 1 {
		return nil, fmt.Errorf("subcommand required (adduser, deluser, tls)")
	}

	commandName := args[0]

	switch commandName {
	case tls.CommandName:
		command, err = tls.NewCommand(lc, args[1:])
	case adduser.CommandName:
		command, err = adduser.NewCommand(lc, configuration, args[1:])
	case deluser.CommandName:
		command, err = deluser.NewCommand(lc, configuration, args[1:])
	default:
		command = nil
		err = fmt.Errorf("unsupported command %s", commandName)
	}

	return command, err
}
