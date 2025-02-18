//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package createtoken

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/container"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/tokenprovider"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v4/secrets"
)

const (
	CommandName  string = "createToken"
	entityIdFlag string = "entityId"
)

type command struct {
	waitGroup *sync.WaitGroup
	dic       *di.Container

	// options
	entityId string
}

// NewCommand creates a new cmd and parses through options if any
func NewCommand(
	dic *di.Container,
	args []string) (interfaces.Command, error) {
	cmd := command{
		waitGroup: &sync.WaitGroup{},
		dic:       dic,
	}
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	var entityId string

	lc.Debugf("Initialize '%s' command with args: %v", CommandName, args)

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&entityId, entityIdFlag, "", "")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("unable to parse command '%s': %w", strings.Join(args, " "), err)
	}

	if len(entityId) == 0 {
		return nil, fmt.Errorf("%s %s: argument --entityId is required", os.Args[0], CommandName)
	}
	cmd.entityId = entityId
	return &cmd, nil
}

// GetCommandName returns the name of this command
func (c *command) GetCommandName() string {
	return CommandName
}

// Execute implements Command and runs this command
// command createtoken regenerate a token file for the specified entityId argument
func (c *command) Execute() (int, error) {
	lc := bootstrapContainer.LoggingClientFrom(c.dic.Get)
	lc.Infof("Security file token provider - executing %s command .......", CommandName)

	if err := regenToken(c.entityId, c.dic); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	return interfaces.StatusCodeExitNormal, nil
}

// regenToken invokes RegenToken from fileProvider to recreate the token file for the specified entity id
func regenToken(entityId string, dic *di.Container) errors.EdgeX {
	cfg := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	var requester internal.HttpCaller

	clientConfig := types.SecretConfig{
		Type:     secrets.DefaultSecretStore,
		Host:     cfg.SecretStore.Host,
		Port:     cfg.SecretStore.Port,
		Protocol: cfg.SecretStore.Protocol,
	}
	client, err := secrets.NewSecretStoreClient(clientConfig, lc, requester)
	if err != nil {
		lc.Errorf("error occurred creating SecretStoreClient: %v", err)
		return errors.NewCommonEdgeXWrapper(err)
	}

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenProvider := authtokenloader.NewAuthTokenLoader(fileOpener)

	fileProvider := tokenprovider.NewTokenProvider(lc, fileOpener, tokenProvider, client)
	fileProvider.SetConfiguration(cfg.SecretStore, cfg.TokenFileProvider)

	err = fileProvider.RegenToken(entityId)
	if err != nil {
		lc.Errorf("error occurred while re-generating token: %v", err)
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}
