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
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
)

type CertificateLoader interface {
	Load() (*CertPair, error)
}

type certificate struct {
	client    internal.HttpCaller
	certPath  string
	tokenPath string
}

func NewCertificateLoader(r internal.HttpCaller, certPath string, tokenPath string) CertificateLoader {
	return certificate{client: r, certPath: certPath, tokenPath: tokenPath}
}

type CertCollect struct {
	Pair CertPair `json:"data"`
}

type CertPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

type auth struct {
	Token string `json:"root_token"`
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
	a := auth{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return a.Token, err
	}

	err = json.Unmarshal(raw, &a)
	return a.Token, err
}

func (cs certificate) retrieve(t string) (*CertPair, error) {
	tokens := []string{Configuration.SecretService.GetSecretSvcBaseURL(), cs.certPath}
	req, err := http.NewRequest(http.MethodGet, strings.Join(tokens, "/"), nil)
	if err != nil {
		e := fmt.Sprintf("failed to retrieve certificate on path %s with error %s", cs.certPath, err.Error())
		LoggingClient.Error(e)
		return nil, err
	}
	req.Header.Add(VaultToken, t)

	resp, err := cs.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to retrieve certificate on path %s with error %s", cs.certPath, err.Error())
		LoggingClient.Error(e)
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
		LoggingClient.Error(err.Error())
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
