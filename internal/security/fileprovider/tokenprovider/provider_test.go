// Copyright (c) 2019-2023 Intel Corporation
// Copyright (c) 2025 IOTech Ltd
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
// SPDX-License-Identifier: Apache-2.0

package tokenprovider

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	secretstoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"
	loaderMock "github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/authtokenloader/mocks"
	fileMock "github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer/mocks"
	"github.com/edgexfoundry/go-mod-secrets/v4/secrets/mocks"

	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

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
	expectedSvc1TokenRoleParam := map[string]any{"allowed_entity_aliases": []string{"service1"}, "allowed_policies": []string{"edgex-service-service1"}, "name": "service1", "orphan": true, "renewable": true}
	expectedSvc2TokenRoleParam := map[string]any{"allowed_entity_aliases": []string{"service2"}, "allowed_policies": []string{"edgex-service-service2"}, "name": "service2", "orphan": true, "renewable": true}

	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-service1", expectedService1Policy).Return(nil)
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-service2", expectedService2Policy).Return(nil)

	mockSecretStoreClient.On("CreateOrUpdateIdentity", "fake-priv-token", "service1", map[string]string{"name": "service1"}, []string{"edgex-service-service1"}).Return("service1id", nil)
	mockSecretStoreClient.On("CreateOrUpdateUser", "fake-priv-token", "", "service1", mock.AnythingOfType("string"), "", []string{}).Return(nil)
	mockSecretStoreClient.On("LookupAuthHandle", "fake-priv-token", "").Return(`{"data":{"userpass/":{"accessor","accessorid"}}}`, nil)
	mockSecretStoreClient.On("BindUserToIdentity", "fake-priv-token", "service1id", "{\"data\":{\"userpass/\":{\"accessor\",\"accessorid\"}}}", "service1").Return(nil)
	mockSecretStoreClient.On("CreateOrUpdateIdentityRole", "fake-priv-token", "service1", "edgex-identity", "{\"name\": \"service1\"}", "", "").Return(nil)
	mockSecretStoreClient.On("InternalServiceLogin", "fake-priv-token", "", "service1", mock.AnythingOfType("string")).Return(createTokenResponse(), nil)
	mockSecretStoreClient.On("CreateOrUpdateTokenRole", "fake-priv-token", "service1", expectedSvc1TokenRoleParam).Return(nil)

	mockSecretStoreClient.On("CreateOrUpdateIdentity", "fake-priv-token", "service2", map[string]string{"name": "service2"}, []string{"edgex-service-service2"}).Return("service2id", nil)
	mockSecretStoreClient.On("CreateOrUpdateUser", "fake-priv-token", "", "service2", mock.AnythingOfType("string"), "", []string{}).Return(nil)
	mockSecretStoreClient.On("LookupAuthHandle", "fake-priv-token", "").Return(`{"data":{"userpass/":{"accessor","accessorid"}}}`, nil)
	mockSecretStoreClient.On("BindUserToIdentity", "fake-priv-token", "service2id", "{\"data\":{\"userpass/\":{\"accessor\",\"accessorid\"}}}", "service2").Return(nil)
	mockSecretStoreClient.On("CreateOrUpdateIdentityRole", "fake-priv-token", "service2", "edgex-identity", "{\"name\": \"service2\"}", "", "").Return(nil)
	mockSecretStoreClient.On("InternalServiceLogin", "fake-priv-token", "", "service2", mock.AnythingOfType("string")).Return(createTokenResponse(), nil)
	mockSecretStoreClient.On("CreateOrUpdateTokenRole", "fake-priv-token", "service2", expectedSvc2TokenRoleParam).Return(nil)

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
	expectedSvcTokenRoleParam := map[string]any{"allowed_entity_aliases": []string{"myservice"}, "allowed_policies": []string{"edgex-service-myservice"}, "name": "myservice", "orphan": true, "renewable": true}
	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-myservice", expectedService1Policy).Return(nil)
	mockSecretStoreClient.On("CreateOrUpdateIdentity", "fake-priv-token", "myservice", map[string]string{"name": "myservice"}, []string{"edgex-service-myservice"}).Return("myserviceid", nil)
	mockSecretStoreClient.On("CreateOrUpdateUser", "fake-priv-token", "", "myservice", mock.AnythingOfType("string"), "", []string{}).Return(nil)
	mockSecretStoreClient.On("LookupAuthHandle", "fake-priv-token", "").Return(`{"data":{"userpass/":{"accessor","accessorid"}}}`, nil)
	mockSecretStoreClient.On("BindUserToIdentity", "fake-priv-token", "myserviceid", "{\"data\":{\"userpass/\":{\"accessor\",\"accessorid\"}}}", "myservice").Return(nil)
	mockSecretStoreClient.On("CreateOrUpdateIdentityRole", "fake-priv-token", "myservice", "edgex-identity", "{\"name\": \"myservice\"}", "", "").Return(nil)
	mockSecretStoreClient.On("InternalServiceLogin", "fake-priv-token", "", "myservice", mock.AnythingOfType("string")).Return(createTokenResponse(), nil)
	mockSecretStoreClient.On("CreateOrUpdateTokenRole", "fake-priv-token", "myservice", expectedSvcTokenRoleParam).Return(nil)

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
	expectedSvcTokenRoleParam := map[string]any{"allowed_entity_aliases": []string{"myservice"}, "allowed_policies": []string{"edgex-service-myservice"}, "name": "myservice", "orphan": true, "renewable": true}
	service1Buffer := new(bytes.Buffer)
	mockFileIoPerformer.On("MkdirAll", expectedService1Dir, os.FileMode(0700)).Return(nil)
	mockFileIoPerformer.On("OpenFileReader", configFile, os.O_RDONLY, os.FileMode(0400)).Return(strings.NewReader(`{"myservice":{"file_permissions":{"uid":0,"gid":0,"mode_octal":"0664"}}}`), nil)
	mockFileIoPerformer.On("OpenFileWriter", expectedService1File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0600)).Return(&writeCloserBuffer{service1Buffer}, nil)

	mockAuthTokenLoader := &loaderMock.AuthTokenLoader{}
	mockAuthTokenLoader.On("Load", privilegedTokenPath).Return("fake-priv-token", nil)

	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-myservice", "{}").Return(nil)
	mockSecretStoreClient.On("CreateOrUpdateIdentity", "fake-priv-token", "myservice", map[string]string{"name": "myservice"}, []string{"edgex-service-myservice"}).Return("myserviceid", nil)
	mockSecretStoreClient.On("CreateOrUpdateUser", "fake-priv-token", "", "myservice", mock.AnythingOfType("string"), "", []string{}).Return(nil)
	mockSecretStoreClient.On("LookupAuthHandle", "fake-priv-token", "").Return(`{"data":{"userpass/":{"accessor","accessorid"}}}`, nil)
	mockSecretStoreClient.On("BindUserToIdentity", "fake-priv-token", "myserviceid", "{\"data\":{\"userpass/\":{\"accessor\",\"accessorid\"}}}", "myservice").Return(nil)
	mockSecretStoreClient.On("CreateOrUpdateIdentityRole", "fake-priv-token", "myservice", "edgex-identity", "{\"name\": \"myservice\"}", "", "").Return(nil)
	mockSecretStoreClient.On("InternalServiceLogin", "fake-priv-token", "", "myservice", mock.AnythingOfType("string")).Return(createTokenResponse(), nil)
	mockSecretStoreClient.On("CreateOrUpdateTokenRole", "fake-priv-token", "myservice", expectedSvcTokenRoleParam).Return(nil)

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
			"identity/oidc/introspect": map[string]interface{}{
				"capabilities": []string{"create", "update"},
			},
			"identity/oidc/token/" + serviceName: map[string]interface{}{
				"capabilities": []string{"read"},
			},
			"secret/edgex/" + serviceName + "/*": map[string]interface{}{
				"capabilities": []string{"create", "update", "delete", "list", "read"},
			},
		},
	}
	expectedService1Policy, err := json.Marshal(&policy)
	require.NoError(t, err)
	expectedSvcTokenRoleParam1 := map[string]any{"allowed_entity_aliases": []string{serviceName}, "allowed_policies": []string{"edgex-service-" + serviceName}, "name": serviceName, "orphan": true, "renewable": true}
	mockSecretStoreClient := &mocks.SecretStoreClient{}
	mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-"+serviceName, string(expectedService1Policy)).Return(nil)
	mockSecretStoreClient.On("CreateOrUpdateIdentity", "fake-priv-token", serviceName, map[string]string{"name": serviceName}, []string{"edgex-service-" + serviceName}).Return("myserviceid", nil)
	mockSecretStoreClient.On("CreateOrUpdateUser", "fake-priv-token", "", serviceName, mock.AnythingOfType("string"), "1h", []string{}).Return(nil)
	mockSecretStoreClient.On("LookupAuthHandle", "fake-priv-token", "").Return(`{"data":{"userpass/":{"accessor","accessorid"}}}`, nil)
	mockSecretStoreClient.On("BindUserToIdentity", "fake-priv-token", "myserviceid", "{\"data\":{\"userpass/\":{\"accessor\",\"accessorid\"}}}", serviceName).Return(nil)
	mockSecretStoreClient.On("CreateOrUpdateIdentityRole", "fake-priv-token", serviceName, "edgex-identity", "{\"name\": \""+serviceName+"\"}", "", "").Return(nil)
	mockSecretStoreClient.On("InternalServiceLogin", "fake-priv-token", "", serviceName, mock.AnythingOfType("string")).Return(createTokenResponse(), nil)
	mockSecretStoreClient.On("CreateOrUpdateTokenRole", "fake-priv-token", serviceName, expectedSvcTokenRoleParam1).Return(nil)

	// setup expected things for additional services from env if any

	for service := range expectedTokenConfigs {
		policy := map[string]interface{}{
			"path": map[string]interface{}{
				"identity/oidc/introspect": map[string]interface{}{
					"capabilities": []string{"create", "update"},
				},
				"identity/oidc/token/" + service: map[string]interface{}{
					"capabilities": []string{"read"},
				},
				"secret/edgex/" + service + "/*": map[string]interface{}{
					"capabilities": []string{"create", "update", "delete", "list", "read"},
				},
			},
		}
		expectedServicePolicy, err := json.Marshal(&policy)
		require.NoError(t, err)
		expectedSvcTokenRoleParam2 := map[string]any{"allowed_entity_aliases": []string{service}, "allowed_policies": []string{"edgex-service-" + service}, "name": service, "orphan": true, "renewable": true}

		mockSecretStoreClient.On("InstallPolicy", "fake-priv-token", "edgex-service-"+service, string(expectedServicePolicy)).Return(nil)
		mockSecretStoreClient.On("CreateOrUpdateIdentity", "fake-priv-token", service, map[string]string{"name": service}, []string{"edgex-service-" + service}).Return("myserviceid", nil)
		mockSecretStoreClient.On("CreateOrUpdateUser", "fake-priv-token", "", service, mock.AnythingOfType("string"), "1h", []string{}).Return(nil)
		mockSecretStoreClient.On("LookupAuthHandle", "fake-priv-token", "").Return(`{"data":{"userpass/":{"accessor","accessorid"}}}`, nil)
		mockSecretStoreClient.On("BindUserToIdentity", "fake-priv-token", "myserviceid", "{\"data\":{\"userpass/\":{\"accessor\",\"accessorid\"}}}", service).Return(nil)
		mockSecretStoreClient.On("CreateOrUpdateIdentityRole", "fake-priv-token", service, "edgex-identity", "{\"name\": \""+service+"\"}", "", "").Return(nil)
		mockSecretStoreClient.On("InternalServiceLogin", "fake-priv-token", "", service, mock.AnythingOfType("string")).Return(createTokenResponse(), nil)
		mockSecretStoreClient.On("CreateOrUpdateTokenRole", "fake-priv-token", service, expectedSvcTokenRoleParam2).Return(nil)
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
