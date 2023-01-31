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

package gate

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
	// the command name for gating the stages of bootstrapping on other services for security
	CommandName string = "gate"
)

type cmd struct {
	cntx          context.Context
	waitGroup     *sync.WaitGroup
	loggingClient logger.LoggingClient
	config        *config.ConfigurationStruct
}

// NewCommand creates a new cmd and parses through options if any
func NewCommand(
	ctx context.Context,
	wg *sync.WaitGroup,
	lc logger.LoggingClient,
	conf *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	cmd := cmd{
		loggingClient: lc,
		config:        conf,
		cntx:          ctx,
		waitGroup:     wg,
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "configDir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	return &cmd, nil
}

// Execute implements Command and runs this command
// command gate gates stages for bootstrapping EdgeX services for security
func (c *cmd) Execute() (statusCode int, err error) {
	c.loggingClient.Infof("Security bootstrapper running %s", CommandName)

	bootstrapServer := tcp.NewTcpServer()
	c.loggingClient.Debugf("init phase: attempts to start up the listener on bootstrap host: %s, port: %d",
		c.config.StageGate.BootStrapper.Host, c.config.StageGate.BootStrapper.StartPort)

	// in a separate go-routine so it won't block the main thread execution
	go openGatingSemaphorePort(bootstrapServer, c.config.StageGate.BootStrapper.StartPort, c.loggingClient,
		"Raising bootstrap semaphore for secure bootstrapping")

	// wait on for others to be done: each of tcp dialers is a blocking call
	c.loggingClient.Debug("Waiting on dependent semaphores required to raise the ready-to-run semaphore ...")
	if err := tcp.DialTcp(
		c.config.StageGate.Registry.Host,
		c.config.StageGate.Registry.ReadyPort,
		c.loggingClient); err != nil {
		retErr := fmt.Errorf("found error while waiting for readiness of Registry at %s:%d, err: %v",
			c.config.StageGate.Registry.Host, c.config.StageGate.Registry.ReadyPort, err)
		return interfaces.StatusCodeExitWithError, retErr
	}
	c.loggingClient.Info("Registry is ready")

	if err := tcp.DialTcp(
		c.config.StageGate.Database.Host,
		c.config.StageGate.Database.ReadyPort,
		c.loggingClient); err != nil {
		retErr := fmt.Errorf("found error while waiting for readiness of Database at %s:%d, err: %v",
			c.config.StageGate.Database.Host, c.config.StageGate.Database.ReadyPort, err)
		return interfaces.StatusCodeExitWithError, retErr
	}
	c.loggingClient.Info("Database is ready")

	// Reached ready-to-run phase
	c.loggingClient.Debugf("ready-to-run phase: attempts to start up the listener on ready-to-run port: %d",
		c.config.StageGate.Ready.ToRunPort)

	readyToRunServer := tcp.NewTcpServer()

	go openGatingSemaphorePort(readyToRunServer, c.config.StageGate.Ready.ToRunPort, c.loggingClient,
		"Raising ready-to-run semaphore for secure bootstrapping")

	// keep running until ctx done
	c.waitGroup.Add(1)
	go func() {
		defer c.waitGroup.Done()

		<-c.cntx.Done()
		c.loggingClient.Info("security bootstrapper finished")
	}()

	return
}

// GetCommandName returns the name of this command
func (c *cmd) GetCommandName() string {
	return CommandName
}

func openGatingSemaphorePort(tcpServer *tcp.TcpServer, portNum int, lc logger.LoggingClient, raisingMsg string) {
	lc.Info(raisingMsg)
	if err := tcpServer.StartListener(portNum, lc, ""); err != nil {
		// listener is blocking forever until some internal critical error happens
		lc.Error(err.Error())
		os.Exit(1)
	}
}
