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
package fileprovider

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/authtokenloader/mocks"
	. "github.com/edgexfoundry/edgex-go/internal/security/fileioperformer/mocks"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	. "github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/*
Test cases:

1. Create multiple service tokens with no defaults
2. Create a service with no defaults and custom policy
3. Create a service with no defaults and custom token parameters
4. Create a service with defaults for policy and token parameters
*/

// TestMultipleTokensWithNoDefaults
func TestMultipleTokensWithNoDefaults(t *testing.T) {
	// Arrange
	privilegedTokenPath := "/dummy/privileged/token.json"
	configFile := "token-config.json"
	outputDir := "/outputdir"
	outputFilename := "secrets-token.json"

	mockLogger := logger.MockLogger{}

	mockFileIoPerformer := &MockFileIoPerformer{}
	expectedService1Dir := filepath.Join(outputDir, "service1")
	expectedService1File := filepath.Join(expectedService1Dir, outputFilename)
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(`{"service1":{},"service2":{}}`), nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)
	expectedService2Dir := filepath.Join(outputDir, "service2")
	expectedService2File := filepath.Join(expectedService2Dir, outputFilename)
	service2Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService2Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService2File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service2Buffer}, nil)

	mockAuthTokenLoader := &MockAuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	expectedService1Policy := "{}"
	expectedService2Policy := "{}"
	expectedService1Parameters := makeMetaServiceName("service1")
	expectedService2Parameters := makeMetaServiceName("service2")
	mockSecretStoreClient := &MockSecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-service1", expectedService1Policy).Return(http.StatusNoContent, nil)
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-service2", expectedService2Policy).Return(http.StatusNoContent, nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters, mock.Anything).
		Run(func(args mock.Arguments) {
			setCreateTokenResponse("service1", args.Get(2).(*interface{}))
		}).
		Return(http.StatusOK, nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService2Parameters, mock.Anything).
		Run(func(args mock.Arguments) {
			setCreateTokenResponse("service2", args.Get(2).(*interface{}))
		}).
		Return(http.StatusOK, nil)

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreclient.SecretServiceInfo{}, TokenFileProviderInfo{
		PrivilegedTokenPath: privilegedTokenPath,
		ConfigFile:          configFile,
		OutputDir:           outputDir,
		OutputFilename:      outputFilename,
	})

	// Act
	err := p.Run()

	// Assert
	// - {OutputDir}/service1/{OutputFilename} w/proper contents
	// - {OutputDir}/service2/{OutputFilename} w/proper contents
	// - Correct policy for service1
	// - Correct policy for service2
	// - All other expectations met
	assert.NoError(t, err)
	mockFileIoPerformer.AssertExpectations(t)
	mockAuthTokenLoader.AssertExpectations(t)
	mockSecretStoreClient.AssertExpectations(t)
	assert.Equal(t, expectedTokenFile("service1"), service1Buffer.Bytes())
	assert.Equal(t, expectedTokenFile("service2"), service2Buffer.Bytes())
}

func setCreateTokenResponse(serviceName string, retval *interface{}) {
	t := make(map[string]interface{})
	t["request_id"] = "f00341c1-fad5-f6e6-13fd-235617f858a1"
	t["auth"] = make(map[string]interface{})
	t["auth"].(map[string]interface{})["client_token"] = "s.wOrq9dO9kzOcuvB06CMviJhZ"
	t["auth"].(map[string]interface{})["accessor"] = "B6oixijqmeR4bsLOJH88Ska9"
	(*retval) = t
}

func makeMetaServiceName(serviceName string) map[string]interface{} {
	createTokenParameters := make(map[string]interface{})
	meta := make(map[string]interface{})
	meta["edgex-service-name"] = serviceName
	createTokenParameters["meta"] = meta
	return createTokenParameters
}

func expectedTokenFile(serviceName string) []byte {
	var tokenResponse interface{}
	setCreateTokenResponse(serviceName, &tokenResponse)
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(tokenResponse)
	// Debugging note: take care to not write out the buffer or it will disturb the read pointer
	return b.Bytes()
}

// TestNoDefaultsCustomPolicy
func TestNoDefaultsCustomPolicy(t *testing.T) {
	// Arrange
	privilegedTokenPath := "/dummy/privileged/token.json"
	configFile := "token-config.json"
	outputDir := "/outputdir"
	outputFilename := "secrets-token.json"

	mockLogger := logger.MockLogger{}

	mockFileIoPerformer := &MockFileIoPerformer{}
	expectedService1Dir := filepath.Join(outputDir, "myservice")
	expectedService1File := filepath.Join(expectedService1Dir, outputFilename)
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(`{"myservice":{"custom_policy":{"path":{"secret/non/standard/location/*":{"capabilities":["list","read"]}}}}}`), nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)

	mockAuthTokenLoader := &MockAuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	expectedService1Policy := `{"path":{"secret/non/standard/location/*":{"capabilities":["list","read"]}}}`
	expectedService1Parameters := makeMetaServiceName("myservice")
	mockSecretStoreClient := &MockSecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-myservice", expectedService1Policy).Return(http.StatusNoContent, nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters, mock.Anything).
		Run(func(args mock.Arguments) {
			setCreateTokenResponse("myservice", args.Get(2).(*interface{}))
		}).
		Return(http.StatusOK, nil)

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreclient.SecretServiceInfo{}, TokenFileProviderInfo{
		PrivilegedTokenPath: privilegedTokenPath,
		ConfigFile:          configFile,
		OutputDir:           outputDir,
		OutputFilename:      outputFilename,
	})

	// Act
	err := p.Run()

	// Assert
	// - {OutputDir}/myservice/{OutputFilename} w/proper contents
	// - Correct policy for myservice
	// - All other expectations met
	assert.NoError(t, err)
	mockFileIoPerformer.AssertExpectations(t)
	mockAuthTokenLoader.AssertExpectations(t)
	mockSecretStoreClient.AssertExpectations(t)
	assert.Equal(t, expectedTokenFile("myservice"), service1Buffer.Bytes())
}

// TestNoDefaultsCustomTokenParameters
func TestNoDefaultsCustomTokenParameters(t *testing.T) {
	// Arrange
	privilegedTokenPath := "/dummy/privileged/token.json"
	configFile := "token-config.json"
	outputDir := "/outputdir"
	outputFilename := "secrets-token.json"

	mockLogger := logger.MockLogger{}

	mockFileIoPerformer := &MockFileIoPerformer{}
	expectedService1Dir := filepath.Join(outputDir, "myservice")
	expectedService1File := filepath.Join(expectedService1Dir, outputFilename)
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(`{"myservice":{"custom_token_parameters":{"key1":"value1"}}}`), nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)

	mockAuthTokenLoader := &MockAuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	expectedService1Policy := "{}"
	expectedService1Parameters := make(map[string]interface{})
	expectedService1Parameters["key1"] = "value1"
	expectedService1Parameters["meta"] = makeMetaServiceName("myservice")["meta"]
	mockSecretStoreClient := &MockSecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-myservice", expectedService1Policy).Return(http.StatusNoContent, nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters, mock.Anything).
		Run(func(args mock.Arguments) {
			setCreateTokenResponse("myservice", args.Get(2).(*interface{}))
		}).
		Return(http.StatusOK, nil)

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreclient.SecretServiceInfo{}, TokenFileProviderInfo{
		PrivilegedTokenPath: privilegedTokenPath,
		ConfigFile:          configFile,
		OutputDir:           outputDir,
		OutputFilename:      outputFilename,
	})

	// Act
	err := p.Run()

	// Assert
	// - {OutputDir}/myservice/{OutputFilename} w/proper contents
	// - Correct token parameters for myservice
	// - All other expectations met
	assert.NoError(t, err)
	mockFileIoPerformer.AssertExpectations(t)
	mockAuthTokenLoader.AssertExpectations(t)
	mockSecretStoreClient.AssertExpectations(t)
	assert.Equal(t, expectedTokenFile("myservice"), service1Buffer.Bytes())
}

// TestTokenUsingDefaults
func TestTokenUsingDefaults(t *testing.T) {
	// Arrange
	privilegedTokenPath := "/dummy/privileged/token.json"
	configFile := "token-config.json"
	outputDir := "/outputdir"
	outputFilename := "secrets-token.json"

	mockLogger := logger.MockLogger{}

	mockFileIoPerformer := &MockFileIoPerformer{}
	expectedService1Dir := filepath.Join(outputDir, "myservice")
	expectedService1File := filepath.Join(expectedService1Dir, outputFilename)
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(`{"myservice":{"edgex_use_defaults":true}}`), nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)

	mockAuthTokenLoader := &MockAuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	expectedService1Policy := `{"path":{"secret/edgex/myservice/*":{"capabilities":["create","update","delete","list","read"]}}}`
	expectedService1Parameters := makeDefaultTokenParameters("myservice")
	expectedService1Parameters["meta"] = makeMetaServiceName("myservice")["meta"]
	mockSecretStoreClient := &MockSecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-myservice", expectedService1Policy).Return(http.StatusNoContent, nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters, mock.Anything).
		Run(func(args mock.Arguments) {
			setCreateTokenResponse("myservice", args.Get(2).(*interface{}))
		}).
		Return(http.StatusOK, nil)

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreclient.SecretServiceInfo{}, TokenFileProviderInfo{
		PrivilegedTokenPath: privilegedTokenPath,
		ConfigFile:          configFile,
		OutputDir:           outputDir,
		OutputFilename:      outputFilename,
	})

	// Act
	err := p.Run()

	// Assert
	// - {OutputDir}/myservice/{OutputFilename} w/proper contents
	// - Correct token parameters for myservice
	// - All other expectations met
	assert.NoError(t, err)
	mockFileIoPerformer.AssertExpectations(t)
	mockAuthTokenLoader.AssertExpectations(t)
	mockSecretStoreClient.AssertExpectations(t)
	assert.Equal(t, expectedTokenFile("myservice"), service1Buffer.Bytes())
}

func TestErrorLoading1(t *testing.T) {
	// Arrange
	mockLogger := logger.MockLogger{}
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockAuthTokenLoader := &MockAuthTokenLoader{}
	mockAuthTokenLoader.On("Load", "tokenpath").Return("atoken", errors.New("an error"))
	mockSecretStoreClient := &MockSecretStoreClient{}

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreclient.SecretServiceInfo{}, TokenFileProviderInfo{
		PrivilegedTokenPath: "tokenpath",
	})

	// Act
	err := p.Run()

	// Assert
	assert.Error(t, err)
	mockFileIoPerformer.AssertExpectations(t)
	mockAuthTokenLoader.AssertExpectations(t)
	mockSecretStoreClient.AssertExpectations(t)
}

func TestErrorLoading2(t *testing.T) {
	// Arrange
	mockLogger := logger.MockLogger{}
	mockFileIoPerformer := &MockFileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "", os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(""), errors.New("an error"))
	mockAuthTokenLoader := &MockAuthTokenLoader{}
	mockAuthTokenLoader.On("Load", "tokenpath").Return("atoken", nil)
	mockSecretStoreClient := &MockSecretStoreClient{}

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreclient.SecretServiceInfo{}, TokenFileProviderInfo{
		PrivilegedTokenPath: "tokenpath",
	})

	// Act
	err := p.Run()

	// Assert
	assert.Error(t, err)
	mockFileIoPerformer.AssertExpectations(t)
	mockAuthTokenLoader.AssertExpectations(t)
	mockSecretStoreClient.AssertExpectations(t)
}

//
// mocks
//

type writeCloserBuffer struct {
	*bytes.Buffer
}

func (wcb *writeCloserBuffer) Close() error {
	return nil
}
