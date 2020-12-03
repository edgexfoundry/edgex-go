//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package adduser

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
	CommandName    string = "adduser"
	urlEncodedForm string = "application/x-www-form-urlencoded"
)

type cmd struct {
	loggingClient logger.LoggingClient
	client        internal.HttpCaller
	configuration *config.ConfigurationStruct
	tokenType     string
	username      string
	group         string

	/* jwt vars */
	algorithm     string
	publicKeyPath string
	jwtID         string

	/* oauth2 vars */
	clientID     string
	clientSecret string
	redirectUris string
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
	flagSet.StringVar(&cmd.tokenType, "token-type", "", "Type of token to create: jwt or oauth2")
	flagSet.StringVar(&cmd.username, "user", "", "Username of the user to add")
	flagSet.StringVar(&cmd.group, "group", "admin", "Group to which the user belongs, defaults to 'admin'")

	flagSet.StringVar(&cmd.algorithm, "algorithm", "", "Algorithm used for signing the JWT, RS256 or ES256")
	flagSet.StringVar(&cmd.publicKeyPath, "public_key", "", "Public key (in PEM format) used to validate the JWT.")
	flagSet.StringVar(&cmd.jwtID, "id", "", "ID to use for linkage with JWT claim (usually the 'iss' field)")

	flagSet.StringVar(&cmd.clientID, "client_id", "", "Optional manually-specified OAuth2 client_id.")
	flagSet.StringVar(&cmd.clientSecret, "client_secret", "", "Optional manually-specified OAuth2 client_secret.")
	flagSet.StringVar(&cmd.redirectUris, "redirect_uris", "https://localhost", "OAuth2 redirect URL for browser-based users (default: https://localhost)")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse command: %s: %w", strings.Join(args, " "), err)
	}
	if cmd.tokenType == "" {
		return nil, fmt.Errorf("%s proxy adduser: argument --token-type is required", os.Args[0])
	}
	if cmd.tokenType != interfaces.JwtTokenType && cmd.tokenType != interfaces.OAuth2TokenType {
		return nil, fmt.Errorf("%s proxy adduser: argument --token-type must be either 'jwt' or 'oauth2'", os.Args[0])
	}
	if cmd.username == "" {
		return nil, fmt.Errorf("%s proxy adduser: argument --user is required", os.Args[0])
	}
	if cmd.tokenType == interfaces.JwtTokenType && cmd.algorithm == "" {
		return nil, fmt.Errorf("%s proxy adduser: argument --algorithm is required", os.Args[0])
	}
	if cmd.tokenType == interfaces.JwtTokenType && cmd.algorithm != "RS256" && cmd.algorithm != "ES256" {
		return nil, fmt.Errorf("%s proxy adduser: argument --algorithm must be either 'RS256' or 'ES256'", os.Args[0])
	}
	if cmd.tokenType == interfaces.JwtTokenType && cmd.publicKeyPath == "" {
		return nil, fmt.Errorf("%s proxy adduser: argument --public_key is required", os.Args[0])
	}
	if cmd.tokenType == interfaces.OAuth2TokenType && cmd.redirectUris == "" {
		return nil, fmt.Errorf("%s proxy adduser: argument --redirect_uris is required", os.Args[0])
	}

	return &cmd, err
}

func (c *cmd) Execute() (statusCode int, err error) {
	switch c.tokenType {
	case interfaces.JwtTokenType:
		statusCode, err = c.ExecuteAddJwt()
	case interfaces.OAuth2TokenType:
		statusCode, err = c.ExecuteAddOAuth2()
	default:
		statusCode = interfaces.StatusCodeExitWithError
		err = fmt.Errorf("unsupported token type %s", c.tokenType)
	}

	return
}

func (c *cmd) createConsumer() error {
	// Create kong consumer with the specified username
	// https://docs.konghq.com/hub/kong-inc/jwt/#create-a-consumer

	form := url.Values{
		"username": []string{c.username},
	}
	kongURL := strings.Join([]string{c.configuration.KongURL.GetProxyBaseURL(), "consumers"}, "/")
	c.loggingClient.Info(fmt.Sprintf("creating consumer (user) on the endpoint of %s", kongURL))

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return fmt.Errorf("Failed to prepare new consumer request %s: %w", c.username, err)
	}
	req.Header.Add(clients.ContentType, urlEncodedForm)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send new consumer request %s: %w", c.username, err)
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
		c.loggingClient.Info(fmt.Sprintf("created consumer (user) '%s'", c.username))
	case http.StatusConflict:
		c.loggingClient.Info(fmt.Sprintf("consumer '%s' already created", c.username))
	default:
		c.loggingClient.Error(fmt.Sprintf("%s", responseBody))
		return fmt.Errorf("Create consumer request failed with code: %d", resp.StatusCode)
	}

	return nil
}

func (c *cmd) addUserToGroup() error {
	// Associate the consumer with a group
	// https://docs.konghq.com/hub/kong-inc/acl/#associating-consumers

	form := url.Values{
		"group": []string{c.group},
	}
	kongURL := strings.Join([]string{c.configuration.KongURL.GetProxyBaseURL(), "consumers", c.username, "acls"}, "/")
	c.loggingClient.Info(fmt.Sprintf("Associating consumer to acl using endpoint %s", kongURL))

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return fmt.Errorf("Failed to build request to associate consumer %s to group %s: %w", c.username, c.group, err)
	}
	req.Header.Add(clients.ContentType, urlEncodedForm)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to submit request to associate consumer %s to group %s: %w", c.username, c.group, err)
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
		c.loggingClient.Info(fmt.Sprintf("associated consumer %s to group %s", c.username, c.group))
	case http.StatusConflict:
		c.loggingClient.Info(fmt.Sprintf("consumer %s already associated to group %s", c.username, c.group))
	default:
		c.loggingClient.Error(fmt.Sprintf("%s", responseBody))
		return fmt.Errorf("Failed to associate consumer to group with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *cmd) ExecuteAddJwt() (int, error) {
	publicKey, err := ioutil.ReadFile(c.publicKeyPath)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to read public key from file %s: %w", c.publicKeyPath, err)
	}

	if err := c.createConsumer(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}
	if err := c.addUserToGroup(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	// Associate a JWT credential with the consumer
	// https://docs.konghq.com/hub/kong-inc/jwt/#create-a-jwt-credential

	form := url.Values{
		"algorithm":      []string{c.algorithm},
		"rsa_public_key": []string{string(publicKey)},
		"secret":         []string{"required-but-not-used-see-documentation"},
	}

	if len(c.jwtID) > 0 {
		// Kong creates random key if one is not supplied.
		form.Set("key", c.jwtID)
	}

	kongURL := strings.Join([]string{c.configuration.KongURL.GetProxyBaseURL(), "consumers", c.username, "jwt"}, "/")
	c.loggingClient.Info(fmt.Sprintf("associating JWT on the endpoint of %s", kongURL))

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to prepare request to associate JWT to user %s: %w", c.username, err)
	}
	req.Header.Add(clients.ContentType, urlEncodedForm)
	resp, err := c.client.Do(req)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to send request to associate JWT to user %s: %w", c.username, err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	var parsedResponse map[string]interface{}

	switch resp.StatusCode {
	case http.StatusCreated:
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&parsedResponse); err != nil {
			return interfaces.StatusCodeExitWithError, fmt.Errorf("Unable to parse associate JWT response: %w", err)
		}
	case http.StatusConflict:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Associate JWT request failed (likely due to duplicate ID) with code: %d", resp.StatusCode)
	default:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Associate JWT request failed with code: %d", resp.StatusCode)
	}

	outputKey := fmt.Sprintf("%s", parsedResponse["key"])

	fmt.Printf("%s\n", outputKey)

	return interfaces.StatusCodeExitNormal, nil
}

func (c *cmd) ExecuteAddOAuth2() (statusCode int, err error) {
	if err := c.createConsumer(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}
	if err := c.addUserToGroup(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	// Create an OAuth application
	// https://docs.konghq.com/hub/kong-inc/oauth2/#create-an-application

	form := url.Values{
		"name":          []string{c.username}, // use username as application name
		"redirect_uris": []string{c.redirectUris},
	}

	// Client ID and client secret are auto-generated if not supplied
	if len(c.clientID) > 0 {
		form.Set("client_id", c.clientID)
	}
	if len(c.clientID) > 0 {
		form.Set("client_secret", c.clientSecret)
	}

	kongURL := strings.Join([]string{c.configuration.KongURL.GetProxyBaseURL(), "consumers", c.username, "oauth2"}, "/")
	c.loggingClient.Info(fmt.Sprintf("creating oauth application at %s", kongURL))

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to prepare request to create oauth application %s: %w", c.username, err)
	}
	req.Header.Add(clients.ContentType, urlEncodedForm)
	resp, err := c.client.Do(req)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Failed to send request to create oauth application %s: %w", c.username, err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	var parsedResponse map[string]interface{}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&parsedResponse); err != nil {
			return interfaces.StatusCodeExitWithError, fmt.Errorf("Unable to parse associate JWT response: %w", err)
		}
	case http.StatusConflict:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Associate JWT request failed (likely due to duplicate ID) with code: %d", resp.StatusCode)
	default:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Associate JWT request failed with code: %d", resp.StatusCode)
	}

	clientID := fmt.Sprintf("%s", parsedResponse["client_id"])
	clientSecret := fmt.Sprintf("%s", parsedResponse["client_secret"])

	err = json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
		"client_id":     clientID,
		"client_secret": clientSecret,
	})
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("Marshaling of client id and secret failed")
	}

	return interfaces.StatusCodeExitNormal, nil
}
