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

package genpassword

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	CommandName string = "genPassword"

	randomBytesLength = 33 // 264 bits of entropy
)

type cmd struct {
	loggingClient logger.LoggingClient
	config        *config.ConfigurationStruct
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
	flagSet.StringVar(&dummy, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	return &cmd, nil
}

// Execute implements Command and runs this command
// command genPassword generates a random password
func (c *cmd) Execute() (int, error) {
	c.loggingClient.Infof("Security bootstrapper running %s", CommandName)

	randomBytes := make([]byte, randomBytesLength)
	_, err := rand.Read(randomBytes) // all of salt guaranteed to be filled if err==nil
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	randPass := base64.StdEncoding.EncodeToString(randomBytes)
	// output the randPass to stdout
	fmt.Fprintln(os.Stdout, randPass)

	return interfaces.StatusCodeExitNormal, nil
}

// GetCommandName returns the name of this command
func (c *cmd) GetCommandName() string {
	return CommandName
}
