//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//

package vault

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
)

func (c *Client) CreateToken(token string, parameters map[string]interface{}) (map[string]interface{}, error) {
	response := make(map[string]interface{})

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 CreateTokenAPI,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "create token",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	return response, err
}

func (c *Client) ListTokenAccessors(token string) ([]string, error) {
	var response ListTokenAccessorsResponse

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               "LIST",
		Path:                 ListAccessorsAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "list token accessors",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	if err != nil {
		return nil, err
	}

	return response.Data.Keys, nil
}

func (c *Client) RevokeTokenAccessor(token string, accessor string) error {
	parameters := RevokeTokenAccessorRequest{Accessor: accessor}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 RevokeAccessorAPI,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "revoke token accessor",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}

func (c *Client) LookupTokenAccessor(token string, accessor string) (types.TokenMetadata, error) {
	parameters := LookupAccessorRequest{Accessor: accessor}
	response := &TokenLookupResponse{}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 LookupAccessorAPI,
		JSONObject:           parameters,
		BodyReader:           nil,
		OperationDescription: "lookup accessor",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       response,
	})

	return response.Data, err
}

func (c *Client) LookupToken(token string) (types.TokenMetadata, error) {
	var response TokenLookupResponse

	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodGet,
		Path:                 LookupSelfAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "lookup self token",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})

	return response.Data, err
}

func (c *Client) RevokeToken(token string) error {
	_, err := c.doRequest(RequestArgs{
		AuthToken:            token,
		Method:               http.MethodPost,
		Path:                 RevokeSelfAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "revoke self token",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})

	return err
}
