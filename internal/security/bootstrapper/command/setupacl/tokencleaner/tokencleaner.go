/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *******************************************************************************/

package tokencleaner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/setupacl/share"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
)

const (
	consulReadSelfTokenAPI = "/v1/acl/token/self"
	// require bootstrap token or acl:write permission to delete tokens, %s is token's accessorID
	consulDeleteTokenAPI = "/v1/acl/token/%s"
)

// TokenCleaner cleans up the old consul tokens from Consul server agent
// based on the tokenID stored under the token base directory and token file name
type TokenCleaner struct {
	tokenApiUrlBase string
	tokenBaseDirAbs string
	tokenFileName   string
	lc              logger.LoggingClient
	httpCaller      internal.HttpCaller
}

// SelfTokenInfo is the token information about its token ID and accessor ID
type SelfTokenInfo struct {
	AccessorID string `json:"AccessorID"`
	SecretID   string `json:"SecretID"`
}

// NewTokenCleaner returns a new instance of TokenCleaner
// the tokenFileName cannot be empty
func NewTokenCleaner(tokenApiUrlBase string,
	tokenBaseDir string,
	tokenFileName string,
	lc logger.LoggingClient,
	caller internal.HttpCaller) (*TokenCleaner, error) {
	tokenBaseDirAbs, err := filepath.Abs(tokenBaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for tokenBaseDir: %v", err)
	}

	if len(tokenFileName) == 0 {
		return nil, errors.New("token file name is empty")
	}

	return &TokenCleaner{
		tokenApiUrlBase: tokenApiUrlBase,
		tokenBaseDirAbs: tokenBaseDirAbs,
		tokenFileName:   tokenFileName,
		lc:              lc,
		httpCaller:      caller,
	}, nil
}

// ScrubOldTokens walks through the directories underneath the access token base directory, finds the matching token file name,
// and deletes the old token from registry agent server based on the tokenID read from the token file if any
// the deletion requires the ACL write permission and hence using bootstrap ACL token
// the caller should calls this to clean up the old tokens before creating new tokens, and the caller could ingore
// the returned error from the Scrubber as it is not that critical if unable to delete the old tokens
func (tr *TokenCleaner) ScrubOldTokens(bootstrapACLToken string) error {
	// if the base directory to clean does not even exist, nothing to be cleaned
	if !tr.isBaseDirExist() {
		tr.lc.Infof("the token base directory %s does not exist, skip scrubbing old tokens", tr.tokenBaseDirAbs)
		return nil
	}

	if len(bootstrapACLToken) == 0 {
		return errors.New("required bootstrap token is empty")
	}

	// walk through the base directory and found the matched token file names and
	// read the token out of it and delete the token from Consul
	walkErr := filepath.Walk(tr.tokenBaseDirAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("found error while walking through token base directory %s: %v", tr.tokenBaseDirAbs, err)
		}

		if info.IsDir() {
			return nil
		}

		fileName := filepath.Base(path)
		if fileName != tr.tokenFileName {
			return nil
		}

		token, err := tr.readTokenFromFile(path)
		if err != nil {
			tr.lc.Warnf("found read token from file error: %v", err)
			return err
		}

		// need to lookup self to obtain the token's accessor as Consul's delete token API
		// only take accessor ID
		tokenSelf, err := tr.retrieveSelfToken(path, token)
		if err != nil {
			tr.lc.Warnf("found read self token error: %v", err)
			return err
		} else if tokenSelf == nil {
			return nil
		}

		// purge Consul token from agent server
		if err := tr.deleteToken(path, tokenSelf.AccessorID, bootstrapACLToken); err != nil {
			tr.lc.Warnf("found delete token error: %v", err)
			// ignore the delete error as it will return status code 500 internal server errors if the token to delete
			// doesn't exist, hence here we return nil so that the filepath.WalkFunc will ignore the error and continue
			// in the best effort to delete tokens
			return nil
		}

		tr.lc.Infof("successfully scrubbed old token for %s from server", path)

		return nil
	})

	return walkErr
}

func (tr *TokenCleaner) readTokenFromFile(filePath string) (string, error) {
	fileOpener := fileioperformer.NewDefaultFileIoPerformer()
	reader, err := fileOpener.OpenFileReader(filePath, os.O_RDONLY, 0400)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("failed to open file reader: %v", err)
	}

	token, err := ioutil.ReadAll(reader)
	if err != nil {
		return share.EmptyToken, fmt.Errorf("failed to read token from reader: %v", err)
	}

	return strings.TrimSpace(string(token)), nil
}

func (tr *TokenCleaner) retrieveSelfToken(filepath, token string) (*SelfTokenInfo, error) {
	if len(token) == 0 {
		tr.lc.Infof("tokenID is empty for path %s, skip readTokenSelf", filepath)
		return nil, nil
	}

	readTokenSelfURL := tr.tokenApiUrlBase + consulReadSelfTokenAPI
	req, err := http.NewRequest(http.MethodGet, readTokenSelfURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare readTokenSelfURL request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, token)
	resp, err := tr.httpCaller.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send readTokenSelfURL request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	tokenSelfResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read readTokenSelfURL response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var tokenSelf SelfTokenInfo
		if err := json.NewDecoder(bytes.NewReader(tokenSelfResp)).Decode(&tokenSelf); err != nil {
			return nil, fmt.Errorf("failed to decode ConsulToken json data: %v", err)
		}

		return &tokenSelf, nil
	case http.StatusForbidden:
		// in this case, the self token is not found, so we just return nil
		return nil, nil
	default:
		return nil, fmt.Errorf("failed to read self token for path %s with status code= %d: %s",
			filepath, resp.StatusCode, string(tokenSelfResp))
	}
}

func (tr *TokenCleaner) deleteToken(path, tokenAccessorID, bootstrapACLToken string) error {
	if len(tokenAccessorID) == 0 {
		tr.lc.Infof("tokenAccessorID is empty for path %s, skip deleteToken", path)
		return nil
	}

	delateTokenURL := tr.tokenApiUrlBase + fmt.Sprintf(consulDeleteTokenAPI, tokenAccessorID)
	req, err := http.NewRequest(http.MethodDelete, delateTokenURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("Failed to prepare delateTokenURL request for http URL: %w", err)
	}

	req.Header.Add(share.ConsulTokenHeader, bootstrapACLToken)
	resp, err := tr.httpCaller.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send delateTokenURL request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	tokenSelfResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read delateTokenURL response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if "true" == string(tokenSelfResp) {
			return nil
		}
		return fmt.Errorf("unable to delete old token for service filepath %s", path)
	default:
		return fmt.Errorf("failed to delete old token for service filepath %s with status code= %d: %s",
			path, resp.StatusCode, string(tokenSelfResp))
	}
}

func (tr *TokenCleaner) isBaseDirExist() bool {
	statInfo, err := os.Stat(tr.tokenBaseDirAbs)
	if err == nil {
		return statInfo.IsDir()
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}
