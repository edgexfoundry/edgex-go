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
package cert

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testJSONFileName = "testX509Config_test.json"

func TestNewX509ConfigReader(t *testing.T) {
	// normal case
	reader := NewX509ConfigReader(testJSONFileName)

	assert := assert.New(t)
	assert.NotNil(reader)

	// empty filename case
	emptyFileName := ""
	reader = NewX509ConfigReader(emptyFileName)
	assert.Nil(reader)
}

func TestX509ConfigRead(t *testing.T) {
	// normal case
	if reader := NewX509ConfigReader(testJSONFileName); reader != nil {
		x509Config, readErr := reader.Read()

		assert := assert.New(t)
		assert.Nil(readErr)
		assert.NotEmpty(x509Config)
		assert.Equal("true", x509Config.CreateNewRootCA)
		assert.Equal("./config", x509Config.WorkingDir)
		assert.Equal("pki", x509Config.PKISetupDir)
		assert.Equal("TestCA", x509Config.RootCA.CAName)
		assert.Equal("testhost", x509Config.TLSServer.TLSHost)
	}
}

func TestX509ConfigReadMissingFile(t *testing.T) {
	// fake missing file name
	fakeMissingFile := "fakeMissing.json"
	if reader := NewX509ConfigReader(fakeMissingFile); reader != nil {
		x509Config, readErr := reader.Read()

		assert := assert.New(t)
		assert.NotNil(readErr)
		assert.Empty(x509Config)
	}
}

func TestX509ConfigReadBadFile(t *testing.T) {
	badX509ConfigFile := "badX509Config_test.json.test"
	if reader := NewX509ConfigReader(badX509ConfigFile); reader != nil {
		x509Config, readErr := reader.Read()

		assert := assert.New(t)
		// syntax error
		assert.NotNil(readErr)
		assert.Empty(x509Config)
	}
}

func TestGetPkiOutputDir(t *testing.T) {
	if reader := NewX509ConfigReader(testJSONFileName); reader != nil {
		x509Config, _ := reader.Read()
		pkiDir, err := x509Config.GetPkiOutputDir()

		assert := assert.New(t)
		assert.Nil(err)
		assert.NotEmpty(pkiDir)
		assert.Equal(true, strings.Contains(pkiDir,
			filepath.Join(filepath.Base(x509Config.WorkingDir), x509Config.PKISetupDir, x509Config.RootCA.CAName)))
	}
}

func TestGetCAPemFileName(t *testing.T) {
	if reader := NewX509ConfigReader(testJSONFileName); reader != nil {
		x509Config, _ := reader.Read()
		pemFileName := x509Config.GetCAPemFileName()

		assert.Equal(t, x509Config.RootCA.CAName+certFileExt, pemFileName)
	}
}

func TestGetCAPrivateKeyFileName(t *testing.T) {
	if reader := NewX509ConfigReader(testJSONFileName); reader != nil {
		x509Config, _ := reader.Read()
		prvKeyFileName := x509Config.GetCAPrivateKeyFileName()

		assert.Equal(t, x509Config.RootCA.CAName+skFileExt, prvKeyFileName)
	}
}

func TestGetTLSPemFileName(t *testing.T) {
	if reader := NewX509ConfigReader(testJSONFileName); reader != nil {
		x509Config, _ := reader.Read()
		pemFileName := x509Config.GetTLSPemFileName()

		assert.Equal(t, x509Config.TLSServer.TLSHost+certFileExt, pemFileName)
	}
}

func TestGetTLSPrivateKeyFileName(t *testing.T) {
	if reader := NewX509ConfigReader(testJSONFileName); reader != nil {
		x509Config, _ := reader.Read()
		prvKeyFileName := x509Config.GetTLSPrivateKeyFileName()

		assert.Equal(t, x509Config.TLSServer.TLSHost+skFileExt, prvKeyFileName)
	}
}
