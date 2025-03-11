//
// Copyright (c) 2019 Intel Corporation
// Copyright (c) 2024-2025 IOTech Ltd
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

package authtokenloader

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"
)

type tokenProvider struct {
	fileOpener fileioperformer.FileIoPerformer
}

// NewAuthTokenLoader creates a new TokenParser
func NewAuthTokenLoader(opener fileioperformer.FileIoPerformer) AuthTokenLoader {
	return &tokenProvider{fileOpener: opener}
}

func (p *tokenProvider) Load(path string) (authToken string, err error) {
	reader, err := p.fileOpener.OpenFileReader(path, os.O_RDONLY, 0400)
	if err != nil {
		return
	}
	readCloser := fileioperformer.MakeReadCloser(reader)
	fileContents, err := io.ReadAll(readCloser)
	if err != nil {
		return
	}
	defer readCloser.Close()

	var parsedContents secretStoreTokenFile
	err = json.Unmarshal(fileContents, &parsedContents)
	if err != nil {
		return
	}

	// Look for token first in "auth"/"client_token"
	// and then in "root_token"
	// and fail if no token is found at all
	if parsedContents.Auth.ClientToken != "" {
		authToken = parsedContents.Auth.ClientToken
	} else if parsedContents.RootToken != "" {
		authToken = parsedContents.RootToken
	} else {
		err = fmt.Errorf("Unable to find authentication token in %s", path)
	}
	return
}

func (p *tokenProvider) ReadEntityId(path string) (entityId string, err error) {
	reader, err := p.fileOpener.OpenFileReader(path, os.O_RDONLY, 0400)
	if err != nil {
		return
	}
	readCloser := fileioperformer.MakeReadCloser(reader)
	fileContents, err := io.ReadAll(readCloser)
	if err != nil {
		return
	}
	defer readCloser.Close()

	var parsedContents secretStoreTokenFile
	err = json.Unmarshal(fileContents, &parsedContents)
	if err != nil {
		return
	}

	if parsedContents.Auth.EntityId != "" {
		entityId = parsedContents.Auth.EntityId
	} else {
		err = fmt.Errorf("Unable to find entity id in %s", path)
	}
	return
}
