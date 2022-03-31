/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package secretstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type passwordGenerator struct {
	generatorImplementation CredentialGenerator
}

// NewPasswordGenerator wires up a pluggable password generator
// or defaults to a built-in implementation if
// the pluggable configuration is missing
func NewPasswordGenerator(lc logger.LoggingClient, passwordProvider string, passwordProviderArgs []string) CredentialGenerator {
	gk := &passwordGenerator{
		generatorImplementation: NewDefaultCredentialGenerator(),
	}
	if passwordProvider != "" {
		pp := NewPasswordProvider(lc, NewDefaultExecRunner())
		err := pp.SetConfiguration(passwordProvider, passwordProviderArgs)
		if err != nil {
			lc.Warnf("Could not configure password generator %s: error: %s", passwordProvider, err.Error())
			return gk // fall-back to builtin
		}
		gk.generatorImplementation = pp
	}
	return gk
}

// Generate delegates password generation to underlying implementation
func (gk *passwordGenerator) Generate(ctx context.Context) (string, error) {
	return gk.generatorImplementation.Generate(ctx)
}

type CredCollect struct {
	Pair UserPasswordPair `json:"data"`
}

type UserPasswordPair struct {
	User     string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Cred struct {
	client               internal.HttpCaller
	rootToken            string
	generator            CredentialGenerator
	secretServiceBaseURL string
	loggingClient        logger.LoggingClient
}

func NewCred(
	caller internal.HttpCaller,
	rootToken string,
	generator CredentialGenerator,
	secretServiceBaseURL string,
	lc logger.LoggingClient) Cred {

	return Cred{
		client:               caller,
		rootToken:            rootToken,
		generator:            generator,
		secretServiceBaseURL: secretServiceBaseURL,
		loggingClient:        lc,
	}
}

func (cr *Cred) AlreadyInStore(path string) (bool, error) {
	pair, err := cr.getUserPasswordPair(path)
	if err != nil {
		if err == errNotFound {
			return false, nil
		}
		return false, err
	}
	if len(pair.User) > 0 && len(pair.Password) > 0 {
		return true, nil
	}
	return false, nil
}

func (cr *Cred) getUserPasswordPair(path string) (*UserPasswordPair, error) {
	pair, err := cr.retrieve(path)
	if err != nil {
		return nil, err
	}
	return pair, nil
}

func (cr *Cred) retrieve(path string) (*UserPasswordPair, error) {
	credUrl, err := cr.credPathURL(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, credUrl, nil)
	if err != nil {
		e := fmt.Errorf("error creating http request: %v", err.Error())
		cr.loggingClient.Error(e.Error())
		return nil, e
	}

	req.Header.Set(VaultToken, cr.rootToken)
	resp, err := cr.client.Do(req)
	if err != nil {
		e := fmt.Errorf("failed to retrieve the credential pair on path %s with error %s", path, err.Error())
		cr.loggingClient.Error(e.Error())
		return nil, e
	}
	defer func() { _ = resp.Body.Close() }()

	cred := CredCollect{}

	if resp.StatusCode == http.StatusNotFound {
		cr.loggingClient.Infof("credential pair NOT found in secret store @/%s, status: %s", path, resp.Status)
		return nil, errNotFound
	} else if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("failed to retrieve the credential pair on path %s with error code %d", path, resp.StatusCode)
		cr.loggingClient.Error(e.Error())
		return nil, e
	}

	if err = json.NewDecoder(resp.Body).Decode(&cred); err != nil {
		e := fmt.Errorf("error decoding json response when retrieving credential pair: %s", err.Error())
		cr.loggingClient.Error(e.Error())
		return nil, e
	}

	return &cred.Pair, nil
}

func (cr *Cred) credPathURL(path string) (string, error) {
	baseURL, err := url.Parse(cr.secretServiceBaseURL)
	if err != nil {
		e := fmt.Errorf("error parsing secret-service url:  %s", err.Error())
		cr.loggingClient.Error(e.Error())
		return "", err
	}

	p, err := url.Parse(path)
	if err != nil {
		e := fmt.Errorf("error parsing secret-service credpath: %s", err.Error())
		cr.loggingClient.Error(e.Error())
		return "", err
	}

	fullURL := baseURL.ResolveReference(p)
	return fullURL.String(), nil
}

// GeneratePassword is a pass-through to the password generator
func (cr *Cred) GeneratePassword(ctx context.Context) (string, error) {
	return cr.generator.Generate(ctx)
}

func (cr *Cred) UploadToStore(pair *UserPasswordPair, path string) error {
	cr.loggingClient.Debug("trying to upload the credential pair into secret store")
	jsonBytes, err := json.Marshal(pair)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(jsonBytes)

	credURL, err := cr.credPathURL(path)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, credURL, body)
	if err != nil {
		e := fmt.Errorf("error creating http request: %v", err.Error())
		cr.loggingClient.Error(e.Error())
		return e
	}

	req.Header.Set(VaultToken, cr.rootToken)
	resp, err := cr.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to upload the credential pair on path %s: %s", path, err.Error())
		cr.loggingClient.Error(e)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		e := fmt.Errorf("failed to load the credential pair to the secret store: %s %s", resp.Status, string(b))
		cr.loggingClient.Error(e.Error())
		return e
	}

	cr.loggingClient.Info("successfully uploaded the credential pair into secret store")
	return nil
}
