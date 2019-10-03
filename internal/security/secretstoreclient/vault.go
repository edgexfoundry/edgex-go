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
	"net/http"
	"net/url"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/fileioperformer"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

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
	resp, err := vc.commonJSONRequest(nil, http.MethodGet, VaultHealthAPI, struct{}{})
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
	initResp := InitResponse{}

	vc.logger.Info(fmt.Sprintf("vault init strategy (SSS parameters): shares=%d threshold=%d", initRequest.SecretShares, initRequest.SecretThreshold))

	resp, err := vc.commonJSONRequest(nil, http.MethodPost, VaultInitAPI, &initRequest)
	code, err := vc.processResponse(resp, err, "initialize secret store", http.StatusOK, &initResp)
	err = json.NewEncoder(vmkWriter).Encode(initResp) // Write the init response to disk
	return code, err
}

func (vc *vaultClient) Unseal(config SecretServiceInfo, vmkReader io.Reader) (statusCode int, err error) {
	vc.logger.Info(fmt.Sprintf("Vault unsealing Process. Applying key shares."))
	initResp := InitResponse{}
	readCloser := fileioperformer.MakeReadCloser(vmkReader)
	defer readCloser.Close()

	if err = json.NewDecoder(vmkReader).Decode(&initResp); err != nil {
		vc.logger.Error(fmt.Sprintf("failed to build the JSON structure from the init response body: %s", err.Error()))
		return 0, err
	}

	keyCounter := 1
	for _, key := range initResp.KeysBase64 {
		unsealRequest := UnsealRequest{
			Key: key,
		}
		resp, err := vc.commonJSONRequest(nil, http.MethodPost, VaultUnsealAPI, &unsealRequest)
		if err != nil {
			vc.logger.Error(fmt.Sprintf("failed to send the Vault init request: %s", err.Error()))
			return 0, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("vault unseal request failed with status code: %s", resp.Status)
			vc.logger.Error(err.Error())
			return resp.StatusCode, err
		}

		unsealResponse := UnsealResponse{}
		if err = json.NewDecoder(resp.Body).Decode(&unsealResponse); err != nil {
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

func (vc *vaultClient) InstallPolicy(token string, policyName string, policyDocument string) (statusCode int, err error) {
	path := fmt.Sprintf(CreatePolicyPath, url.PathEscape(policyName))
	request := UpdateACLPolicyRequest{Policy: policyDocument}
	resp, err := vc.commonJSONRequest(&token, http.MethodPut, path, request)
	return vc.processResponse(resp, err, "install policy", http.StatusNoContent, nil)
}

func (vc *vaultClient) CreateToken(token string, parameters map[string]interface{}, response interface{}) (statusCode int, err error) {
	resp, err := vc.commonJSONRequest(&token, http.MethodPost, CreateTokenAPI, parameters)
	return vc.processResponse(resp, err, "create token", http.StatusOK, response)
}

func (vc *vaultClient) buildURL(path string) string {
	return (&url.URL{
		Scheme: vc.scheme,
		Host:   vc.host,
		Path:   path,
	}).String()
}

func (vc *vaultClient) commonJSONRequest(token *string, method string, path string, jsonBody interface{}) (*http.Response, error) {
	body, err := json.Marshal(jsonBody)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to marshal request body: %s", err.Error()))
		return nil, err
	}

	return vc.commonRequest(token, method, path, bytes.NewReader(body))
}

func (vc *vaultClient) commonRequest(token *string, method string, path string, bodyReader io.Reader) (*http.Response, error) {
	url := vc.buildURL(path)
	req, err := http.NewRequest(method, url, bodyReader)
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

// processResponse takes the output of the commonRequest method
// and performs standard processing based the the following:
// operationDescription contains a message describing the operation in logs
// expectedStatusCode is the expected status code for success
// responseObject is the address of a JSON struct, or nil
// If the response is HTTP 200 status and responseObject is not nil,
// this function will attempt to decode the object
// delegate is a callback function called to handle the decoded response
func (vc *vaultClient) processResponse(resp *http.Response, responseError error,
	operationDescription string, expectedStatusCode int,
	responseObject interface{}) (int, error) {

	if responseError != nil {
		vc.logger.Error(fmt.Sprintf("unable to make request to %s failed: %s", operationDescription, responseError.Error()))
		return 0, responseError
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatusCode {
		err := fmt.Errorf("request to %s failed with status: %s", operationDescription, resp.Status)
		vc.logger.Error(err.Error())
		return resp.StatusCode, err
	}

	if responseObject != nil {
		err := json.NewDecoder(resp.Body).Decode(responseObject)
		if err != nil {
			vc.logger.Error(fmt.Sprintf("failed to parse response body: %s", err.Error()))
			return resp.StatusCode, err
		}
	}

	vc.logger.Info(fmt.Sprintf("successfully made request to %s", operationDescription))
	return resp.StatusCode, nil
}
