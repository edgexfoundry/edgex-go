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
	"os"
	"path/filepath"
	"strings"
	"testing"

	secretstoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	loaderMock "github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/authtokenloader/mocks"
	fileMock "github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer/mocks"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets/mocks"

	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

/*
Test cases:

1. Create multiple service tokens with no defaults
2. Create a service with no defaults and custom policy
3. Create a service with no defaults and custom token parameters
4. Create a service with defaults for policy and token parameters
*/

const (
	privilegedTokenPath = "/dummy/privileged/token.json"
	configFile          = "token-config.json"
	outputDir           = "/outputdir"
	outputFilename      = "secrets-token.json"
)

// TestMultipleTokensWithNoDefaults
func TestMultipleTokensWithNoDefaults(t *testing.T) {
	// Arrange
	mockLogger := logger.MockLogger{}

	mockFileIoPerformer := &fileMock.FileIoPerformer{}
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

	mockAuthTokenLoader := &loaderMock.AuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	expectedService1Policy := "{}"
	expectedService2Policy := "{}"
	expectedService1Parameters := makeMetaServiceName("service1")
	expectedService2Parameters := makeMetaServiceName("service2")
	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-service1", expectedService1Policy).
		Return(nil)
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-service2", expectedService2Policy).
		Return(nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters).
		Return(createTokenResponse(), nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService2Parameters).
		Return(createTokenResponse(), nil)

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{
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
	assert.Equal(t, expectedTokenFile(), service1Buffer.Bytes())
	assert.Equal(t, expectedTokenFile(), service2Buffer.Bytes())
}

func createTokenResponse() map[string]interface{} {
	// Create some kind of fake response to send back from the SecretStoreClient API
	// Doesn't need to be accurate, as we are not testing the return values from Vault,
	// just making sure we form the call correctly.
	t := make(map[string]interface{})
	t["request_id"] = "f00341c1-fad5-f6e6-13fd-235617f858a1"
	t["auth"] = make(map[string]interface{})
	t["auth"].(map[string]interface{})["client_token"] = "s.wOrq9dO9kzOcuvB06CMviJhZ"
	t["auth"].(map[string]interface{})["accessor"] = "B6oixijqmeR4bsLOJH88Ska9"
	return t
}

func makeMetaServiceName(serviceName string) map[string]interface{} {
	createTokenParameters := make(map[string]interface{})
	meta := make(map[string]interface{})
	meta["edgex-service-name"] = serviceName
	createTokenParameters["meta"] = meta
	return createTokenParameters
}

func expectedTokenFile() []byte {
	tokenResponse := createTokenResponse()
	b := new(bytes.Buffer)
	_ = json.NewEncoder(b).Encode(tokenResponse)
	// Debugging note: take care to not write out the buffer or it will disturb the read pointer
	return b.Bytes()
}

// TestNoDefaultsCustomPolicy
func TestNoDefaultsCustomPolicy(t *testing.T) {
	// Arrange
	mockLogger := logger.MockLogger{}

	mockFileIoPerformer := &fileMock.FileIoPerformer{}
	expectedService1Dir := filepath.Join(outputDir, "myservice")
	expectedService1File := filepath.Join(expectedService1Dir, outputFilename)
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(`{"myservice":{"custom_policy":{"path":{"secret/non/standard/location/*":{"capabilities":["list","read"]}}}}}`), nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)

	mockAuthTokenLoader := &loaderMock.AuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	expectedService1Policy := `{"path":{"secret/non/standard/location/*":{"capabilities":["list","read"]}}}`
	expectedService1Parameters := makeMetaServiceName("myservice")
	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-myservice", expectedService1Policy).
		Return(nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters).
		Return(createTokenResponse(), nil)

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{
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
	assert.Equal(t, expectedTokenFile(), service1Buffer.Bytes())
}

// TestNoDefaultsCustomTokenParameters
func TestNoDefaultsCustomTokenParameters(t *testing.T) {
	// Arrange
	mockLogger := logger.MockLogger{}

	mockFileIoPerformer := &fileMock.FileIoPerformer{}
	expectedService1Dir := filepath.Join(outputDir, "myservice")
	expectedService1File := filepath.Join(expectedService1Dir, outputFilename)
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(`{"myservice":{"custom_token_parameters":{"key1":"value1"}}}`), nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)

	mockAuthTokenLoader := &loaderMock.AuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	expectedService1Policy := "{}"
	expectedService1Parameters := make(map[string]interface{})
	expectedService1Parameters["key1"] = "value1"
	expectedService1Parameters["meta"] = makeMetaServiceName("myservice")["meta"]
	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-myservice", expectedService1Policy).
		Return(nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters).
		Return(createTokenResponse(), nil)

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{
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
	assert.Equal(t, expectedTokenFile(), service1Buffer.Bytes())
}

// TestTokenUsingDefaults
func TestTokenUsingDefaults(t *testing.T) {
	// Good cases:
	// case to mimic normal service name from config file
	err := runTokensWithDefault("myservice", "", t)
	require.NoError(t, err)
	// case to mimic normal service name with one additional service from env list
	err = runTokensWithDefault("myservice2", "additional-service1", t)
	require.NoError(t, err)
	// case to mimic normal service name with two additional services from env list
	err = runTokensWithDefault("myservice3", "additional-service1,new-service-2", t)
	require.NoError(t, err)
	// case to mimic normal service name with two additional services and empty service name from env list
	err = runTokensWithDefault("myservice", "additional-service1,,new-service-2", t)
	require.NoError(t, err)
	// case to mimic normal service name with one additional service and leading + trailing commas from env list
	err = runTokensWithDefault("myservice", ",addtionalservice1,", t)
	require.NoError(t, err)
	// case to mimic normal service name with some additional services with special character names from env list
	err = runTokensWithDefault("myservice", "test.service,test~name", t)
	require.NoError(t, err)

	// Negative cases:
	// case to mimic normal service name with an invalid service name from env list
	err = runTokensWithDefault("myservice", "/service1,,\\new-service-2", t)
	require.Error(t, err, "expect error due to invalid service name from the list in env")
	// case to mimic normal service name with an invalid service name from env list
	err = runTokensWithDefault("myservice", "../service1", t)
	require.Error(t, err, "expect error due to invalid service name from the list in env")
	// case to mimic normal service name with one additional service and URL unsafe characters from env list
	err = runTokensWithDefault("myservice", "core:!@%#$&service*()+func[x]", t)
	require.Error(t, err, "expect error due to invalid service name from the list in env")
}

// TestTokenFilePermissions
func TestTokenFilePermissions(t *testing.T) {
	// Arrange
	mockLogger := logger.MockLogger{}

	mockFileIoPerformer := &fileMock.FileIoPerformer{}
	expectedService1Dir := filepath.Join(outputDir, "myservice")
	expectedService1File := filepath.Join(expectedService1Dir, outputFilename)
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(`{"myservice":{"file_permissions":{"uid":0,"gid":0,"mode_octal":"0664"}}}`), nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)

	mockAuthTokenLoader := &loaderMock.AuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	expectedService1Parameters := makeMetaServiceName("myservice")
	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-myservice", "{}").
		Return(nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters).
		Return(createTokenResponse(), nil)

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{
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
	assert.Equal(t, expectedTokenFile(), service1Buffer.Bytes())
}
func TestErrorLoading1(t *testing.T) {
	// Arrange
	mockLogger := logger.MockLogger{}
	mockFileIoPerformer := &fileMock.FileIoPerformer{}
	mockAuthTokenLoader := &loaderMock.AuthTokenLoader{}
	mockAuthTokenLoader.On("Load", "tokenpath").Return("atoken", errors.New("an error"))
	mockSecretStoreClient := &mocks.SecretStoreClient{}

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{
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
	mockFileIoPerformer := &fileMock.FileIoPerformer{}
	mockFileIoPerformer.On("OpenFileReader", "", os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(""), errors.New("an error"))
	mockAuthTokenLoader := &loaderMock.AuthTokenLoader{}
	mockAuthTokenLoader.On("Load", "tokenpath").Return("atoken", nil)
	mockSecretStoreClient := &mocks.SecretStoreClient{}

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{
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

func (wcb *writeCloserBuffer) Chmod(_ os.FileMode) error {
	return nil
}

func (wcb *writeCloserBuffer) Chown(_ int, _ int) error {
	return nil
}

func runTokensWithDefault(serviceName string, additionalKeysEnv string, t *testing.T) error {
	// Arrange
	mockLogger := logger.MockLogger{}
	originalEnv := os.Getenv(addSecretstoreTokensEnvKey)
	defer func() {
		_ = os.Setenv(addSecretstoreTokensEnvKey, originalEnv)
	}()

	_ = os.Setenv(addSecretstoreTokensEnvKey, additionalKeysEnv)

	mockFileIoPerformer := &fileMock.FileIoPerformer{}
	expectedService1Dir := filepath.Join(outputDir, serviceName)
	expectedService1File := filepath.Join(expectedService1Dir, outputFilename)
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)

	// setup expected behaviors for additional env services
	expectedEnvServiceBufMap := make(map[string]*bytes.Buffer)
	expectedTokenConfigs, errFromEnv := GetTokenConfigFromEnv()

	jsonStr := `{"` + serviceName + `"` + `:{"edgex_use_defaults":true}`
	count := 1
	for service := range expectedTokenConfigs {
		jsonStr += `,`
		jsonStr += `"` + service
		jsonStr += `"` + `:{"edgex_use_defaults":true}`

		count++
	}
	jsonStr += `}`

	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(
		strings.NewReader(jsonStr), nil)

	for service := range expectedTokenConfigs {
		expectedSrvDir := filepath.Join(outputDir, service)
		expectedSrvFile := filepath.Join(expectedSrvDir, outputFilename)
		expectedSrvBuf := new(bytes.Buffer)
		expectedEnvServiceBufMap[service] = expectedSrvBuf

		mockFileIoPerformer.On("MkdirAll", expectedSrvDir, os.FileMode(0700)).Return(nil)
		mockFileIoPerformer.On("OpenFileWriter", expectedSrvFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
			os.FileMode(0600)).Return(&writeCloserBuffer{expectedSrvBuf}, nil)
	}

	mockAuthTokenLoader := &loaderMock.AuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	policy := map[string]interface{}{
		"path": map[string]interface{}{
			"secret/edgex/" + serviceName + "/*": map[string]interface{}{
				"capabilities": []string{"create", "update", "delete", "list", "read"},
			},
			"consul/creds/" + serviceName: map[string]interface{}{
				"capabilities": []string{"read"},
			},
		},
	}
	expectedService1Policy, err := json.Marshal(&policy)
	require.NoError(t, err)
	expectedService1Parameters := makeDefaultTokenParameters(serviceName, "1h", "1h")
	expectedService1Parameters["meta"] = makeMetaServiceName(serviceName)["meta"]
	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-"+serviceName, string(expectedService1Policy)).
		Return(nil)
	mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedService1Parameters, mock.Anything).
		Return(createTokenResponse(), nil)

	// setup expected things for additional services from env if any

	for service := range expectedTokenConfigs {
		policy := map[string]interface{}{
			"path": map[string]interface{}{
				"secret/edgex/" + service + "/*": map[string]interface{}{
					"capabilities": []string{"create", "update", "delete", "list", "read"},
				},
				"consul/creds/" + service: map[string]interface{}{
					"capabilities": []string{"read"},
				},
			},
		}
		expectedServicePolicy, err := json.Marshal(&policy)
		require.NoError(t, err)

		expectedServiceParameters := makeDefaultTokenParameters(service, "1h", "1h")

		expectedServiceParameters["meta"] = makeMetaServiceName(service)["meta"]

		mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-"+service, string(expectedServicePolicy)).
			Return(nil)
		mockSecretStoreClient.On("CreateToken", "fake-priv-token", expectedServiceParameters, mock.Anything).
			Return(createTokenResponse(), nil)
	}

	p := NewTokenProvider(mockLogger, mockFileIoPerformer, mockAuthTokenLoader, mockSecretStoreClient)
	p.SetConfiguration(secretstoreConfig.SecretStoreInfo{}, config.TokenFileProviderInfo{
		PrivilegedTokenPath: privilegedTokenPath,
		ConfigFile:          configFile,
		OutputDir:           outputDir,
		OutputFilename:      outputFilename,
		DefaultTokenTTL:     "1h",
	})

	// Act
	err = p.Run()

	if errFromEnv != nil {
		return errFromEnv
	}

	// Assert
	// - {OutputDir}/myservice/{OutputFilename} w/proper contents
	// - Correct policy for serviceName
	// - Correct token parameters for serviceName
	// - All other expectations met

	assert.NoError(t, err)
	mockFileIoPerformer.AssertExpectations(t)
	mockAuthTokenLoader.AssertExpectations(t)
	mockSecretStoreClient.AssertExpectations(t)
	assert.Equal(t, expectedTokenFile(), service1Buffer.Bytes())

	// verify the expected token files for additional services from env
	for service := range expectedTokenConfigs {
		assert.Equal(t, expectedTokenFile(), expectedEnvServiceBufMap[service].Bytes())
	}

	return nil
}
