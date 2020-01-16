/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
 * @author: Tingyu Zeng, Dell
 *******************************************************************************/
package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/pkg/token/fileioperformer"
)

type CertificateLoader interface {
	Load() (*CertPair, error)
}

type certificate struct {
	client               internal.HttpCaller
	certPath             string
	tokenPath            string
	secretServiceBaseUrl string
	loggingClient        logger.LoggingClient
}

func NewCertificateLoader(
	r internal.HttpCaller,
	certPath string,
	tokenPath string,
	secretServiceBaseUrl string,
	lc logger.LoggingClient) CertificateLoader {

	return certificate{
		client:               r,
		certPath:             certPath,
		tokenPath:            tokenPath,
		loggingClient:        lc,
		secretServiceBaseUrl: secretServiceBaseUrl,
	}
}

type CertCollect struct {
	Pair CertPair `json:"data"`
}

type CertPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

func (cs certificate) Load() (*CertPair, error) {
	t, err := cs.getAccessToken(cs.tokenPath)
	if err != nil {
		return nil, err
	}
	cp, err := cs.retrieve(t)
	if err != nil {
		return nil, err
	}
	err = cs.validate(cp)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func (cs certificate) getAccessToken(filename string) (string, error) {
	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	tokenLoader := authtokenloader.NewAuthTokenLoader(fileOpener)
	t, err := tokenLoader.Load(filename)
	return t, err
}

func (cs certificate) retrieve(t string) (*CertPair, error) {
	tokens := []string{cs.secretServiceBaseUrl, cs.certPath}
	req, err := http.NewRequest(http.MethodGet, strings.Join(tokens, "/"), nil)
	if err != nil {
		e := fmt.Sprintf("failed to retrieve certificate on path %s with error %s", cs.certPath, err.Error())
		cs.loggingClient.Error(e)
		return nil, err
	}
	req.Header.Add(VaultToken, t)

	resp, err := cs.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to retrieve certificate on path %s with error %s", cs.certPath, err.Error())
		cs.loggingClient.Error(e)
		return nil, err
	}
	defer resp.Body.Close()

	cc := CertCollect{}
	switch resp.StatusCode {
	case http.StatusOK:
		if err = json.NewDecoder(resp.Body).Decode(&cc); err != nil {
			return nil, err
		}
		break
	default:
		err = fmt.Errorf("failed to retrieve certificate on path %s with error code %d", cs.certPath, resp.StatusCode)
		cs.loggingClient.Error(err.Error())
		return nil, err
	}
	return &cc.Pair, nil
}

func (cs certificate) validate(cp *CertPair) error {
	if len(cp.Cert) > 0 && len(cp.Key) > 0 {
		return nil
	}
	return errors.New("empty certificate pair")
}
