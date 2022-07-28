//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/config/command/help"
	"github.com/edgexfoundry/edgex-go/internal/security/config/command/proxy"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

type Bootstrap struct {
	exitStatusCode int
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by this utility
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	var command interfaces.Command
	var err error

	var confdir string
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagSet.StringVar(&confdir, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors
	err = flagSet.Parse(os.Args[1:])
	if err != nil {
		lc.Error(err.Error())
		return false
	}

	subcommandArgs := []string{}

	commandName := flagSet.Arg(0)
	if flag.NArg() > 0 {
		subcommandArgs = flag.Args()[1:]
	}

	switch commandName {
	case help.CommandName:
		command, err = help.NewCommand(lc, configuration, subcommandArgs)
	case proxy.CommandName:
		command, err = proxy.NewCommand(lc, configuration, subcommandArgs)
	default:
		lc.Error(fmt.Sprintf("unsupported command %s", commandName))
		b.exitStatusCode = interfaces.StatusCodeNoOptionSelected
		return false
	}

	if err != nil {
		lc.Error(err.Error())
		b.exitStatusCode = interfaces.StatusCodeExitWithError
		return false
	}

	exitStatusCode, err := command.Execute()
	if err != nil {
		lc.Error(err.Error())
		return false
	}
	b.exitStatusCode = exitStatusCode
	return true
}

func (b *Bootstrap) ExitStatusCode() int {
	return b.exitStatusCode
}
