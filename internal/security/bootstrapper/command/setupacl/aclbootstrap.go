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

package setupacl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
)

// generateBootStrapACLToken should only be called once per Consul agent
func (c *cmd) generateBootStrapACLToken() (*types.BootStrapACLTokenInfo, error) {
	aclBootstrapURL, err := c.getRegistryApiUrl(consulACLBootstrapAPI)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, aclBootstrapURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare request for http URL: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request for http URL: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body of bootstrap ACL: %w", err)
	}

	var bootstrapACLToken types.BootStrapACLTokenInfo
	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&bootstrapACLToken); err != nil {
			return nil, fmt.Errorf("failed to decode bootstrapACLToken json data: %v", err)
		}
		return &bootstrapACLToken, nil
	default:
		return nil, fmt.Errorf("failed to bootstrap Consul's ACL via URL [%s] and status code= %d: %s", aclBootstrapURL,
			resp.StatusCode, string(responseBody))
	}
}

func (c *cmd) saveBootstrapACLToken(tokenInfoToBeSaved *types.BootStrapACLTokenInfo) error {
	// Write the token to the specified file
	tokenFileAbsPath, err := filepath.Abs(c.configuration.StageGate.Registry.ACL.BootstrapTokenPath)
	if err != nil {
		return fmt.Errorf("failed to convert tokenFile to absolute path %s: %s",
			c.configuration.StageGate.Registry.ACL.BootstrapTokenPath, err.Error())
	}

	// create the directory of tokenfile if not exists yet
	dirOfToken := filepath.Dir(tokenFileAbsPath)
	fileIoPerformer := fileioperformer.NewDefaultFileIoPerformer()
	if err := fileIoPerformer.MkdirAll(dirOfToken, 0700); err != nil {
		return fmt.Errorf("failed to create tokenpath base dir: %s", err.Error())
	}

	fileWriter, err := fileIoPerformer.OpenFileWriter(tokenFileAbsPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file writer %s: %s", tokenFileAbsPath, err.Error())
	}

	if err := json.NewEncoder(fileWriter).Encode(tokenInfoToBeSaved); err != nil {
		_ = fileWriter.Close()
		return fmt.Errorf("failed to write bootstrap token: %s", err.Error())
	}

	if err := fileWriter.Close(); err != nil {
		return fmt.Errorf("failed to close token file: %s", err.Error())
	}

	c.loggingClient.Infof("bootstrap token is written to %s", tokenFileAbsPath)

	return nil
}
