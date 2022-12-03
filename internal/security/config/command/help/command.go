//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package help

import (
	"flag"
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/security/config/command"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	CommandName = "help"
)

type cmd struct {
	loggingClient logger.LoggingClient
	flagSet       *flag.FlagSet
}

func NewCommand(
	lc logger.LoggingClient,
	args []string) (interfaces.Command, error) {

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	return &cmd{
		loggingClient: lc,
		flagSet:       flagSet,
	}, nil
}

func (c *cmd) Execute() (statusCode int, err error) {
	command.HelpCallback()
	return interfaces.StatusCodeExitNormal, nil
}
