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
 * @author: Daniel Harms, Dell
 * @author: Tingyu Zeng, Dell
 *******************************************************************************/

package secretstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	url2 "net/url"

	"github.com/edgexfoundry/edgex-go/internal"
)

type CertCollect struct {
	Pair CertPair `json:"data"`
}

type CertPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

type Certs struct {
	client    internal.HttpCaller
	certPath  string
	tokenPath string
}

type auth struct {
	Token string `json:"root_token"`
}

func NewCerts(caller internal.HttpCaller, certPath string, tokenPath string) Certs {
	return Certs{client: caller, certPath: certPath, tokenPath: tokenPath}
}

func (cs *Certs) certPathUrl() (string, error) {
	baseURL, err := url2.Parse(Configuration.SecretService.GetSecretSvcBaseURL())
	if err != nil {
		e := fmt.Errorf("error parsing secret-service url.  check server and port properties")
		LoggingClient.Error(e.Error())
		return "", err
	}

	certPath, err := url2.Parse(cs.certPath)
	if err != nil {
		e := fmt.Errorf("error parsing secret-service certpath.  check certpath property")
		LoggingClient.Error(e.Error())
		return "", err
	}

	fullUrl := baseURL.ResolveReference(certPath)
	return fullUrl.String(), nil
}

func (cs *Certs) retrieve(t string) (*CertPair, error) {
	url, err := cs.certPathUrl()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		e := fmt.Errorf("error creating http request: %v", err.Error())
		LoggingClient.Error(e.Error())
		return nil, e
	}

	req.Header.Set(VaultToken, t)
	resp, err := cs.client.Do(req)
	if err != nil {
		e := fmt.Errorf("failed to retrieve the proxy cert on path %s with error %s", cs.certPath, err.Error())
		LoggingClient.Error(e.Error())
		return nil, e
	}
	defer resp.Body.Close()

	cc := CertCollect{}

	if resp.StatusCode == http.StatusNotFound {
		e := fmt.Errorf("proxy cert pair NOT found in secret store @/%s, status: %s", cs.certPath, resp.Status)
		LoggingClient.Info(e.Error())
	} else if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("failed to retrieve the proxy cert pair on path %s with error code %d", cs.certPath, resp.StatusCode)
		LoggingClient.Error(e.Error())
		return nil, e
	}

	if err = json.NewDecoder(resp.Body).Decode(&cc); err != nil {
		e := fmt.Errorf("Error decoding json response when retrieving proxy cert pair: %s", err.Error())
		LoggingClient.Error(e.Error())
		return nil, err
	}

	return &cc.Pair, nil
}

func (cs *Certs) AlreadyinStore() (bool, error) {
	cp, err := cs.getCertPair()
	if err != nil {
		return false, err
	}
	if len(cp.Cert) > 0 && len(cp.Key) > 0 {
		return true, nil
	}
	return false, nil
}

func (cs *Certs) getCertPair() (*CertPair, error) {
	t, err := GetAccessToken(cs.tokenPath)
	if err != nil {
		return &CertPair{"", ""}, err
	}
	cp, err := cs.retrieve(t)
	if err != nil {
		return &CertPair{"", ""}, err
	}
	return cp, nil
}

func (cs *Certs) ReadFrom(certPath string, keyPath string) (*CertPair, error) {
	certPEMBlock, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	cert := string(certPEMBlock)

	keyPEMBlock, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	key := string(keyPEMBlock)

	cp := CertPair{
		Cert: cert,
		Key:  key,
	}

	return &cp, nil
}

func GetAccessToken(filename string) (string, error) {
	a := auth{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(raw, &a)
	if err != nil {
		return "", err
	}

	return a.Token, nil
}

func (cs *Certs) UploadToStore(cp *CertPair) error {
	t, err := GetAccessToken(cs.tokenPath)
	if err != nil {
		return err
	}

	LoggingClient.Info("trying to upload the proxy cert pair into secret store")
	jsonBytes, err := json.Marshal(cp)
	body := bytes.NewBuffer(jsonBytes)

	url, err := cs.certPathUrl()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		e := fmt.Errorf("error creating http request: %v", err.Error())
		LoggingClient.Error(e.Error())
		return e
	}

	req.Header.Set(VaultToken, t)
	resp, err := cs.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to upload the proxy cert pair on path %s with error %s", cs.certPath, err.Error())
		LoggingClient.Error(e)
		return err
	}
	defer resp.Body.Close()

	if !statusSuccess(resp.StatusCode) {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		e := fmt.Errorf("failed to load the proxy cert pair to the secret store: %s,%s", resp.Status, string(b))
		LoggingClient.Error(e.Error())
		return e
	}

	LoggingClient.Info("successful on uploading the proxy cert pair into secret store")
	return nil
}

func statusSuccess(status int) bool {
	return status >= 200 && status < 300
}
