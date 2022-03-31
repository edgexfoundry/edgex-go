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
	CommandName string = "adduser"
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
	jwt           string

	/* oauth2 vars */
	clientID     string
	clientSecret string
	redirectUris string
}

// NewCommand will instantiate a command type in order to add a user to the
// currently running Kong gateway.
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
	//flagSet.StringVar(&cmd.tokenType, "token-type", "", "Type of token to create: jwt or oauth2")  	// #3564: Deprecate for Ireland; might bring back in Jakarta
	flagSet.StringVar(&cmd.tokenType, "token-type", "jwt", "Type of token to create: jwt")
	flagSet.StringVar(&cmd.username, "user", "", "Username of the user to add")
	flagSet.StringVar(&cmd.group, "group", "gateway-group", "Group to which the user belongs, defaults to 'gateway-group'")

	flagSet.StringVar(&cmd.algorithm, "algorithm", "", "Algorithm used for signing the JWT, RS256 or ES256")
	flagSet.StringVar(&cmd.publicKeyPath, "public_key", "", "Public key (in PEM format) used to validate the JWT.")
	flagSet.StringVar(&cmd.jwtID, "id", "", "ID to use for linkage with JWT claim (usually the 'iss' field)")
	flagSet.StringVar(&cmd.jwt, "jwt", "", "The JWT for use when accessing the Kong Admin API")

	flagSet.StringVar(&cmd.clientID, "client_id", "", "Optional manually-specified OAuth2 client_id.")
	flagSet.StringVar(&cmd.clientSecret, "client_secret", "", "Optional manually-specified OAuth2 client_secret.")
	flagSet.StringVar(&cmd.redirectUris, "redirect_uris", "https://localhost", "OAuth2 redirect URL for browser-based users (default: https://localhost)")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("unable to parse command: %s: %w", strings.Join(args, " "), err)
	}
	if cmd.tokenType == "" {
		return nil, fmt.Errorf("%s proxy adduser: argument --token-type is required", os.Args[0])
	}
	// #3564: Deprecate for Ireland; commenting in case user community needs back in Jakarta stabilization release
	//if cmd.tokenType != interfaces.JwtTokenType && cmd.tokenType != interfaces.OAuth2TokenType {
	//	return nil, fmt.Errorf("%s proxy adduser: argument --token-type must be either 'jwt' or 'oauth2'", os.Args[0])
	//}
	if cmd.tokenType != interfaces.JwtTokenType {
		return nil, fmt.Errorf("%s proxy adduser: argument --token-type must be 'jwt'", os.Args[0])
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
	if cmd.jwt == "" {
		return nil, fmt.Errorf("%s proxy adduser: argument --jwt is required", os.Args[0])
	}

	return &cmd, err
}

// Execute runs the command for adding a user and branches off by the type
// of token that was selected.
func (c *cmd) Execute() (statusCode int, err error) {
	switch c.tokenType {
	case interfaces.JwtTokenType:
		statusCode, err = c.ExecuteAddJwt()
	// #3564: Deprecate for Ireland; commenting in case user community needs back in Jakarta stabilization release
	//case interfaces.OAuth2TokenType:
	//	statusCode, err = c.ExecuteAddOAuth2()
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
	kongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(), "consumers"}, "/")
	c.loggingClient.Infof("creating consumer (user) on the endpoint of %s", kongURL)

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return fmt.Errorf("Failed to prepare new consumer request %s: %w", c.username, err)
	}
	req.Header.Add(common.ContentType, proxyCommon.UrlEncodedForm)
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.jwt)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send new consumer request %s: %w", c.username, err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
		c.loggingClient.Infof("created consumer (user) '%s'", c.username)
	case http.StatusConflict:
		c.loggingClient.Infof("consumer '%s' already created", c.username)
	default:
		c.loggingClient.Error(string(responseBody))
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
	kongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(), "consumers", c.username, "acls"}, "/")
	c.loggingClient.Infof("Associating consumer to acl using endpoint %s", kongURL)

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return fmt.Errorf("Failed to build request to associate consumer %s to group %s: %w", c.username, c.group, err)
	}
	req.Header.Add(common.ContentType, proxyCommon.UrlEncodedForm)
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.jwt)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to submit request to associate consumer %s to group %s: %w", c.username, c.group, err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
		c.loggingClient.Infof("associated consumer %s to group %s", c.username, c.group)
	case http.StatusConflict:
		c.loggingClient.Infof("consumer %s already associated to group %s", c.username, c.group)
	default:
		c.loggingClient.Error(string(responseBody))
		return fmt.Errorf("failed to associate consumer to group with status: %d", resp.StatusCode)
	}

	return nil
}

// ExecuteAddJwt will add a user to the Kong gateway, assign the user to the
// group specified, and then create a JWT entry for the user.
func (c *cmd) ExecuteAddJwt() (int, error) {
	publicKey, err := os.ReadFile(c.publicKeyPath)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to read public key from file %s: %w", c.publicKeyPath, err)
	}

	// Create a custom consumer in Kong
	if err := c.createConsumer(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	// Assign custom consumer to a group in Kong
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

	kongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(), "consumers", c.username, "jwt"}, "/")
	c.loggingClient.Infof("associating JWT on the endpoint of %s", kongURL)

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to prepare request to associate JWT to user %s: %w", c.username, err)
	}
	req.Header.Add(common.ContentType, proxyCommon.UrlEncodedForm)
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.jwt)
	resp, err := c.client.Do(req)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to send request to associate JWT to user %s: %w", c.username, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	var parsedResponse map[string]interface{}

	switch resp.StatusCode {
	case http.StatusCreated:
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&parsedResponse); err != nil {
			return interfaces.StatusCodeExitWithError, fmt.Errorf("unable to parse associate JWT response: %w", err)
		}
	case http.StatusConflict:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("associate JWT request failed (likely due to duplicate ID) with code: %d", resp.StatusCode)
	default:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("associate JWT request failed with code: %d", resp.StatusCode)
	}

	outputKey := fmt.Sprintf("%s", parsedResponse["key"])

	fmt.Printf("%s\n", outputKey)

	return interfaces.StatusCodeExitNormal, nil
}

// ExecuteAddOAuth2 (placeholder)
func (c *cmd) ExecuteAddOAuth2() (statusCode int, err error) {

	// Create a custom consumer in Kong
	if err := c.createConsumer(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	// Assign custom consumer to a group in Kong
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

	kongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(), "consumers", c.username, "oauth2"}, "/")
	c.loggingClient.Infof("creating oauth application at %s", kongURL)

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, kongURL, strings.NewReader(formVal))
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to prepare request to create oauth application %s: %w", c.username, err)
	}
	req.Header.Add(common.ContentType, proxyCommon.UrlEncodedForm)
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.jwt)
	resp, err := c.client.Do(req)
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("failed to send request to create oauth application %s: %w", c.username, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	var parsedResponse map[string]interface{}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&parsedResponse); err != nil {
			return interfaces.StatusCodeExitWithError, fmt.Errorf("unable to parse associate JWT response: %w", err)
		}
	case http.StatusConflict:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("associate JWT request failed (likely due to duplicate ID) with code: %d", resp.StatusCode)
	default:
		return interfaces.StatusCodeExitWithError, fmt.Errorf("associate JWT request failed with code: %d", resp.StatusCode)
	}

	clientID := fmt.Sprintf("%s", parsedResponse["client_id"])
	clientSecret := fmt.Sprintf("%s", parsedResponse["client_secret"])

	err = json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
		"client_id":     clientID,
		"client_secret": clientSecret,
	})
	if err != nil {
		return interfaces.StatusCodeExitWithError, fmt.Errorf("marshaling of client id and secret failed")
	}

	return interfaces.StatusCodeExitNormal, nil
}
