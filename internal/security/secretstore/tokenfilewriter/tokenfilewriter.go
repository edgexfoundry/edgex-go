/*******************************************************************************
 * Copyright 2021 Intel Corporation
 * Copyright (C) 2024 IOTech Ltd
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

package tokenfilewriter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/tokencreatable"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v4/secrets"
)

// TokenFileWriter is a mechanism to generates a token and writes it into a file specified by configuration
type TokenFileWriter struct {
	logClient    logger.LoggingClient
	secretClient secrets.SecretStoreClient
	fileOpener   fileioperformer.FileIoPerformer
}

// NewWriter instantiates a TokenFileWriter instance
func NewWriter(lc logger.LoggingClient,
	sc secrets.SecretStoreClient,
	fileOpener fileioperformer.FileIoPerformer) TokenFileWriter {
	return TokenFileWriter{
		logClient:    lc,
		secretClient: sc,
		fileOpener:   fileOpener,
	}
}

// CreateAndWrite generates a new token and writes it to the file specified by tokenFilePath
// the generation of the token requires root token privilege
// it overwrites the file if already exists
// returns error if anything fails during the whole process
func (w TokenFileWriter) CreateAndWrite(rootToken string, tokenFilePath string,
	createTokenFunc tokencreatable.CreateTokenFunc) (tokencreatable.RevokeFunc, error) {
	if len(rootToken) == 0 {
		return nil, fmt.Errorf("rootToken is required")
	}

	// creates the token
	createTokenFuncName := getFunctionName(createTokenFunc)
	tokenCreated, revokeTokenFunc, createErr := createTokenFunc(rootToken)
	if createErr != nil {
		return nil, fmt.Errorf("failed to create token with %s: %s", createTokenFuncName, createErr.Error())
	}

	w.logClient.Infof("created token with %s", createTokenFuncName)

	var fileErr error
	defer func() {
		// call revokeTokenFunc if there is some fileErr and revokeTokenFunc itself is not nil
		if fileErr != nil && revokeTokenFunc != nil {
			revokeTokenFunc()
		}
	}()

	// Write the created token to the specified file
	tokenFileAbsPath, fileErr := filepath.Abs(tokenFilePath)
	if fileErr != nil {
		return nil, fmt.Errorf("failed to convert tokenFile to absolute path %s: %s", tokenFilePath, fileErr.Error())
	}

	dirOfCreatedToken := filepath.Dir(tokenFileAbsPath)
	fileErr = w.fileOpener.MkdirAll(dirOfCreatedToken, 0700)
	if fileErr != nil {
		return nil, fmt.Errorf("failed to create tokenpath base dir: %s", fileErr.Error())
	}

	fileWriter, fileErr := w.fileOpener.OpenFileWriter(tokenFileAbsPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if fileErr != nil {
		return nil, fmt.Errorf("failed to create token file %s: %s", tokenFileAbsPath, fileErr.Error())
	}

	if fileErr = json.NewEncoder(fileWriter).Encode(tokenCreated); fileErr != nil {
		_ = fileWriter.Close()
		return nil, fmt.Errorf("failed to write created token: %s", fileErr.Error())
	}

	if fileErr = fileWriter.Close(); fileErr != nil {
		return nil, fmt.Errorf("failed to close token file: %s", fileErr.Error())
	}

	w.logClient.Infof("token is written to %s", tokenFilePath)

	return revokeTokenFunc, nil
}

func getFunctionName(f interface{}) string {
	createTokenFuncName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	// On runtime, this will get us something like:
	// github.com/edgexfoundry/edgex-go/internal/security/secretstore.(TokenFileWriter).CreateMgmtTokenForConsulSecretsEngine-fm
	// but we only want to get the last part of string after last "/",
	// i.e. secretstore.(TokenFileWriter).CreateMgmtTokenForConsulSecretsEngine-fm
	elementName := strings.Split(createTokenFuncName, "/")
	return elementName[len(elementName)-1]
}
