/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Corp.
 * Copyright 2024-2025 IOTech Ltd
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
 *******************************************************************************/

package openbao

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
)

func (c *Client) HealthCheck() (int, error) {
	// According to the OpenBao API documentation (https://openbao.org/api-docs/system/health/),
	// the /sys/health endpoints of the standby node and perfstandby node
	// will return the status code specified in the activecode field, i.e., 200, when standbyok=true and perfstandbyok=true are set.
	healthAPI := HealthAPI + "?standbyok=true&perfstandbyok=true"

	code, err := c.doRequest(RequestArgs{
		AuthToken:            "",
		Method:               http.MethodGet,
		Path:                 healthAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "health check",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       nil,
	})

	// If code is 0 there was a more serious error that prevented request for executing
	if code != 0 {
		c.lc.Infof("Secret store health check HTTP status: StatusCode: %d", code)
	}

	return code, err
}

func (c *Client) Init(secretThreshold int, secretShares int) (types.InitResponse, error) {
	c.lc.Infof("secret store init strategy (SSS parameters): shares=%d threshold=%d",
		secretShares,
		secretThreshold)

	request := InitRequest{
		SecretShares:    secretShares,
		SecretThreshold: secretThreshold,
	}

	response := types.InitResponse{}
	_, err := c.doRequest(RequestArgs{
		AuthToken:            "",
		Method:               http.MethodPost,
		Path:                 InitAPI,
		JSONObject:           &request,
		BodyReader:           nil,
		OperationDescription: "initialize secret store",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	return response, err
}

func (c *Client) Unseal(keysBase64 []string) error {
	c.lc.Infof("Secret store unsealing Process. Applying key shares.")

	secretShares := len(keysBase64)

	keyCounter := 1
	for _, key := range keysBase64 {
		request := UnsealRequest{Key: key}
		response := UnsealResponse{}

		_, err := c.doRequest(RequestArgs{
			AuthToken:            "",
			Method:               http.MethodPost,
			Path:                 UnsealAPI,
			JSONObject:           &request,
			BodyReader:           nil,
			OperationDescription: "unseal secret store",
			ExpectedStatusCode:   http.StatusOK,
			ResponseObject:       &response,
		})

		if err != nil {
			c.lc.Error(fmt.Sprintf("Error applying key share %d/%d: %s", keyCounter, secretShares, err.Error()))
			return err
		}

		c.lc.Info(fmt.Sprintf("Secret store key share %d/%d successfully applied.", keyCounter, secretShares))
		if !response.Sealed {
			c.lc.Info("Secret store key share threshold reached. Unsealing complete.")
			return nil
		}
		keyCounter++
	}

	return fmt.Errorf("%d", 1)
}

func (c *Client) InstallPolicy(token string, policyName string, policyDocument string) error {
	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPut,
		Path:                 fmt.Sprintf(CreatePolicyPath, url.PathEscape(policyName)),
		JSONObject:           UpdateACLPolicyRequest{Policy: policyDocument},
		BodyReader:           nil,
		OperationDescription: "install policy",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) EnableKVSecretEngine(token string, mountPoint string, kvVersion string) error {
	urlPath := path.Join(MountsAPI, mountPoint)
	parameters := EnableSecretsEngineRequest{
		Type:        KeyValue,
		Description: "key/value secret storage",
		Options: &SecretsEngineOptions{
			Version: kvVersion,
		},
	}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 urlPath,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "update mounts for KV",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) CheckSecretEngineInstalled(token string, mountPoint string, engine string) (bool, error) {
	var response ListSecretEnginesResponse

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodGet,
		Path:                 MountsAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "query mounts for '" + engine + "'",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	if err != nil {
		return false, err
	}

	if mountData := response.Data[mountPoint]; mountData.Type == engine {
		return true, nil
	}

	return false, nil
}

func (c *Client) CreateOrUpdateIdentity(secretStoreToken string, name string, metadata map[string]string, policies []string) (string, error) {
	urlPath := path.Join(namedEntityAPI, name)
	request := CreateUpdateEntityRequest{Metadata: metadata, Policies: policies}
	response := CreateUpdateEntityResponse{}

	code, err := c.doRequest(RequestArgs{
		AuthToken:            secretStoreToken,
		Method:               http.MethodPost,
		Path:                 urlPath,
		JSONObject:           request,
		BodyReader:           nil,
		OperationDescription: "Create/Update Entity",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	// Update doesn't return the entity id; just return blank, look it up later

	if code == http.StatusNoContent {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return response.Data.ID, nil
}

func (c *Client) DeleteIdentity(secretStoreToken string, name string) error {
	urlPath := path.Join(namedEntityAPI, name)

	_, err := c.doRequest(RequestArgs{
		AuthToken:            secretStoreToken,
		Method:               http.MethodDelete,
		Path:                 urlPath,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "Delete Entity",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) LookupIdentity(secretStoreToken string, name string) (string, error) {
	urlPath := path.Join(namedEntityAPI, name)
	response := ReadEntityByNameResponse{}

	code, err := c.doRequest(RequestArgs{
		AuthToken:            secretStoreToken,
		Method:               http.MethodGet,
		Path:                 urlPath,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "Read Entity",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	// 404 is a valid response, either if there are no entities at all
	// or if the one we are looking for isn't there

	if code == http.StatusNotFound {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return response.Data.ID, nil
}

func (c *Client) GetIdentityByEntityId(secretStoreToken string, entityId string) (types.EntityMetadata, error) {
	urlPath := path.Join(idEntityAPI, entityId)
	response := ReadEntityByIdResponse{}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            secretStoreToken,
		Method:               http.MethodGet,
		Path:                 urlPath,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "Read Entity",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	if err != nil {
		return types.EntityMetadata{}, err
	}

	return response.Data, nil
}

func (c *Client) CheckAuthMethodEnabled(token string, mountPoint string, authType string) (bool, error) {
	var response ListSecretEnginesResponse

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodGet,
		Path:                 authAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "query auth engine for " + mountPoint + " (" + authType + ")",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	if err != nil {
		return false, err
	}

	if !strings.HasSuffix(mountPoint, "/") {
		mountPoint += "/"
	}

	if mountData := response.Data[mountPoint]; mountData.Type == authType {
		return true, nil
	}

	return false, nil
}

func (c *Client) EnablePasswordAuth(token string, mountPoint string) error {
	urlPath := path.Join(authAPI, mountPoint)
	parameters := EnableAuthMethodRequest{
		Type: UsernamePasswordAuthMethod,
	}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 urlPath,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "Enable password auth",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) LookupAuthHandle(token string, mountPoint string) (string, error) {
	response := ListAuthMethodsResponse{}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodGet,
		Path:                 authAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "List auth methods",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	if err != nil {
		return "", err
	}

	info, ok := response.Data[mountPoint+"/"]
	if !ok {
		return "", fmt.Errorf("Mount point %s not found", mountPoint)
	}
	return info.Accessor, nil
}

func (c *Client) CreateOrUpdateUser(token string, mountPoint string, username string, password string, tokenTTL string, tokenPolicies []string) error {

	const UserNameRE string = "^[-a-zA-Z0-9_]+$"

	re, err := regexp.Compile(UserNameRE)
	if err != nil {
		return err
	}

	if !re.MatchString(username) {
		return fmt.Errorf("invalid username: %s", username)
	}

	urlPath := path.Join(authMountBase, mountPoint, "users", username)
	parameters := CreateOrUpdateUserRequest{
		Password:      password,
		TokenPeriod:   tokenTTL, // setting Period means that the token is renewable indefinitely
		TokenPolicies: tokenPolicies,
	}

	_, err = c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 urlPath,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "Create or update user",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) DeleteUser(token string, mountPoint string, username string) error {

	const UserNameRE string = "^[-a-zA-Z0-9_]+$"

	re, err := regexp.Compile(UserNameRE)
	if err != nil {
		return err
	}

	if !re.MatchString(username) {
		return fmt.Errorf("invalid username: %s", username)
	}

	urlPath := path.Join(authMountBase, mountPoint, "users", username)

	_, err = c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodDelete,
		Path:                 urlPath,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "Delete user",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) BindUserToIdentity(token string, identityId string, authHandle string, username string) error {
	parameters := CreateEntityAliasRequest{
		Name:          username,
		CanonicalID:   identityId,
		MountAccessor: authHandle,
	}

	code, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 entityAliasAPI,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "Create entity alias (assign login to identity)",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       nil,
	})

	// If this is a duplicate bind (HTTP no content), this is also OK

	if code == http.StatusNoContent {
		return nil
	}

	return err
}

func (c *Client) InternalServiceLogin(token string, authEngine string, username string, password string) (map[string]interface{}, error) {
	urlPath := path.Join(authMountBase, authEngine, "login", username)
	parameters := UserPassLoginRequest{
		Password: password,
	}

	response := make(map[string]interface{})

	_, err := c.doRequest(RequestArgs{
		AuthToken:            "", // No token required to login
		Method:               http.MethodPost,
		Path:                 urlPath,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "login user (service)",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	return response, err

}

func (c *Client) CheckIdentityKeyExists(token string, keyName string) (bool, error) {
	response := ListNamedKeysResponse{}

	code, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               "LIST",
		Path:                 oidcKeyAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "List Named Keys",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	// 404 is a valid response, either if there are no entities at all
	// or if the one we are looking for isn't there

	if code == http.StatusNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, v := range response.Data.Keys {
		if v == keyName {
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) CreateNamedIdentityKey(token string, keyName string, algorithm string) error {

	request := CreateNamedKeyRequest{
		AllowedClientIDs: []string{"*"},
		Algorithm:        algorithm,
	}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 path.Join(oidcKeyAPI, keyName),
		JSONObject:           request,
		BodyReader:           nil,
		OperationDescription: "Create Named Key",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) CreateOrUpdateIdentityRole(token string, roleName string, keyName string, template string, audience string, jwtTTL string) error {

	var templatePointer *string = nil
	if template != "" {
		templatePointer = &template
	}

	request := CreateOrUpdateIdentityRoleRequest{
		ClientID: audience,
		Key:      keyName,
		Template: templatePointer, // optional field
		TokenTTL: jwtTTL,
	}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 path.Join(oidcRoleAPI, roleName),
		JSONObject:           request,
		BodyReader:           nil,
		OperationDescription: "Create OIDC JWT role",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}
