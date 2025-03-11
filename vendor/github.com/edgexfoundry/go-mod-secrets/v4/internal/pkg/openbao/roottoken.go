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

package openbao

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
)

func (c *Client) RegenRootToken(keys []string) (string, error) {
	// cancel any previous generation attempt
	// start root token generation --> nonce, otp
	// provide keys, nonce --> encoded token
	// encoded token + otp --> root token
	// caller should revoke root token when done with it

	if err := c.rootTokenCancelPrevious(); err != nil {
		c.lc.Warn(fmt.Sprintf("failed to cancel previous root token generation: %s", err.Error()))
		// Not fatal, continue
	}

	nonce, otp, err := c.rootTokenStartGeneration()
	if err != nil {
		c.lc.Error(fmt.Sprintf("failed to start root token generation: %s", err.Error()))
		return "", err
	}

	encodedToken, err := c.rootTokenSubmitKeys(keys, nonce)
	if err != nil {
		c.lc.Error(fmt.Sprintf("failed to generate new root token: %s", err.Error()))
		return "", err
	} else if encodedToken == "" {
		err := errors.New("insufficient key shares to regenerate root token")
		c.lc.Error(err.Error())
		return "", err
	}

	newRootToken, err := c.rootTokenDecodeToken(encodedToken, otp)
	if err != nil {
		c.lc.Error(fmt.Sprintf("failed to decode root token: %s", err.Error()))
		return "", err
	}

	return newRootToken, nil
}

func (c *Client) rootTokenCancelPrevious() error {
	_, err := c.doRequest(RequestArgs{
		AuthToken:            "",
		Method:               http.MethodDelete,
		Path:                 RootTokenControlAPI,
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "cancel previous root token generation",
		ExpectedStatusCode:   http.StatusNoContent,
		ResponseObject:       nil,
	})
	return err
}

func (c *Client) rootTokenStartGeneration() (string, string, error) {
	response := &RootTokenControlResponse{}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            "",
		Method:               http.MethodPut,
		Path:                 RootTokenControlAPI,
		JSONObject:           struct{}{},
		BodyReader:           nil,
		OperationDescription: "start root token generation",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       response,
	})

	return response.Nonce, response.Otp, err
}

func (c *Client) rootTokenSubmitKeys(keys []string, nonce string) (string, error) {
	var encodedToken string

	for _, key := range keys {
		complete, token, err := c.rootTokenSubmitKey(key, nonce)
		if err != nil {
			c.lc.Error(fmt.Sprintf("root token retrieval aborted due to error: %s", err.Error()))
			return "", err
		} else if complete {
			encodedToken = token
			break
		}
	}

	return encodedToken, nil
}

func (c *Client) rootTokenSubmitKey(key string, nonce string) (bool, string, error) {
	params := RootTokenRetrievalRequest{
		Key:   key,
		Nonce: nonce,
	}
	response := &RootTokenRetrievalResponse{}

	_, err := c.doRequest(RequestArgs{
		AuthToken:            "",
		Method:               http.MethodPut,
		Path:                 RootTokenRetrievalAPI,
		JSONObject:           params,
		BodyReader:           nil,
		OperationDescription: "submit root token key share",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       response,
	})

	if err != nil {
		return false, "", err
	}

	var encodedToken string
	if response.Complete {
		encodedToken = response.EncodedToken
	}

	return response.Complete, encodedToken, nil
}

func (c *Client) rootTokenDecodeToken(encodedToken string, otp string) (string, error) {
	// otp is base62 ascii pad cast to raw bytes
	// encodedToken is un-padded (raw) base64
	// XOR the otp bytes with the encoded token bytes to recover the root token

	encodedBytes, err := base64.RawStdEncoding.DecodeString(encodedToken)
	if err != nil {
		c.lc.Error(fmt.Sprintf("failed to base64 decode the encoded token: %s", err.Error()))
		return "", err
	}

	otpBytes := []byte(otp)

	if len(encodedBytes) != len(otpBytes) {
		return "", fmt.Errorf("Invalid input to rootTokenDecodeToken - array length mismatch: %d != %d", len(encodedBytes), len(otpBytes))
	}

	decodedBytes := make([]byte, len(encodedBytes))

	for i := range decodedBytes {
		decodedBytes[i] = encodedBytes[i] ^ otpBytes[i]
	}

	return string(decodedBytes), nil
}
