/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *
 *******************************************************************************/

package listen

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/tcp"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	CommandName string = "listenTcp"
)

type cmd struct {
	loggingClient logger.LoggingClient
	config        *config.ConfigurationStruct

	// options:
	tcpHost string
	tcpPort int
}

// NewCommand creates a new cmd and parses through options if any
func NewCommand(
	_ context.Context,
	_ *sync.WaitGroup,
	lc logger.LoggingClient,
	conf *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	cmd := cmd{
		loggingClient: lc,
		config:        conf,
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "configDir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors
	flagSet.StringVar(&cmd.tcpHost, "host", "", "the hostname of TCP server to listen ")

	flagSet.IntVar(&cmd.tcpPort, "port", 0, "the port number of TCP server to listen ")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	if cmd.tcpPort == 0 {
		return nil, fmt.Errorf("%s %s: argument --port is required", os.Args[0], CommandName)
	}

	return &cmd, nil
}

// Execute implements Command and runs this command
// command listenTcp starts a TCP listener with configured port and host
func (c *cmd) Execute() (int, error) {
	c.loggingClient.Infof("Security bootstrapper running %s", CommandName)

	tcpServer := tcp.NewTcpServer()

	// block and listening forever until internal error
	if err := tcpServer.StartListener(c.tcpPort, c.loggingClient, c.tcpHost); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	return interfaces.StatusCodeExitNormal, nil
}

// GetCommandName returns the name of this command
func (c *cmd) GetCommandName() string {
	return CommandName
}
