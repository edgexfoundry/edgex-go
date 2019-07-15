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
 * @version: 1.1.0
 *******************************************************************************/
package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dghubble/sling"
	"io/ioutil"
	"net/http"
)

type CertConfig interface {
	GetCertPath() string
	GetTokenPath() string
}

type Certs struct {
	Connect Requestor
	Cfg     CertConfig
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

func (cs *Certs) getCertPair() (*CertPair, error) {
	t, err := cs.getSecret(cs.Cfg.GetTokenPath())
	if err != nil {
		return &CertPair{"", ""}, err
	}
	cp, err := cs.retrieve(t)
	if err != nil {
		return &CertPair{"", ""}, err
	}
	err = cs.validate(cp)
	if err != nil {
		return &CertPair{"", ""}, err
	}
	return cp, nil
}

func (cs *Certs) getSecret(filename string) (string, error) {
	a := auth{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return a.Token, err
	}

	err = json.Unmarshal(raw, &a)
	return a.Token, err
}

func (cs *Certs) retrieve(t string) (*CertPair, error) {
	s := sling.New().Set(VaultToken, t)
	req, err := s.New().Base(cs.Connect.GetSecretSvcBaseURL()).Get(cs.Cfg.GetCertPath()).Request()
	resp, err := cs.Connect.GetHTTPClient().Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to retrieve certificate on path %s with error %s", cs.Cfg.GetCertPath(), err.Error())
		lc.Info(e)
		return nil, err
	}
	defer resp.Body.Close()

	cc := CertCollect{}
	switch resp.StatusCode {
	case http.StatusOK:
		if err = json.NewDecoder(resp.Body).Decode(&cc); err != nil {
			return nil, err
		}
	default:
		e := fmt.Sprintf("failed to retrieve certificate on path %s with error code %d", cs.Cfg.GetCertPath(), resp.StatusCode)
		lc.Info(e)
		return nil, err
	}
	return &cc.Pair, nil
}

func (cs *Certs) validate(cp *CertPair) error {
	if len(cp.Cert) > 0 && len(cp.Key) > 0 {
		return nil
	}
	return errors.New("empty certificate pair")
}
