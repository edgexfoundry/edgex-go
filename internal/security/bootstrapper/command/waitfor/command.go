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

package waitfor

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	CommandName string = "waitFor"
)

type cmd struct {
	waitGroup     *sync.WaitGroup
	loggingClient logger.LoggingClient
	configuration *config.ConfigurationStruct

	// options
	uris          uriFlagsVar
	timeout       time.Duration
	retryInterval time.Duration

	// internal states
	parsedURIs []url.URL
}

// NewCommand creates a new cmd and parses through options if any
func NewCommand(
	_ context.Context,
	_ *sync.WaitGroup,
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	// input sanity checks
	defaultTimeout, err := time.ParseDuration(configuration.StageGate.WaitFor.Timeout)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse duration for StageGate.WaitFor.Timeout: %s: %w",
			configuration.StageGate.WaitFor.Timeout, err)
	} else if defaultTimeout <= 0 {
		return nil, fmt.Errorf("Expect positive time duration (> 0) for StageGate.WaitFor.Timeout: %s",
			configuration.StageGate.WaitFor.Timeout)
	}

	defaultRetryInterval, err := time.ParseDuration(configuration.StageGate.WaitFor.RetryInterval)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse duration for StageGate.WaitFor.RetryInterval: %s: %w",
			configuration.StageGate.WaitFor.RetryInterval, err)
	} else if defaultRetryInterval <= 0 {
		return nil, fmt.Errorf("Expect positive time duration (> 0) for StageGate.WaitFor.RetryInterval: %s",
			configuration.StageGate.WaitFor.RetryInterval)
	}

	cmd := cmd{
		waitGroup:     &sync.WaitGroup{},
		loggingClient: lc,
		configuration: configuration,
	}

	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	flagSet.Var(&cmd.uris, "uri", "Service (tcp/tcp4/tcp6/http/https/unix/file) to wait for before this one starts. "+
		"Can be passed multiple times. e.g. tcp://db:5432")

	flagSet.DurationVar(&cmd.timeout, "timeout", defaultTimeout, "Timeout duration of waiting for services")

	flagSet.DurationVar(&cmd.retryInterval, "retryInterval", defaultRetryInterval, "Duration to pause before retrying")

	err = flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}

	if len(cmd.uris) == 0 {
		return nil, fmt.Errorf("%s %s: argument --uri is required", os.Args[0], CommandName)
	}

	return &cmd, nil
}

// GetCommandName returns the name of this command
func (c *cmd) GetCommandName() string {
	return CommandName
}

// Execute implements Command and runs this command
// command waitFor waits services to be connected and responded
func (c *cmd) Execute() (int, error) {
	c.loggingClient.Infof("Security bootstrapper running %s", CommandName)

	for _, rawURI := range c.uris {
		uri, err := url.Parse(rawURI)
		if err != nil {
			return interfaces.StatusCodeExitWithError, err
		}

		c.parsedURIs = append(c.parsedURIs, *uri)
	}

	if err := c.waitForDependencies(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	return interfaces.StatusCodeExitNormal, nil
}

// waitForDependencies implements very similar ideas from dockerize
func (c *cmd) waitForDependencies() error {
	dependencyChan := make(chan struct{})
	waitErr := make(chan error, 1)

	go func() {
		for _, uri := range c.parsedURIs {
			c.loggingClient.Infof("Waiting for: [%s] with timeout: [%s]", uri.String(), c.timeout.String())

			switch uri.Scheme {
			case "file":
				c.waitForFile(uri)
			case "tcp", "tcp4", "tcp6":
				c.waitForSocket(uri.Scheme, uri.Host)
			case "unix":
				c.waitForSocket(uri.Scheme, uri.Path)
			case "http", "https":
				c.waitForHTTP(uri)
			default:
				waitErr <- fmt.Errorf("invalid host protocol provided: %s. supported protocols are: file, tcp, tcp4, "+
					"tcp6, unix and http", uri.Scheme)
				return
			}
		}

		c.waitGroup.Wait()
		close(dependencyChan)
	}()

	select {
	case err := <-waitErr:
		return err
	case <-dependencyChan:
		break
	case <-time.After(c.timeout):
		return fmt.Errorf("Timeout after %s waiting on dependencies to become available: %v", c.timeout, c.uris)
	}

	return nil
}

func (c *cmd) waitForFile(fileUri url.URL) {
	c.waitGroup.Add(1)
	go func(uri url.URL) {
		defer c.waitGroup.Done()
		ticker := time.NewTicker(c.retryInterval)
		defer ticker.Stop()
		var err error
		for range ticker.C {
			if _, err = os.Stat(uri.Path); err == nil {
				c.loggingClient.Infof("File %s had been generated", uri.String())
				return
			} else if os.IsNotExist(err) {
				// file not exists at this moment
				c.loggingClient.Infof("File %s not exists: %v. Sleeping %s",
					uri.String(), err, c.retryInterval)
			} else {
				// file may or may not exist: see err for details
				c.loggingClient.Infof("Problem with check file %s exist: %v. Sleeping %s",
					uri.String(), err, c.retryInterval)
			}
		}
	}(fileUri)
}

func (c *cmd) waitForSocket(scheme, addr string) {
	c.waitGroup.Add(1)
	go func(schm, adr string) {
		defer c.waitGroup.Done()
		for {
			conn, err := net.DialTimeout(schm, adr, c.timeout)
			if err != nil {
				c.loggingClient.Infof("Problem with dial: %v. Sleeping %s", err.Error(), c.retryInterval)
				time.Sleep(c.retryInterval)
			}

			if conn != nil {
				c.loggingClient.Infof("Connected to %s://%s", schm, adr)
				return
			}
		}
	}(scheme, addr)
}

func (c *cmd) waitForHTTP(uri url.URL) {
	c.waitGroup.Add(1)
	go func(httpURL url.URL) {
		client := &http.Client{
			Timeout: c.timeout,
		}

		defer c.waitGroup.Done()
		for {
			req, err := http.NewRequest(http.MethodGet, httpURL.String(), nil)
			if err != nil {
				c.loggingClient.Infof("Problem with dial: %v. Sleeping %s", err.Error(), c.retryInterval)
				time.Sleep(c.retryInterval)
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				c.loggingClient.Infof("Problem with request: %s. Sleeping %s", err.Error(), c.retryInterval)
				time.Sleep(c.retryInterval)
			} else if err == nil && resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
				// dependency is treated as ok if http status code is between 200 and 300
				c.loggingClient.Infof("Received %d from %s", resp.StatusCode, httpURL.String())
				return
			} else {
				c.loggingClient.Infof("Received %d from %s. Sleeping %s", resp.StatusCode,
					httpURL.String(), c.retryInterval)
				time.Sleep(c.retryInterval)
			}
		}
	}(uri)
}
