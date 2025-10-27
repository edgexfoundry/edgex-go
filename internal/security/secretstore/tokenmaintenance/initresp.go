/*******************************************************************************
 * Copyright 2025 IOTech Ltd
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

package tokenmaintenance

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
)

// LoadInitResponse load the resp-init.json file content from secret store
func LoadInitResponse(
	lc logger.LoggingClient,
	fileOpener fileioperformer.FileIoPerformer,
	secretConfig config.SecretStoreInfo,
	initResponse *types.InitResponse) error {

	absPath := filepath.Join(secretConfig.TokenFolderPath, secretConfig.TokenFile)

	tokenFile, err := fileOpener.OpenFileReader(absPath, os.O_RDONLY, 0400)
	if err != nil {
		lc.Errorf("could not read master key shares file %s: %w", absPath, err)
		return err
	}
	tokenFileCloseable := fileioperformer.MakeReadCloser(tokenFile)
	defer func() { _ = tokenFileCloseable.Close() }()

	decoder := json.NewDecoder(tokenFileCloseable)
	if decoder == nil {
		err := errors.New("failed to create JSON decoder")
		lc.Error(err.Error())
		return err
	}
	if err := decoder.Decode(initResponse); err != nil {
		lc.Errorf("unable to read token file at %s with error: %w", absPath, err)
		return err
	}

	return nil
}
