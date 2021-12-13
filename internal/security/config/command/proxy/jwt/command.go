//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package jwt

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	jwt "github.com/golang-jwt/jwt/v4"
)

const (
	CommandName string = "jwt"
)

type cmd struct {
	loggingClient  logger.LoggingClient
	configuration  *config.ConfigurationStruct
	algorithm      string
	privateKeyPath string
	jwtID          string
	expiration     string
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

	flagSet.StringVar(&cmd.algorithm, "algorithm", "", "Algorithm used for signing the JWT, RS256 or ES256")
	flagSet.StringVar(&cmd.privateKeyPath, "private_key", "", "Private key used to sign the JWT (PEM-encoded)")
	flagSet.StringVar(&cmd.jwtID, "id", "", "The 'key' field (ID) from the 'adduser' command")
	flagSet.StringVar(&cmd.expiration, "expiration", "1h", "Duration of generated jwt expressed as a golang-parseable duration value (default: 1h)")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}
	if cmd.algorithm != "RS256" && cmd.algorithm != "ES256" {
		return nil, fmt.Errorf("%s proxy jwt: argument --algorithm must be either 'RS256' or 'ES256'", os.Args[0])
	}
	if cmd.privateKeyPath == "" {
		return nil, fmt.Errorf("%s proxy jwt: argument --private_key is required", os.Args[0])
	}
	if cmd.jwtID == "" {
		return nil, fmt.Errorf("%s proxy jwt: argument --id is required", os.Args[0])
	}

	return &cmd, nil
}

func (c *cmd) Execute() (int, error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		Issuer:    c.jwtID,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}
	if len(c.expiration) > 0 {
		duration, err := time.ParseDuration(c.expiration)
		if err != nil {
			return interfaces.StatusCodeExitWithError, fmt.Errorf("Could not parse JWT duration: %w", err)
		}
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(duration))
	}

	bytes, err := os.ReadFile(c.privateKeyPath)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Could not read private key from file %s: %w", c.privateKeyPath, err)
	}

	var signingMethod jwt.SigningMethod
	var key interface{}

	switch c.algorithm {
	case interfaces.RS256:
		signingMethod = jwt.SigningMethodRS256
		key, err = jwt.ParseRSAPrivateKeyFromPEM(bytes)
	case interfaces.ES256:
		signingMethod = jwt.SigningMethodES256
		eckey, err := jwt.ParseECPrivateKeyFromPEM(bytes)
		if err == nil && eckey.Params().BitSize != 256 {
			return interfaces.StatusCodeExitWithError, fmt.Errorf("EC key bit size is incorrect (%d instead of 256)", eckey.Params().BitSize)
		}
		key = eckey
	default:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Unrecognized signing algorith %s", c.algorithm)
	}
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Could not parse %s private key from %s: %w", c.algorithm, c.privateKeyPath, err)
	}

	token := jwt.NewWithClaims(signingMethod, claims)

	signedToken, err := token.SignedString(key)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Could not sign JWT: %w", err)
	}

	fmt.Printf("%s\n", signedToken)

	return interfaces.StatusCodeExitNormal, nil
}
