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

package tokenconfig

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/security/fileioperformer"
)

type tokenConfigParser struct {
	fileOpener fileioperformer.FileIoPerformer
	tokenConf  TokenConfFile
}

// NewTokenConfigParser creates a new TokenConfigParser
func NewTokenConfigParser(opener fileioperformer.FileIoPerformer) TokenConfigParser {
	return &tokenConfigParser{fileOpener: opener}
}

func (p *tokenConfigParser) Load(path string) error {
	reader, err := p.fileOpener.OpenFileReader(path, os.O_RDONLY, 0400)
	if err != nil {
		return err
	}
	readCloser := fileioperformer.MakeReadCloser(reader)
	fileContents, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	var parsedContents TokenConfFile
	err = json.Unmarshal(fileContents, &parsedContents)
	if err != nil {
		return err
	}

	p.tokenConf = parsedContents
	return nil
}

func (p *tokenConfigParser) ServiceKeys() []string {
	serviceNames := make([]string, 0, len(p.tokenConf))
	for serviceName := range p.tokenConf {
		serviceNames = append(serviceNames, serviceName)
	}
	return serviceNames
}

func (p tokenConfigParser) GetServiceConfig(service string) ServiceKey {
	return p.tokenConf[service]
}
