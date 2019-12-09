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

func (vc *vaultClient) HealthCheck() (int, error) {
	code, err := vc.doRequest(commonRequestArgs{
		AuthToken:            "",
		Method:               http.MethodGet,
		Path:                 VaultHealthAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "health check",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       nil,
	})

	// Heath check returns 5xx codes when unhealthy;
	// return error object only if we don't get numeric code back
	if code == 0 {
		return 0, err
	}
	vc.logger.Info(fmt.Sprintf("vault health check HTTP status: StatusCode: %d", code))
	return code, nil
}

func (vc *vaultClient) Init(config SecretServiceInfo, vmkWriter io.Writer) (statusCode int, err error) {
	initRequest := InitRequest{
		SecretShares:    config.VaultSecretShares,
		SecretThreshold: config.VaultSecretThreshold,
	}
	initResp := InitResponse{}

	vc.logger.Info(fmt.Sprintf("vault init strategy (SSS parameters): shares=%d threshold=%d", initRequest.SecretShares, initRequest.SecretThreshold))

	code, err := vc.doRequest(commonRequestArgs{
		AuthToken:            "",
		Method:               http.MethodPost,
		Path:                 VaultInitAPI,
		JSONObject:           &initRequest,
		BodyReader:           nil,
		OperationDescription: "initialize secret store",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &initResp,
	})

	err = json.NewEncoder(vmkWriter).Encode(initResp) // Write the init response to disk
	return code, err
}

func (vc *vaultClient) Unseal(config SecretServiceInfo, vmkReader io.Reader) (int, error) {
	vc.logger.Info(fmt.Sprintf("Vault unsealing Process. Applying key shares."))
	initResp := InitResponse{}
	readCloser := fileioperformer.MakeReadCloser(vmkReader)
	defer readCloser.Close()

	if err := json.NewDecoder(vmkReader).Decode(&initResp); err != nil {
		vc.logger.Error(fmt.Sprintf("failed to build the JSON structure from the init response body: %s", err.Error()))
		return 0, err
	}

	keyCounter := 1
	for _, key := range initResp.KeysBase64 {
		unsealResponse := UnsealResponse{}
		code, err := vc.doRequest(commonRequestArgs{
			AuthToken:            "",
			Method:               http.MethodPost,
			Path:                 VaultUnsealAPI,
			JSONObject:           &UnsealRequest{Key: key},
			BodyReader:           nil,
			OperationDescription: "unseal secret store",
			ExpectedStatusCode:   http.StatusOK,
			ResponseObject:       &unsealResponse,
		})

		if err != nil {
			vc.logger.Error(fmt.Sprintf("Error applying key share %d/%d: %s", keyCounter, config.VaultSecretShares, err.Error()))
			return 0, err
		}

		vc.logger.Info(fmt.Sprintf("Vault key share %d/%d successfully applied.", keyCounter, config.VaultSecretShares))
		if !unsealResponse.Sealed {
			vc.logger.Info("Vault key share threshold reached. Unsealing complete.")
			return code, nil
		}
		keyCounter++
	}
	return 0, fmt.Errorf("%d", 1)
}

func (vc *vaultClient) InstallPolicy(token string, policyName string, policyDocument string) (int, error) {
	return vc.doRequest(commonRequestArgs{
		AuthToken:            token,
		Method:               http.MethodPut,
		Path:                 fmt.Sprintf(CreatePolicyPath, url.PathEscape(policyName)),
		JSONObject:           UpdateACLPolicyRequest{Policy: policyDocument},
		BodyReader:           nil,
		OperationDescription: "install policy",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})
}

func (vc *vaultClient) CreateToken(token string, parameters map[string]interface{}, response interface{}) (int, error) {
	return vc.doRequest(commonRequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 CreateTokenAPI,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "create token",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       response,
	})
}

func (vc *vaultClient) ListAccessors(token string, accessors *[]string) (statusCode int, err error) {
	var response ListTokenAccessorsResponse
	code, err := vc.doRequest(commonRequestArgs{
		AuthToken:            token,
		Method:               "LIST",
		Path:                 ListAccessorsAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "list token accessors",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})
	*accessors = response.Data.Keys
	return code, err
}

func (vc *vaultClient) RevokeAccessor(token string, accessor string) (statusCode int, err error) {
	parameters := RevokeTokenAccessorRequest{Accessor: accessor}
	return vc.doRequest(commonRequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 RevokeAccessorAPI,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "revoke token accessor",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})
}

func (vc *vaultClient) LookupAccessor(token string, accessor string, tokenMetadata *TokenMetadata) (statusCode int, err error) {
	parameters := LookupAccessorRequest{Accessor: accessor}
	var response TokenLookupResponse
	code, err := vc.doRequest(commonRequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 LookupAccessorAPI,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "lookup accessor",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})
	*tokenMetadata = response.Data
	return code, err
}

func (vc *vaultClient) LookupSelf(token string, tokenMetadata *TokenMetadata) (statusCode int, err error) {
	var response TokenLookupResponse
	code, err := vc.doRequest(commonRequestArgs{
		AuthToken:            token,
		Method:               http.MethodGet,
		Path:                 LookupSelfAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "lookup self token",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})
	*tokenMetadata = response.Data
	return code, err
}

func (vc *vaultClient) RevokeSelf(token string) (statusCode int, err error) {
	return vc.doRequest(commonRequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 RevokeSelfAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "revoke self token",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})
}
