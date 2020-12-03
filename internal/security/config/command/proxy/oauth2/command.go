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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
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
}

func NewCommand(
	lc logger.LoggingClient,
	configuration *config.ConfigurationStruct,
	args []string) (interfaces.Command, error) {

	cmd := cmd{
		loggingClient: lc,
		client:        secretstoreclient.NewRequestor(lc).Insecure(),
		configuration: configuration,
	}
	var dummy string

	flagSet := flag.NewFlagSet(CommandName, flag.ContinueOnError)
	flagSet.StringVar(&dummy, "confdir", "", "") // handled by bootstrap; duplicated here to prevent arg parsing errors

	flagSet.StringVar(&cmd.clientID, "client_id", "", "Optional manually-specified OAuth2 client_id. Will be generated if not present. Equivalent to a username.")
	flagSet.StringVar(&cmd.clientSecret, "client_secret", "", "Optional manually-specified OAuth2 client_secret. Will be generated if not present. Equivalent to a password.")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}
	if cmd.clientID == "" {
		return nil, fmt.Errorf("%s proxy oauth2: argument --client_id is required", os.Args[0])
	}
	if cmd.clientSecret == "" {
		return nil, fmt.Errorf("%s proxy oauth2: argument --client_secret is required", os.Args[0])
	}

	return &cmd, nil
}

func (c *cmd) Execute() (statusCode int, err error) {
	// Client credentialks grant: https://tools.ietf.org/html/rfc6749#section-4.4
	// curk -k https://kong:8443/{service}/oauth2/token -d "grant_type=client_credentials" -d "scope=" -d "client_id=<clientid>" -d "client_secret=<clientsecret>"

	clientCredentialsForm := url.Values{
		"client_id":     []string{c.clientID},
		"client_secret": []string{c.clientSecret},
		"grant_type":    []string{"client_credentials"},
		"scope":         []string{""},
	}
	kongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(), c.configuration.KongAuth.Resource, "oauth2/token"}, "/")
	c.loggingClient.Info(fmt.Sprintf("creating token on the endpoint of %s", kongURL))

	formVal := clientCredentialsForm.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to create request to get OAuth2 token: %w", err)
	}
	req.Header.Add(clients.ContentType, "application/x-www-form-urlencoded")
	resp, err := c.client.Do(req)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to obtain OAuth2 token: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	if resp.StatusCode != http.StatusOK {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Token request failed with code: %d", resp.StatusCode)
	}

	var parsedResponse map[string]interface{}

	if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&parsedResponse); err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Unable to parse create token response: %w", err)
	}

	if parsedResponse["access_token"] == nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Access token not found in returned response")
	}

	fmt.Printf("%s\n", parsedResponse["access_token"])

	return interfaces.StatusCodeExitNormal, nil
}
