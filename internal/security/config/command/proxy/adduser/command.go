//
// Copyright (c) 2020-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package adduser

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/config/command/proxy/shared"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	secretStoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	CommandName        string = "adduser"
	DefaultTokenTTL    string = "1h"
	DefaultJWTTTL      string = "1h"
	DefaultJWTAudience string = "edgex"
)

type cmd struct {
	loggingClient   logger.LoggingClient
	configuration   *secretStoreConfig.ConfigurationStruct
	proxyUserCommon shared.ProxyUserCommon
	useRootToken    bool
	username        string
	tokenTTL        string
	jwtAudience     string
	jwtTTL          string
}

// NewCommand add a new user to Vault, which can then authenticate through the gateway
func NewCommand(
	lc logger.LoggingClient,
	configuration *secretStoreConfig.ConfigurationStruct,
	args []string) (*cmd, error) {

	cmd := cmd{
		loggingClient: lc,
		configuration: configuration,
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "configDir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors
	flagSet.StringVar(&cmd.username, "user", "", "Username of the user to add")
	flagSet.StringVar(&cmd.tokenTTL, "tokenTTL", DefaultTokenTTL, "Vault token created as a result of vault login lasts this long  (_s, _m, _h, or _d, seconds if no unit)")
	flagSet.StringVar(&cmd.jwtAudience, "jwtAudience", DefaultJWTAudience, "Optionally change JWT audience of generated JWT's")
	flagSet.StringVar(&cmd.jwtTTL, "jwtTTL", DefaultJWTTTL, "JWT created by vault identity provider lasts this long (_s, _m, _h, or _d, seconds if no unit)")
	flagSet.BoolVar(&cmd.useRootToken, "useRootToken", false, "Set to true to TokenFile in config points to a resp-init.json instead of a service token")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, err
	}
	if cmd.username == "" {
		return nil, fmt.Errorf("%s vault adduser: argument --user is required", os.Args[0])
	}

	cmd.proxyUserCommon, err = shared.NewProxyUserCommon(lc, configuration)
	if err != nil {
		lc.Errorf("failed to initialize secret store client: %s", err.Error())
		return nil, err
	}

	return &cmd, err
}

// Execute runs the command for adding a user
func (c *cmd) Execute() (statusCode int, err error) {

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

	credentials, err := c.proxyUserCommon.DoAddUser(privilegedToken, c.username, c.tokenTTL, c.jwtAudience, c.jwtTTL)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	// Output credentials for new account

	err = json.NewEncoder(os.Stdout).Encode(credentials)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	return interfaces.StatusCodeExitNormal, nil

}
