//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

const (
	vaultToken = "X-Vault-Token" // nolint:gosec
)

var errNotFound = errors.New("secret NOT found")

type SecretCollect struct {
	Pair map[string]string `json:"data"`
}

type SecretClient struct {
	client               internal.HttpCaller
	rootToken            string
	secretServiceBaseURL string
	loggingClient        logger.LoggingClient
}

func NewSecretClient(
	caller internal.HttpCaller,
	rootToken string,
	secretServiceBaseURL string,
	lc logger.LoggingClient) SecretClient {

	return SecretClient{
		client:               caller,
		rootToken:            rootToken,
		secretServiceBaseURL: secretServiceBaseURL,
		loggingClient:        lc,
	}
}

func (msc *SecretClient) HasSecret(path string, keys ...string) (bool, error) {
	secret, err := msc.getSecret(path)
	if err != nil {
		if err == errNotFound {
			return false, nil
		}
		return false, err
	}
	if len(secret.Pair) == 0 {
		return false, nil
	}
	for _, key := range keys {
		if _, ok := secret.Pair[key]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func (msc *SecretClient) getSecret(path string) (*SecretCollect, error) {
	secretServiceUrl, err := msc.pathURL(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, secretServiceUrl, nil)
	if err != nil {
		e := fmt.Errorf("error creating http request: %v", err)
		msc.loggingClient.Error(e.Error())
		return nil, e
	}

	req.Header.Set(vaultToken, msc.rootToken)
	resp, err := msc.client.Do(req)
	if err != nil {
		e := fmt.Errorf("failed to retrieve the tls secret on path %s with error: %v", path, err)
		msc.loggingClient.Error(e.Error())
		return nil, e
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		msc.loggingClient.Info(fmt.Sprintf("TLS secret NOT found in secret store @/%s, status: %s", path, resp.Status))
		return nil, errNotFound
	} else if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("failed to retrieve the TLS secret on path %s with error code %d", path, resp.StatusCode)
		msc.loggingClient.Error(e.Error())
		return nil, e
	}

	secrets := &SecretCollect{}
	if err = json.NewDecoder(resp.Body).Decode(secrets); err != nil {
		e := fmt.Errorf("error decoding json response when retrieving secret: %v", err)
		msc.loggingClient.Error(e.Error())
		return nil, e
	}

	return secrets, nil
}

func (msc *SecretClient) pathURL(path string) (string, error) {
	baseURL, err := url.Parse(msc.secretServiceBaseURL)
	if err != nil {
		e := fmt.Errorf("error parsing secret-service url: %v", err)
		msc.loggingClient.Error(e.Error())
		return "", err
	}

	p, err := url.Parse(path)
	if err != nil {
		e := fmt.Errorf("error parsing secret-service path: %v", err)
		msc.loggingClient.Error(e.Error())
		return "", err
	}

	fullURL := baseURL.ResolveReference(p)
	return fullURL.String(), nil
}

func (msc *SecretClient) SetSecrets(secret map[string]string, path string) error {
	msc.loggingClient.Debug("trying to upload the TLS secret into secret store")
	jsonBytes, err := json.Marshal(&secret)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(jsonBytes)

	credURL, err := msc.pathURL(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, credURL, body)
	if err != nil {
		return fmt.Errorf("error creating http request: %v", err)
	}

	req.Header.Set(vaultToken, msc.rootToken)
	resp, err := msc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload the TLS secret on path %s: %v", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to load the TLS secret to the secret store: %s %s", resp.Status, string(b))
	}

	msc.loggingClient.Info("successfully uploaded the TLS secret into secret store")
	return nil
}
