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

package tokenfilewriter

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets/mocks"
)

var lc logger.LoggingClient
var flOpener fileioperformer.FileIoPerformer

func TestMain(m *testing.M) {
	lc = logger.MockLogger{}
	flOpener = fileioperformer.NewDefaultFileIoPerformer()
	os.Exit(m.Run())
}

func TestNewTokenFileWriter(t *testing.T) {
	sc := &mocks.SecretStoreClient{}
	tokenFileWriter := NewWriter(lc, sc, flOpener)
	require.NotEmpty(t, tokenFileWriter)
}

func TestCreateMgmtTokenForConsulSecretsEngine(t *testing.T) {
	createTokenResult := getCreateTokenResultStub()

	testRootToken := "testRootToken"
	testInstallPolicyErr := errors.New("install policy error")
	testCreateTokenErr := errors.New("create token error")

	tests := []struct {
		name               string
		installPolicyError error
		createTokenError   error
	}{
		{"Ok:create token ok", nil, nil},
		{"Bad:install policy error", testInstallPolicyErr, nil},
		{"Bad:create token error", nil, testCreateTokenErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// setup mock secretclient and expected return values for this test
			secretClient := &mocks.SecretStoreClient{}

			secretClient.On("InstallPolicy", testRootToken,
				consulSecretsEngineOpsPolicyName, mock.Anything).
				Return(test.installPolicyError).Once()

			secretClient.On("CreateToken", testRootToken, mock.Anything).
				Return(createTokenResult, test.createTokenError).Once()

			token, _, err := NewWriter(lc, secretClient, flOpener).
				CreateMgmtTokenForConsulSecretsEngine(testRootToken)

			if test.installPolicyError != nil || test.createTokenError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, "createdTestToken", token["auth"].(map[string]interface{})["client_token"])
				secretClient.AssertExpectations(t)
			}
		})
	}
}

func TestCreateAndWriteForConsulSecretEngine(t *testing.T) {
	createTokenResult := getCreateTokenResultStub()

	testInstallPolicyErr := errors.New("install policy error")
	testCreateTokenErr := errors.New("create token error")
	testTokenFileDir := "test"
	testRootToken := "testRootToken"

	tests := []struct {
		name                 string
		rootToken            string
		installPolicyCallErr error
		createTokenCallErr   error
	}{
		{"Ok:CreateAndWrite with client call ok", testRootToken, nil, nil},
		{"Bad:CreateAndWrite with empty token", "", nil, nil},
		{"Bad:CreateAndWrite with install policy call error", testRootToken, testInstallPolicyErr, nil},
		{"Bad:CreateAndWrite with create token call error", testRootToken, nil, testCreateTokenErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// setup mock secretclient and expected return values for this test
			secretClient := &mocks.SecretStoreClient{}

			secretClient.On("InstallPolicy", testRootToken,
				consulSecretsEngineOpsPolicyName, mock.Anything).
				Return(test.installPolicyCallErr).Once()

			secretClient.On("CreateToken", testRootToken, mock.Anything).
				Return(createTokenResult, test.createTokenCallErr).Once()

			testTokenFilePath := filepath.Join(testTokenFileDir,
				"mgmt-token.json"+"_"+strconv.FormatInt(time.Now().UnixNano(), 10))
			fileWriter := NewWriter(lc, secretClient, flOpener)
			_, err := fileWriter.CreateAndWrite(test.rootToken, testTokenFilePath,
				fileWriter.CreateMgmtTokenForConsulSecretsEngine)
			defer func() {
				_ = os.RemoveAll(filepath.Dir(testTokenFilePath))
			}()

			expectError := test.installPolicyCallErr != nil ||
				test.createTokenCallErr != nil ||
				len(test.rootToken) == 0

			if expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.FileExists(t, testTokenFilePath)
			}
		})
	}
}

func TestFileWriteErrorForConsulSecretEngine(t *testing.T) {
	curUser, _ := user.Current()
	if curUser != nil && curUser.Uid == "0" {
		// it is root user then we skip this test as root can have permission to write everything
		t.Log("Skipping this test as it is running as root user")
		t.Skip()
	}

	createTokenResult := getCreateTokenResultStub()

	secretClient := &mocks.SecretStoreClient{}
	testRootToken := "testRootToken"

	secretClient.On("InstallPolicy", testRootToken,
		consulSecretsEngineOpsPolicyName, mock.Anything).
		Return(nil).Once()

	secretClient.On("CreateToken", testRootToken, mock.Anything).
		Return(createTokenResult, nil).Once()

	// literally set to the /root/ directory so that there is no write permission
	// as only the root user can perform this
	testTokenFileDir := "/root/test"
	testTokenFilePath := filepath.Join(testTokenFileDir, "mgmt-token.json")
	fileWriter := NewWriter(lc, secretClient, flOpener)
	_, err := fileWriter.CreateAndWrite(testRootToken, testTokenFilePath,
		fileWriter.CreateMgmtTokenForConsulSecretsEngine)
	require.Error(t, err)
}

func getCreateTokenResultStub() map[string]interface{} {
	createTokenResult := make(map[string]interface{})
	createTokenResult["auth"] = make(map[string]interface{})
	createTokenResult["auth"].(map[string]interface{})["client_token"] = "createdTestToken"
	return createTokenResult
}
