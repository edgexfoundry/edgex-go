//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0'
//

package tls

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
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
)

const (
	CommandName = "tls"

	builtinSNISList = "localhost,kong"
)

type certificateIDs []string

type cmd struct {
	loggingClient           logger.LoggingClient
	client                  internal.HttpCaller
	configuration           *config.ConfigurationStruct
	certificatePath         string
	privateKeyPath          string
	serverNameIndicatorList string
	adminApiJwt             string
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

	flagSet.StringVar(&cmd.certificatePath, "incert", "", "Path to PEM-encoded leaf certificate")
	flagSet.StringVar(&cmd.privateKeyPath, "inkey", "", "Path to PEM-encoded private key")
	flagSet.StringVar(&cmd.serverNameIndicatorList, "snis", "",
		"[Optional] comma-separated extra server name indications list to associate with this certificate")
	flagSet.StringVar(&cmd.adminApiJwt, "admin_api_jwt", "", "JWT required to interact with local Kong Admin API")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("unable to parse command: %s: %w", strings.Join(args, " "), err)
	}
	if cmd.certificatePath == "" {
		return nil, fmt.Errorf("%s proxy tls: argument --incert is required", os.Args[0])
	}
	if cmd.privateKeyPath == "" {
		return nil, fmt.Errorf("%s proxy tls: argument --inkey is required", os.Args[0])
	}
	if cmd.adminApiJwt == "" {
		return nil, fmt.Errorf("%s proxy tls: argument --admin_api_jwt is required", os.Args[0])
	}

	return &cmd, nil
}

func (c *cmd) Execute() (statusCode int, err error) {

	if err := c.uploadProxyTlsCert(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	return
}

func (c *cmd) readCertKeyPairFromFiles() (*bootstrapConfig.CertKeyPair, error) {
	certPem, err := os.ReadFile(c.certificatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS certificate from file %s: %w", c.certificatePath, err)
	}
	prvKey, err := os.ReadFile(c.privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key from file %s: %w", c.privateKeyPath, err)
	}
	return &bootstrapConfig.CertKeyPair{Cert: string(certPem), Key: string(prvKey)}, nil
}

func (c *cmd) uploadProxyTlsCert() error {
	// try to read both files and make sure they are existing
	certKeyPair, err := c.readCertKeyPairFromFiles()
	if err != nil {
		return err
	}

	// to see if any proxy certificates already exists
	// if yes, then delete them all first
	// and then upload the new TLS certificate
	existingCertIDs, err := c.listKongTLSCertificates()
	if err != nil {
		return err
	}

	c.loggingClient.Debugf("number of existing Kong tls certs = %d", len(existingCertIDs))

	if len(existingCertIDs) > 0 {
		// delete the existing certs first
		// Disclaimer: Kong TLS certificate should only be uploaded via secret-config utility
		for _, certID := range existingCertIDs {
			if err := c.deleteKongTLSCertificateById(certID); err != nil {
				return err
			}
		}
	}

	// post the certKeyPair as the new one
	if err := c.postKongTLSCertificate(certKeyPair); err != nil {
		return err
	}

	return nil
}

func (c *cmd) listKongTLSCertificates() (certificateIDs, error) {
	// definition of Kong snis related objects
	// KongCertObj is the structure for part of certificate object returned from Kong's API /snis
	type KongCertObj struct {
		CertId string `json:"id"`
	}

	// SniCertObj is the structure for part of snis object returned from Kong's API /snis
	type SniCertObj struct {
		SnisName string      `json:"name"`
		KongCert KongCertObj `json:"certificate"`
	}

	// KongSnisObj is the top level structure for part of json data object returned from Kong's API /snis
	type KongSnisObj struct {
		Entries []SniCertObj `json:"data,omitempty"`
	}

	// list snis certificates association to see if any already exists
	certKongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(), "snis"}, "/")
	c.loggingClient.Infof("list snis tls certificates on the endpoint of %s", certKongURL)

	req, err := http.NewRequest(http.MethodGet, certKongURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request to list Kong snis tls certs: %w", err)
	}
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.adminApiJwt)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to list Kong snis tls certs: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body to list Kong snis tls certs: %w", err)
	}

	snisCerts := KongSnisObj{}

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&snisCerts); err != nil {
			return nil, fmt.Errorf("unable to parse response from list snis certificate: %w", err)
		}
	default:
		return nil, fmt.Errorf("list Kong tls snis certificates request failed with code: %d", resp.StatusCode)
	}

	serverNameToCertIDs := make(map[string]certificateIDs)
	for _, entry := range snisCerts.Entries {
		if certIDs, exists := serverNameToCertIDs[entry.SnisName]; !exists {
			// initialize for a new sni name
			serverNameToCertIDs[entry.SnisName] = []string{entry.KongCert.CertId}
		} else {
			// update the existing certificateID array
			certIDs = append(certIDs, entry.KongCert.CertId)
			serverNameToCertIDs[entry.SnisName] = certIDs
		}
	}

	// only get to-be-deleted certificates with which unique certIDs that match the server name indicators
	return c.getUniqueCertIDsMatchServerNames(serverNameToCertIDs), nil
}

func (c *cmd) deleteKongTLSCertificateById(certId string) error {

	// Delete the Kong TLS certificate
	delCertKongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(),
		"certificates", certId}, "/")
	c.loggingClient.Infof("deleting tls certificate on the endpoint of %s", delCertKongURL)

	req, err := http.NewRequest(http.MethodDelete, delCertKongURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to prepare request to delete Kong tls cert: %w", err)
	}
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.adminApiJwt)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to delete Kong tls cert: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNoContent:
		c.loggingClient.Info("Successfully deleted Kong tls cert")
	case http.StatusNotFound:
		// not able to find this certId but should be ok to proceed and post a new certificate
		c.loggingClient.Warnf("Unable to delete Kong tls cert because the certificate Id %s not found", certId)
	default:
		return fmt.Errorf("delete Kong tls certificate request failed with code: %d", resp.StatusCode)
	}
	return nil
}

func (c *cmd) postKongTLSCertificate(certKeyPair *bootstrapConfig.CertKeyPair) error {
	postCertKongURL := strings.Join([]string{c.configuration.KongURL.GetSecureURL(),
		"certificates"}, "/")
	c.loggingClient.Infof("posting tls certificate on the endpoint of %s", postCertKongURL)

	form := url.Values{
		"cert": []string{certKeyPair.Cert},
		"key":  []string{certKeyPair.Key},
		"snis": getServerNameIndicators(c.serverNameIndicatorList),
	}

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, postCertKongURL, strings.NewReader(formVal))
	if err != nil {
		return fmt.Errorf("failed to prepare request to post Kong tls cert: %w", err)
	}

	req.Header.Add(common.ContentType, proxyCommon.UrlEncodedForm)
	req.Header.Add(internal.AuthHeaderTitle, internal.BearerLabel+c.adminApiJwt)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to post Kong tls cert: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		c.loggingClient.Info("Successfully posted Kong tls cert")
	case http.StatusBadRequest:
		return fmt.Errorf("BadRequest as unable to post Kong tls cert due to error: %s", string(responseBody))
	default:
		return fmt.Errorf("post Kong tls certificate request failed with code: %d", resp.StatusCode)
	}
	return nil
}

func (c *cmd) getUniqueCertIDsMatchServerNames(snisToCertIDs map[string]certificateIDs) certificateIDs {
	uniqueSnisFromInput := getServerNameIndicators(c.serverNameIndicatorList)
	// collect the unique certificate ids that match the given snis from the input to be deleted
	uniqueCertIDs := make(map[string]bool)
	for _, sniFromInput := range uniqueSnisFromInput {
		if certIDs, exists := snisToCertIDs[sniFromInput]; exists {
			for _, certID := range certIDs {
				uniqueCertIDs[certID] = true
			}
		}
	}

	// get the unique certificate ids
	certIDs := make([]string, 0, len(uniqueCertIDs))
	for certID := range uniqueCertIDs {
		certIDs = append(certIDs, certID)
	}
	return certIDs
}

func getServerNameIndicators(snisList string) []string {
	// this will parse out the internal default server name indications and extra ones from input if given
	var snis []string
	snis = append(snis, parseAndTrimSpaces(builtinSNISList)...)

	uniqueSnisMap := make(map[string]bool)
	for _, s := range snis {
		uniqueSnisMap[s] = true
	}

	// sanitize the user given list
	if len(snisList) > 0 {
		splitSnis := parseAndTrimSpaces(snisList)
		for _, sni := range splitSnis {
			if _, exists := uniqueSnisMap[sni]; !exists {
				uniqueSnisMap[sni] = true
				snis = append(snis, sni)
			}
		}
	}
	return snis
}

func parseAndTrimSpaces(commaSep string) (ret []string) {
	items := strings.Split(commaSep, ",")
	for _, s := range items {
		trimmed := strings.TrimSpace(s)
		if len(trimmed) > 0 { // effective only it is non-empty string
			ret = append(ret, trimmed)
		}
	}
	return
}
