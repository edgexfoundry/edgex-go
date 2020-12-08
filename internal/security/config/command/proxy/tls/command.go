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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/config/command/proxy/common"
	"github.com/edgexfoundry/edgex-go/internal/security/config/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const (
	CommandName = "tls"

	builtinSNISList = "localhost,kong"
)

type cmd struct {
	loggingClient           logger.LoggingClient
	client                  internal.HttpCaller
	configuration           *config.ConfigurationStruct
	certificatePath         string
	privateKeyPath          string
	serveNameIndictionsList string
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

	flagSet.StringVar(&cmd.certificatePath, "incert", "", "Path to PEM-encoded leaf certificate")
	flagSet.StringVar(&cmd.privateKeyPath, "inkey", "", "Path to PEM-encoded private key")
	flagSet.StringVar(&cmd.serveNameIndictionsList, "snis", "",
		"[Optional] comma-separated extra server name indications list to associate with this certificate")

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

	if err := c.uploadProxyTlsCert(); err != nil {
		return interfaces.StatusCodeExitWithError, err
	}

	return
}

func (c *cmd) readCertKeyPairFromFiles() (*bootstrapConfig.CertKeyPair, error) {
	certPem, err := ioutil.ReadFile(c.certificatePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read TLS certificate from file %s: %w", c.certificatePath, err)
	}
	prvKey, err := ioutil.ReadFile(c.privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read private key from file %s: %w", c.privateKeyPath, err)
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

	c.loggingClient.Debug(fmt.Sprintf("number of existing Kong tls certs = %d", len(existingCertIDs)))

	if len(existingCertIDs) > 0 {
		// delete the existing certs first
		// ideally, it should only have one certificate to be deleted
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

func (c *cmd) listKongTLSCertificates() ([]string, error) {
	// list certificates first to see if any already exists
	certKongURL := strings.Join([]string{c.configuration.KongURL.GetProxyBaseURL(), "certificates"}, "/")
	c.loggingClient.Info(fmt.Sprintf("list tls certificates on the endpoint of %s", certKongURL))
	req, err := http.NewRequest(http.MethodGet, certKongURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare request to list Kong tls certs: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request to list Kong tls certs: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body to list Kong tls certs: %w", err)
	}

	var parsedResponse map[string]interface{}

	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&parsedResponse); err != nil {
			return nil, fmt.Errorf("Unable to parse response from list certificate: %w", err)
		}
	default:
		return nil, fmt.Errorf("List Kong tls certificates request failed with code: %d", resp.StatusCode)
	}

	var jsonData []byte
	jsonData, err = json.Marshal(parsedResponse["data"])
	if err != nil {
		return nil, fmt.Errorf("Failed to json marshal parsed response data: %w", err)
	}

	outputData := fmt.Sprintf("%s", jsonData)

	// the list certificate get API returns the array of certificates
	var parsedCertData []map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader([]byte(outputData))).Decode(&parsedCertData); err != nil {
		return nil, fmt.Errorf("Unable to parse response for parsed cert data: %w", err)
	}

	certIDs := make([]string, len(parsedCertData))
	for i, certMap := range parsedCertData {
		certIDs[i] = fmt.Sprintf("%s", certMap["id"])
	}

	return certIDs, nil
}

func (c *cmd) deleteKongTLSCertificateById(certId string) error {
	delCertKongURL := strings.Join([]string{c.configuration.KongURL.GetProxyBaseURL(),
		"certificates", certId}, "/")
	c.loggingClient.Info(fmt.Sprintf("deleting tls certificate on the endpoint of %s", delCertKongURL))
	req, err := http.NewRequest(http.MethodDelete, delCertKongURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("Failed to prepare request to delete Kong tls cert: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request to delete Kong tls cert: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		c.loggingClient.Info("Successfully deleted Kong tls cert")
	case http.StatusNotFound:
		// not able to find this certId but should be ok to proceed and post a new certificate
		c.loggingClient.Warn(fmt.Sprintf("Unable to delete Kong tls cert because the certificate Id %s not found", certId))
	default:
		return fmt.Errorf("Delete Kong tls certificate request failed with code: %d", resp.StatusCode)
	}
	return nil
}

func (c *cmd) postKongTLSCertificate(certKeyPair *bootstrapConfig.CertKeyPair) error {
	postCertKongURL := strings.Join([]string{c.configuration.KongURL.GetProxyBaseURL(),
		"certificates"}, "/")
	c.loggingClient.Info(fmt.Sprintf("posting tls certificate on the endpoint of %s", postCertKongURL))

	form := url.Values{
		"cert": []string{certKeyPair.Cert},
		"key":  []string{certKeyPair.Key},
		"snis": getServerNameIndications(c.serveNameIndictionsList),
	}

	formVal := form.Encode()
	req, err := http.NewRequest(http.MethodPost, postCertKongURL, strings.NewReader(formVal))
	if err != nil {
		return fmt.Errorf("Failed to prepare request to post Kong tls cert: %w", err)
	}

	req.Header.Add(clients.ContentType, common.UrlEncodedForm)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request to post Kong tls cert: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response body: %v", err)
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		c.loggingClient.Info("Successfully posted Kong tls cert")
	case http.StatusBadRequest:
		return fmt.Errorf("BadRequest as unable to post Kong tls cert due to error: %s", string(responseBody))
	default:
		return fmt.Errorf("Post Kong tls certificate request failed with code: %d", resp.StatusCode)
	}
	return nil
}

func getServerNameIndications(snisList string) []string {
	// this will parse out the internal default server name indications and extra ones from input if given
	var snis []string
	snis = append(snis, parseAndTrimSpaces(builtinSNISList)...)

	uniqueSnisMap := make(map[string]bool)
	for _, s := range snis {
		uniqueSnisMap[s] = true
	}

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
	splitStrs := strings.Split(commaSep, ",")
	for _, s := range splitStrs {
		trimmed := strings.TrimSpace(s)
		if len(trimmed) > 0 { // effective only it is non-empty string
			ret = append(ret, trimmed)
		}
	}
	return
}
