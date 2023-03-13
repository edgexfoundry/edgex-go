//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package deluser

import (
	"flag"
	"fmt"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/config/command/proxy/shared"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	secretStoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	CommandName string = "deluser"
)

type cmd struct {
	loggingClient   logger.LoggingClient
	configuration   *secretStoreConfig.ConfigurationStruct
	proxyUserCommon shared.ProxyUserCommon
	useRootToken    bool
	username        string
}

func NewCommand(
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (*cmd, error) {

	cmd := cmd{
		loggingClient: lc,
		configuration: configuration,
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "configDir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	flagSet.StringVar(&cmd.username, "user", "", "Username of the user to delete")
	flagSet.BoolVar(&cmd.useRootToken, "useRootToken", false, "Set to true to TokenFile in config points to a resp-init.json instead of a service token")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, err
	}
	if cmd.username == "" {
		return nil, fmt.Errorf("%s proxy deluser: argument --user is required", os.Args[0])
	}

	cmd.proxyUserCommon, err = shared.NewProxyUserCommon(lc, configuration)
	if err != nil {
		lc.Errorf("failed to initialize secret store client: %s", err.Error())
		return nil, err
	}

	return &cmd, err
}

// Execute runs the command to delete a user
func (c *cmd) Execute() (int, error) {

	// Get a token to use to make the call to Vault

	tokenLoadMethod := c.proxyUserCommon.LoadServiceToken
	if c.useRootToken {
		tokenLoadMethod = c.proxyUserCommon.LoadRootToken
	}

	privilegedToken, revokeFunc, err := tokenLoadMethod()
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}
	defer revokeFunc()

	// Perform requested action

	err = c.proxyUserCommon.DoDeleteUser(privilegedToken, c.username)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	// Output results

	fmt.Printf("Deleted user %s\n", c.username)

	return interfaces.StatusCodeExitNormal, nil

}
