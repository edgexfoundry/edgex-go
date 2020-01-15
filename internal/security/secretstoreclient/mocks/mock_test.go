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

package mocks

import (
	"net/http"
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockInterfaceType(t *testing.T) {
	// Typecast will fail if doesn't implement interface properly
	var iface SecretStoreClient = &MockSecretStoreClient{}
	assert.NotNil(t, iface)
}

func TestMockHealthCheck(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	mockClient.On("HealthCheck").Return(http.StatusOK, nil)

	rc, err := mockClient.HealthCheck()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockInit(t *testing.T) {
	var initResp InitResponse
	mockClient := &MockSecretStoreClient{}
	mockClient.On("Init", 1, 2, &initResp).Return(http.StatusOK, nil)

	rc, err := mockClient.Init(1, 2, &initResp)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockUnseal(t *testing.T) {
	var initResponse InitResponse
	mockClient := &MockSecretStoreClient{}
	mockClient.On("Unseal", &initResponse).Return(http.StatusOK, nil)

	rc, err := mockClient.Unseal(&initResponse)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockInstallPolicy(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	mockClient.On("InstallPolicy", "fake-token", "foo", "bar").Return(http.StatusOK, nil)

	rc, err := mockClient.InstallPolicy("fake-token", "foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockCreateToken(t *testing.T) {
	params := make(map[string]interface{})
	response := make(map[string]interface{})
	mockClient := &MockSecretStoreClient{}
	mockClient.On("CreateToken", "fake-token", params, response).Return(http.StatusOK, nil)

	rc, err := mockClient.CreateToken("fake-token", params, response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockListAccessors(t *testing.T) {
	var response []string
	mockClient := &MockSecretStoreClient{}
	mockClient.On("ListAccessors", "fake-token", mock.Anything).Return(http.StatusOK, nil)

	rc, err := mockClient.ListAccessors("fake-token", &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockRevokeAccessor(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	mockClient.On("RevokeAccessor", "fake-token", "someaccessor").Return(http.StatusNoContent, nil)

	rc, err := mockClient.RevokeAccessor("fake-token", "someaccessor")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rc)
	mockClient.AssertExpectations(t)
}

func TestMockLookupAccessor(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	mockClient.On("LookupAccessor", "fake-token", "8609694a-cdbc-db9b-d345-e782dbb562ed", mock.Anything).Return(http.StatusOK, nil)

	var tokenMetadata TokenMetadata
	rc, err := mockClient.LookupAccessor("fake-token", "8609694a-cdbc-db9b-d345-e782dbb562ed", &tokenMetadata)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockLookupSelf(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	mockClient.On("LookupSelf", "fake-token", mock.Anything).Return(http.StatusOK, nil)

	var tokenMetadata TokenMetadata
	rc, err := mockClient.LookupSelf("fake-token", &tokenMetadata)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockRevokeSelf(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	mockClient.On("RevokeSelf", "fake-token", mock.Anything).Return(http.StatusOK, nil)

	rc, err := mockClient.RevokeSelf("fake-token")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}

func TestMockRegenRootToken(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	var initResponse InitResponse
	mockClient.On("RegenRootToken", &initResponse, mock.Anything).Return(nil)

	var rootToken string
	err := mockClient.RegenRootToken(&initResponse, &rootToken)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestMockCheckSecretEngineInstalled(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	mockClient.On("CheckSecretEngineInstalled", "fake-token", "secrets/", "kv").Return(true, nil)

	installed, err := mockClient.CheckSecretEngineInstalled("fake-token", "secrets/", "kv")
	assert.NoError(t, err)
	assert.True(t, installed)
	mockClient.AssertExpectations(t)
}

func TestMockEnableKVSecretEngine(t *testing.T) {
	mockClient := &MockSecretStoreClient{}
	mockClient.On("EnableKVSecretEngine", "fake-token", "secrets/", "1").Return(http.StatusOK, nil)

	rc, err := mockClient.EnableKVSecretEngine("fake-token", "secrets/", "1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rc)
	mockClient.AssertExpectations(t)
}
