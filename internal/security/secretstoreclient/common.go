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
)

// parameters structure for request method
type commonRequestArgs struct {
	// Authentication token
	AuthToken string
	// HTTP method
	Method string
	// URL path
	Path string
	// If non-nil, passed to JSON serializer and included in request
	JSONObject interface{}
	// Included in HTTP request if JSONObject is nil
	BodyReader io.Reader
	// Description of the operation being performed included in log messages
	OperationDescription string
	// Expected status code to be returned from HTTP request
	ExpectedStatusCode int
	// If non-nil and request succeeded, response body will be serialized here (must be a pointer)
	ResponseObject interface{}
}

func (vc *vaultClient) doRequest(params commonRequestArgs) (int, error) {
	if params.JSONObject != nil {
		body, err := json.Marshal(params.JSONObject)
		if err != nil {
			vc.logger.Error(fmt.Sprintf("failed to marshal request body: %s", err.Error()))
			return 0, err
		}
		params.BodyReader = bytes.NewReader(body)
	}

	url := (&url.URL{
		Scheme: vc.scheme,
		Host:   vc.host,
		Path:   params.Path,
	}).String()

	req, err := http.NewRequest(params.Method, url, params.BodyReader)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to create request object: %s", err.Error()))
		return 0, err
	}

	if params.AuthToken != "" {
		req.Header.Set(VaultToken, params.AuthToken)
	}
	req.Header.Set("Content-Type", JSONContentType)
	resp, err := vc.client.Do(req)

	if err != nil {
		vc.logger.Error(fmt.Sprintf("unable to make request to %s failed: %s", params.OperationDescription, err.Error()))
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != params.ExpectedStatusCode {
		err := fmt.Errorf("request to %s failed with status: %s", params.OperationDescription, resp.Status)
		vc.logger.Error(err.Error())
		return resp.StatusCode, err
	}

	if params.ResponseObject != nil {
		err := json.NewDecoder(resp.Body).Decode(params.ResponseObject)
		if err != nil {
			vc.logger.Error(fmt.Sprintf("failed to parse response body: %s", err.Error()))
			return resp.StatusCode, err
		}
	}

	vc.logger.Info(fmt.Sprintf("successfully made request to %s", params.OperationDescription))
	return resp.StatusCode, nil
}
