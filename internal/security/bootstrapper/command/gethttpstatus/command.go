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

package gethttpstatus

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	CommandName string = "getHttpStatus"
)

type cmd struct {
	loggingClient logger.LoggingClient
	client        internal.HttpCaller
	configuration *config.ConfigurationStruct

	// options
	httpURL string
}

// NewCommand creates a new cmd and parses through options if any
func NewCommand(
	_ context.Context,
	_ *sync.WaitGroup,
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	cmd := cmd{
		loggingClient: lc,
		client:        pkg.NewRequester(lc).Insecure(),
		configuration: configuration,
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors
	flagSet.StringVar(&cmd.httpURL, "url", "", "get the status code returning from the input http URL address")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	if len(cmd.httpURL) == 0 {
		return nil, fmt.Errorf("%s %s: argument --url is required", os.Args[0], CommandName)
	}

	return &cmd, nil
}

// GetCommandName returns the name of this command
func (c *cmd) GetCommandName() string {
	return CommandName
}

// Execute implements Command and runs this command
// command getHttpStatus makes a http GET request and outputs the http status code
func (c *cmd) Execute() (int, error) {
	c.loggingClient.Infof("Security bootstrapper running %s", CommandName)

	_, err := url.Parse(c.httpURL)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	c.loggingClient.Infof("http calls on the endpoint of %s", c.httpURL)
	req, err := http.NewRequest(http.MethodGet, c.httpURL, http.NoBody)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to prepare request for http URL: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to send request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// don't care about the response body, only the status code
	fmt.Fprintln(os.Stdout, resp.StatusCode)

	return interfaces.StatusCodeExitNormal, nil
}
