//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package oauth2

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	proxyCommon "github.com/edgexfoundry/edgex-go/internal/security/config/command/proxy/common"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

const (
	CommandName string = "oauth2"
)

type cmd struct {
	loggingClient logger.LoggingClient
	client        internal.HttpCaller
	configuration *config.ConfigurationStruct
	clientID      string
	clientSecret  string
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

	flagSet.StringVar(&cmd.jwt, "jwt", "", "The JWT for use when accessing the Kong Admin API")
	flagSet.StringVar(&cmd.clientID, "client_id", "", "Optional manually-specified OAuth2 client_id. Will be generated if not present. Equivalent to a username.")
	flagSet.StringVar(&cmd.clientSecret, "client_secret", "", "Optional manually-specified OAuth2 client_secret. Will be generated if not present. Equivalent to a password.")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("unable to parse command: %s: %w", strings.Join(args, " "), err)
	}
	if cmd.clientID == "" {
		return nil, fmt.Errorf("%s proxy oauth2: argument --client_id is required", os.Args[0])
	}
	if cmd.clientSecret == "" {
		return nil, fmt.Errorf("%s proxy oauth2: argument --client_secret is required", os.Args[0])
	}
	if cmd.jwt == "" {
		return nil, fmt.Errorf("%s proxy oauth2: argument --jwt is required", os.Args[0])
	}

	return &cmd, nil
}

func (c *cmd) Execute() (statusCode int, err error) {
	// Client credentials grant: https://tools.ietf.org/html/rfc6749#section-4.4
	// curl -k https://kong:8443/{service}/oauth2/token -d "grant_type=client_credentials" -d "scope=" -d "client_id=<clientid>" -d "client_secret=<clientsecret>"

	// Setup the URL to send a request to
	kongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(), c.configuration.KongAuth.Resource, "oauth2/token"}, "/")
	c.loggingClient.Infof("creating token on the endpoint of %s", kongURL)

	// Setup the request
	req, err := http.NewRequest(
		http.MethodPost,
		kongURL,
		strings.NewReader(
			url.Values{
				"client_id":     []string{c.clientID},
				"client_secret": []string{c.clientSecret},
				"grant_type":    []string{"client_credentials"},
				"scope":         []string{""},
			}.Encode(),
		),
	)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to create request to get OAuth2 token: %w", err)
	}

	// Add header values
	req.Header.Add(common.ContentType, proxyCommon.UrlEncodedForm)
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.jwt)

	// Execute the request
	resp, err := c.client.Do(req)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to obtain OAuth2 token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Get the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("token request failed with code: %d", resp.StatusCode)
	}

	// Parse the response to snag the access_token
	var parsedResponse map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&parsedResponse); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("unable to parse create token response: %w", err)
	}

	// Is there an access token?
	if parsedResponse["access_token"] == nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("access token not found in returned response")
	}

	// Print the access token to STDOUT
	fmt.Printf("%s\n", parsedResponse["access_token"])

	return interfaces.StatusCodeExitNormal, nil
}
