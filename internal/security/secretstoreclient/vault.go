/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2019 Intel Corporation
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
 * @author: Tingyu Zeng, Dell / Alain Pulluelo, ForgeRock AS
 *******************************************************************************/

package secretstoreclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/fileioperformer"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// InitRequest contains a Vault init request regarding the Shamir Secret Sharing (SSS) parameters
type InitRequest struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

// InitResponse contains a Vault init response
type InitResponse struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	RootToken  string   `json:"root_token"`
}

// UnsealRequest contains a Vault unseal request
type UnsealRequest struct {
	Key   string `json:"key"`
	Reset bool   `json:"reset"`
}

// UnsealResponse contains a Vault unseal response
type UnsealResponse struct {
	Sealed   bool `json:"sealed"`
	T        int  `json:"t"`
	N        int  `json:"n"`
	Progress int  `json:"progress"`
}

// UpdateACLPolicyRequest contains a ACL policy create/update request
type UpdateACLPolicyRequest struct {
	Policy string `json:"policy"`
}

type vaultClient struct {
	logger logger.LoggingClient
	client internal.HttpCaller
	scheme string
	host   string
}

func NewSecretStoreClient(logger logger.LoggingClient, r internal.HttpCaller, s string, h string) SecretStoreClient {
	return &vaultClient{
		logger: logger,
		client: r,
		scheme: s,
		host:   h,
	}
}

func (vc *vaultClient) HealthCheck() (statusCode int, err error) {
	url := vc.buildURL(VaultHealthAPI)
	empty := struct{}{}
	resp, err := vc.commonRequest(nil, http.MethodGet, url, &empty)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed on checking status of secret store: %s", err.Error()))
		return 0, err
	}
	defer resp.Body.Close()
	vc.logger.Info(fmt.Sprintf("vault health check HTTP status: %s (StatusCode: %d)", resp.Status, resp.StatusCode))
	return resp.StatusCode, nil
}

func (vc *vaultClient) Init(config SecretServiceInfo, vmkWriter io.Writer) (statusCode int, err error) {
	initRequest := InitRequest{
		SecretShares:    config.VaultSecretShares,
		SecretThreshold: config.VaultSecretThreshold,
	}

	vc.logger.Info(fmt.Sprintf("vault init strategy (SSS parameters): shares=%d threshold=%d", initRequest.SecretShares, initRequest.SecretThreshold))

	url := vc.buildURL(VaultInitAPI)
	resp, err := vc.commonRequest(nil, http.MethodPost, url, &initRequest)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to send Vault init request: %s", err.Error()))
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		vc.logger.Error(fmt.Sprintf("vault init request failed with status: %s", resp.Status))
		return resp.StatusCode, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to fetch the Vault init request response body: %s", err.Error()))
		return 0, err
	}

	initResp := InitResponse{}
	if err = json.Unmarshal(body, &initResp); err != nil {
		vc.logger.Error(fmt.Sprintf("failed to build the JSON structure from the init request response body: %s", err.Error()))
		return 0, err
	}

	_, err = vmkWriter.Write(body)

	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to create Vault init response %s file, HTTP status: %s", config.TokenFolderPath+"/"+config.TokenFile, err.Error()))
		return 0, err
	}

	vc.logger.Info("Vault initialization complete.")
	return resp.StatusCode, nil
}

func (vc *vaultClient) Unseal(config SecretServiceInfo, vmkReader io.Reader) (statusCode int, err error) {
	vc.logger.Info(fmt.Sprintf("Vault unsealing Process. Applying key shares."))
	initResp := InitResponse{}
	readCloser := fileioperformer.MakeReadCloser(vmkReader)
	rawBytes, err := ioutil.ReadAll(readCloser)
	defer readCloser.Close()
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to read the Vault JSON response init file: %s", err.Error()))
		return 0, err
	}

	if err = json.Unmarshal(rawBytes, &initResp); err != nil {
		vc.logger.Error(fmt.Sprintf("failed to build the JSON structure from the init response body: %s", err.Error()))
		return 0, err
	}

	url := vc.buildURL(VaultUnsealAPI)

	keyCounter := 1
	for _, key := range initResp.KeysBase64 {
		unsealRequest := UnsealRequest{
			Key: key,
		}
		resp, err := vc.commonRequest(nil, http.MethodPost, url, &unsealRequest)
		if err != nil {
			vc.logger.Error(fmt.Sprintf("failed to send the Vault init request: %s", err.Error()))
			return 0, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			vc.logger.Error(fmt.Sprintf("vault unseal request failed with status code: %s", resp.Status))
			return resp.StatusCode, err
		}

		unsealedBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			vc.logger.Error(fmt.Sprintf("failed to fetch the Vault unseal request response body: %s", err.Error()))
			return 0, err
		}
		unsealResponse := UnsealResponse{}
		if err = json.Unmarshal(unsealedBody, &unsealResponse); err != nil {
			vc.logger.Error(fmt.Sprintf("failed to build the JSON structure from the unseal request response body: %s", err.Error()))
			return 0, err
		}

		vc.logger.Info(fmt.Sprintf("Vault key share %d/%d successfully applied.", keyCounter, config.VaultSecretShares))
		if !unsealResponse.Sealed {
			vc.logger.Info("Vault key share threshold reached. Unsealing complete.")
			return resp.StatusCode, nil
		}
		keyCounter++
	}
	return 0, fmt.Errorf("%d", 1)
}

func (vc *vaultClient) InstallPolicy(
	token string,
	policyName string,
	policyDocument string) (statusCode int, err error) {

	path := fmt.Sprintf(CreatePolicyPath, url.PathEscape(policyName))
	url := vc.buildURL(path)

	request := UpdateACLPolicyRequest{Policy: policyDocument}
	resp, err := vc.commonRequest(&token, http.MethodPut, url, request)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to send Vault create policy request: %s", err.Error()))
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		vc.logger.Error(fmt.Sprintf("vault create policy request failed with status: %s", resp.Status))
		return resp.StatusCode, err
	}

	vc.logger.Info(fmt.Sprintf("Created vault policy %s", policyName))
	return resp.StatusCode, nil
}

func (vc *vaultClient) CreateToken(token string, parameters map[string]interface{}, response interface{}) (statusCode int, err error) {
	url := vc.buildURL(CreateTokenAPI)

	resp, err := vc.commonRequest(&token, http.MethodPost, url, parameters)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to send Vault create token request: %s", err.Error()))
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		vc.logger.Error(fmt.Sprintf("vault create token request failed with status: %s", resp.Status))
		return resp.StatusCode, err
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to parse response body: %s", err.Error()))
		return 0, err
	}

	tokenName, ok := parameters["display_name"].(string)
	if !ok {
		tokenName = "UNKNOWN DISPLAY_NAME"
	}
	vc.logger.Info(fmt.Sprintf("Created vault token %s", tokenName))
	return resp.StatusCode, nil
}

func (vc *vaultClient) buildURL(path string) string {
	return (&url.URL{
		Scheme: vc.scheme,
		Host:   vc.host,
		Path:   path,
	}).String()
}

func (vc *vaultClient) commonRequest(token *string, method string, url string, jsonBody interface{}) (*http.Response, error) {
	body, err := json.Marshal(jsonBody)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to marshal request body: %s", err.Error()))
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to create request object: %s", err.Error()))
		return nil, err
	}

	if token != nil {
		req.Header.Set(VaultToken, *token)
	}
	req.Header.Set("Content-Type", JSONContentType)
	return vc.client.Do(req)
}
