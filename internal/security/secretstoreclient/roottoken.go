//
// Copyright (c) 2019 Intel Corporation
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
// SPDX-License-Identifier: Apache-2.0'
//

package secretstoreclient

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
)

func (vc *vaultClient) RegenRootToken(initResp *InitResponse, rootToken *string) (err error) {
	// cancel any previous generation attempt
	// start root token generation --> nonce, otp
	// provide keys, nonce --> encoded token
	// encoded token + otp --> root token
	// caller should revoke root token when done with it
	var nonce string
	var otp string
	var encodedToken string

	if err := vc.rootTokenCancelPrevious(); err != nil {
		vc.logger.Warn(fmt.Sprintf("failed to cancel previous root token generation: %s", err.Error()))
		// Not fatal, continue
	}

	if err := vc.rootTokenStartGeneration(&nonce, &otp); err != nil {
		vc.logger.Error(fmt.Sprintf("failed to start root token generation: %s", err.Error()))
		return err
	}

	if err := vc.rootTokenSubmitKeys(initResp, nonce, &encodedToken); err != nil {
		vc.logger.Error(fmt.Sprintf("failed to generate new root token: %s", err.Error()))
		return err
	} else if encodedToken == "" {
		err := errors.New("insufficient key shares to regenerate root token")
		vc.logger.Error(err.Error())
		return err
	}

	if err := vc.rootTokenDecodeToken(encodedToken, otp, rootToken); err != nil {
		vc.logger.Error(fmt.Sprintf("failed to decode root token: %s", err.Error()))
		return err
	}

	return nil
}

func (vc *vaultClient) rootTokenCancelPrevious() error {
	_, err := vc.doRequest(commonRequestArgs{
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

func (vc *vaultClient) rootTokenStartGeneration(nonce *string, otp *string) error {
	var response RootTokenControlResponse
	_, err := vc.doRequest(commonRequestArgs{
		AuthToken:            "",
		Method:               http.MethodPut,
		Path:                 RootTokenControlAPI,
		JSONObject:           struct{}{},
		BodyReader:           nil,
		OperationDescription: "start root token generation",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})
	*nonce = response.Nonce
	*otp = response.Otp
	return err
}

func (vc *vaultClient) rootTokenSubmitKeys(initResp *InitResponse, nonce string, encodedToken *string) error {
	for _, key := range initResp.KeysBase64 {
		complete, err := vc.rootTokenSubmitKey(key, nonce, encodedToken)
		if err != nil {
			vc.logger.Error(fmt.Sprintf("root token retrieval aborted due to error: %s", err.Error()))
			return err
		} else if complete {
			break
		}
	}

	return nil
}

func (vc *vaultClient) rootTokenSubmitKey(key string, nonce string, encodedToken *string) (bool, error) {
	params := RootTokenRetrievalRequest{
		Key:   key,
		Nonce: nonce,
	}
	var response RootTokenRetrievalResponse
	_, err := vc.doRequest(commonRequestArgs{
		AuthToken:            "",
		Method:               http.MethodPut,
		Path:                 RootTokenRetrievalAPI,
		JSONObject:           params,
		BodyReader:           nil,
		OperationDescription: "submit root token keyshare",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &response,
	})
	if response.Complete {
		*encodedToken = response.EncodedToken
	}
	return response.Complete, err
}

func (vc *vaultClient) rootTokenDecodeToken(encodedToken string, otp string, rootToken *string) error {
	// otp is base62 ascii pad cast to raw bytes
	// encodedToken is unpadded (raw) base64
	// XOR the otp bytes with the encoded token bytes to recover the root token

	encodedBytes, err := base64.RawStdEncoding.DecodeString(encodedToken)
	if err != nil {
		vc.logger.Error(fmt.Sprintf("failed to base64 decode the encoded token: %s", err.Error()))
		return err
	}

	otpBytes := []byte(otp)

	if len(encodedBytes) != len(otpBytes) {
		return fmt.Errorf("Invalid input to rootTokenDecodeToken - array length mismatch: %d != %d", len(encodedBytes), len(otpBytes))
	}

	decodedBytes := make([]byte, len(encodedBytes))

	for i := range decodedBytes {
		decodedBytes[i] = encodedBytes[i] ^ otpBytes[i]
	}

	*rootToken = string(decodedBytes)
	return nil
}
