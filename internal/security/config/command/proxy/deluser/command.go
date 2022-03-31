//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package deluser

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	CommandName string = "deluser"
)

type cmd struct {
	loggingClient logger.LoggingClient
	client        internal.HttpCaller
	configuration *config.ConfigurationStruct
	username      string
	jwt           string
}

func NewCommand(
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

	flagSet.StringVar(&cmd.username, "user", "", "Username of the user to delete")
	flagSet.StringVar(&cmd.jwt, "jwt", "", "The JWT for use when accessing the Kong Admin API")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}
	if cmd.username == "" {
		return nil, fmt.Errorf("%s proxy deluser: argument --user is required", os.Args[0])
	}
	if cmd.jwt == "" {
		return nil, fmt.Errorf("%s proxy deluser: argument --jwt is required", os.Args[0])
	}

	return &cmd, err
}

func (c *cmd) Execute() (int, error) {
	// Delete Kong consumer
	// https://docs.konghq.com/2.1.x/admin-api/#delete-consumer

	kongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(), "consumers", c.username}, "/")
	c.loggingClient.Infof("deleting consumer (user) on the endpoint of %s", kongURL)

	// Setup request
	req, err := http.NewRequest(http.MethodDelete, kongURL, nil)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to prepare delete consumer request %s: %w", c.username, err)
	}
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.jwt)

	resp, err := c.client.Do(req)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to send delete consumer request %s: %w", c.username, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		c.loggingClient.Infof("deleted consumer (user) '%s'", c.username)
	default:
		responseBody, _ := io.ReadAll(resp.Body)
		c.loggingClient.Error(fmt.Sprintf("Error response: %s", responseBody))
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Delete consumer request failed with code: %d", resp.StatusCode)
	}

	return interfaces.StatusCodeExitNormal, nil
}
