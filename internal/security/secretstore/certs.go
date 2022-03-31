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
 * @author: Daniel Harms, Dell
 * @author: Tingyu Zeng, Dell
 *******************************************************************************/

package secretstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type CertCollect struct {
	Pair CertPair `json:"data"`
}

type CertPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

type Certs struct {
	client               internal.HttpCaller
	certPath             string
	rootToken            string
	secretServiceBaseURL string
	loggingClient        logger.LoggingClient
}

func NewCerts(
	caller internal.HttpCaller,
	certPath string,
	rootToken string,
	secretServiceBaseURL string,
	lc logger.LoggingClient) Certs {

	return Certs{
		client:               caller,
		certPath:             certPath,
		rootToken:            rootToken,
		secretServiceBaseURL: secretServiceBaseURL,
		loggingClient:        lc,
	}
}

func (cs *Certs) certPathUrl() (string, error) {
	baseURL, err := url.Parse(cs.secretServiceBaseURL)
	if err != nil {
		e := fmt.Errorf("error parsing secret-service url.  check server and port properties")
		cs.loggingClient.Error(e.Error())
		return "", err
	}

	certPath, err := url.Parse(cs.certPath)
	if err != nil {
		e := fmt.Errorf("error parsing secret-service certpath.  check certpath property")
		cs.loggingClient.Error(e.Error())
		return "", err
	}

	fullUrl := baseURL.ResolveReference(certPath)
	return fullUrl.String(), nil
}

func (cs *Certs) retrieve() (*CertPair, error) {
	certUrl, err := cs.certPathUrl()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, certUrl, nil)
	if err != nil {
		e := fmt.Errorf("error creating http request: %v", err.Error())
		cs.loggingClient.Error(e.Error())
		return nil, e
	}

	req.Header.Set(VaultToken, cs.rootToken)
	resp, err := cs.client.Do(req)
	if err != nil {
		e := fmt.Errorf("failed to retrieve the proxy cert on path %s with error %s", cs.certPath, err.Error())
		cs.loggingClient.Error(e.Error())
		return nil, e
	}
	defer func() { _ = resp.Body.Close() }()

	cc := CertCollect{}

	if resp.StatusCode == http.StatusNotFound {
		cs.loggingClient.Infof("proxy cert pair NOT found in secret store @/%s, status: %s", cs.certPath, resp.Status)
		return nil, errNotFound
	} else if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("failed to retrieve the proxy cert pair on path %s with error code %d", cs.certPath, resp.StatusCode)
		cs.loggingClient.Error(e.Error())
		return nil, e
	}

	if err = json.NewDecoder(resp.Body).Decode(&cc); err != nil {
		e := fmt.Errorf("Error decoding json response when retrieving proxy cert pair: %s", err.Error())
		cs.loggingClient.Error(e.Error())
		return nil, e
	}

	return &cc.Pair, nil
}

func (cs *Certs) AlreadyInStore() (bool, error) {
	cp, err := cs.getCertPair()
	if err != nil {
		if err == errNotFound {
			return false, nil
		}
		return false, err
	}
	if len(cp.Cert) > 0 && len(cp.Key) > 0 {
		return true, nil
	}
	return false, nil
}

func (cs *Certs) getCertPair() (*CertPair, error) {
	cp, err := cs.retrieve()
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func (cs *Certs) ReadFrom(certPath string, keyPath string) (*CertPair, error) {
	certPEMBlock, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	cert := string(certPEMBlock)

	keyPEMBlock, err := os.ReadFile(keyPath)
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

func (cs *Certs) UploadToStore(cp *CertPair) error {
	cs.loggingClient.Info("trying to upload the proxy cert pair into secret store")
	jsonBytes, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(jsonBytes)

	certUrl, err := cs.certPathUrl()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, certUrl, body)
	if err != nil {
		e := fmt.Errorf("error creating http request: %v", err.Error())
		cs.loggingClient.Error(e.Error())
		return e
	}

	req.Header.Set(VaultToken, cs.rootToken)
	resp, err := cs.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to upload the proxy cert pair on path %s with error %s", cs.certPath, err.Error())
		cs.loggingClient.Error(e)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		e := fmt.Errorf("failed to load the proxy cert pair to the secret store: %s,%s", resp.Status, string(b))
		cs.loggingClient.Error(e.Error())
		return e
	}

	cs.loggingClient.Info("successful on uploading the proxy cert pair into secret store")
	return nil
}
