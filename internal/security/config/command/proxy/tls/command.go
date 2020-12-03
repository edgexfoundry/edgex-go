//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package tls

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const (
	CommandName = "tls"
)

type cmd struct {
	loggingClient   logger.LoggingClient
	configuration   *config.ConfigurationStruct
	certificatePath string
	privateKeyPath  string
}

func NewCommand(
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	cmd := cmd{
		loggingClient: lc,
		configuration: configuration,
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	flagSet.StringVar(&cmd.certificatePath, "incert", "", "Path to PEM-encoded leaf certificate")
	flagSet.StringVar(&cmd.privateKeyPath, "inkey", "", "Path to PEM-encoded private key")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}
	if cmd.certificatePath == "" {
		return nil, fmt.Errorf("%s proxy tls: argument --incert is required", os.Args[0])
	}
	if cmd.privateKeyPath == "" {
		return nil, fmt.Errorf("%s proxy tls: argument --inkey is required", os.Args[0])
	}

	return &cmd, nil
}

func (c *cmd) Execute() (statusCode int, err error) {
	fmt.Println("TODO: Configure inbound TLS certificate.")
	fmt.Printf("--incert %s\n", c.certificatePath)
	fmt.Printf("--inkey %s\n", c.privateKeyPath)
	err = fmt.Errorf("tls command is unimplemented")
	return interfaces.StatusCodeExitWithError, err
}
