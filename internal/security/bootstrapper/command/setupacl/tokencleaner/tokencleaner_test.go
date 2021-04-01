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

package tokencleaner

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/command/setupacl/share"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/helper"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
)

func TestNewTokenCleaner(t *testing.T) {
	lc := logger.MockLogger{}
	httpCaller := pkg.NewRequester(lc).Insecure()
	testTokenBaseDir := "test-base-dir"
	testTokenFileName := "test-token"

	tests := []struct {
		name          string
		tokenBaseDir  string
		tokenFileName string
		expectedErr   bool
	}{
		{"Good:new token cleaner ok", testTokenBaseDir, testTokenFileName, false},
		{"Bad:new token cleaner with empty token file name", testTokenBaseDir, "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCleaner, err := NewTokenCleaner("http://localhost:18500", test.tokenBaseDir, test.tokenFileName, lc, httpCaller)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tokenCleaner)
			}
		})
	}
}

func TestScrubOldToken(t *testing.T) {
	lc := logger.MockLogger{}
	testTokenBaseDir := "test-base-dir"
	testServiceKey := "service-test"
	testTokenFileName := "test-token"
	testToken := "token-x"
	testBootstrapToken := "test-bootstrap-token"

	// prepare test
	serviceDir := filepath.Join(testTokenBaseDir, testServiceKey)
	err := helper.CreateDirectoryIfNotExists(serviceDir)
	require.NoError(t, err)
	additionalDir := filepath.Join(testTokenBaseDir, "another")
	err = helper.CreateDirectoryIfNotExists(additionalDir)
	require.NoError(t, err)

	// write a token into service directory
	tokenFilePath := filepath.Join(serviceDir, testTokenFileName)
	err = ioutil.WriteFile(tokenFilePath, []byte(testToken), 0600)
	require.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(serviceDir, "anotherFile"), []byte("testingAnother"), 0600)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(testTokenBaseDir)
	}()

	tests := []struct {
		name           string
		tokenBaseDir   string
		tokenFileName  string
		bootstrapToken string
		clientCaller   internal.HttpCaller
		expectedErr    bool
	}{
		{"Good:scrubOldToken ok with skipping non-existing base dir", "non-exist", testTokenFileName, testBootstrapToken, nil, false},
		{"Good:scrubOldToken ok with existing token", testTokenBaseDir, testTokenFileName, testBootstrapToken, &mockAuthHttpCaller{
			authTokenHeader:  share.ConsulTokenHeader,
			authToken:        "token-x",
			getStatusCode:    200,
			returnError:      false,
			getResponse:      `{"AccessorID":"xxxx", "SecretID":"token-x"}`,
			deleteStatusCode: 200,
			deleteResponse:   "true",
		}, false},
		{"Bad:scrubOldToken with empty bootstrap token", testTokenBaseDir, testTokenFileName, share.EmptyToken, nil, true},
		{"Bad:scrubOldToken read self token error", testTokenBaseDir, testTokenFileName, testBootstrapToken, &mockAuthHttpCaller{
			authTokenHeader: share.ConsulTokenHeader,
			authToken:       "token-y",
			getStatusCode:   403,
			returnError:     false,
			getResponse:     "permission denied",
		}, true},
		// ignore the error and thus treat it as good case in the best effort to delete the tokens
		{"Good:scrubOldToken delete token error", testTokenBaseDir, testTokenFileName, "token-z", &mockAuthHttpCaller{
			authTokenHeader:  share.ConsulTokenHeader,
			authToken:        "token-x",
			getStatusCode:    200,
			returnError:      false,
			getResponse:      `{"AccessorID":"xxxx", "SecretID":"token-x"}`,
			deleteStatusCode: 403,
			deleteResponse:   "false",
		}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCleaner, err := NewTokenCleaner("http://localhost:18500", test.tokenBaseDir, test.tokenFileName, lc, test.clientCaller)
			require.NoError(t, err)
			require.NotNil(t, tokenCleaner)
			err = tokenCleaner.ScrubOldTokens(test.bootstrapToken)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReadTokenFromFile(t *testing.T) {
	lc := logger.MockLogger{}
	httpCaller := pkg.NewRequester(lc).Insecure()
	testServiceKey := "test-service-key"
	testTokenBaseDir := "test-base-dir"
	testTokenFileName := "test-token"

	// prepare test
	serviceDir := filepath.Join(testTokenBaseDir, testServiceKey)
	err := helper.CreateDirectoryIfNotExists(serviceDir)
	require.NoError(t, err)
	testToken := "token-x"
	// write a token into service directory
	tokenFilePath := filepath.Join(serviceDir, testTokenFileName)
	err = ioutil.WriteFile(tokenFilePath, []byte(testToken), 0600)
	require.NoError(t, err)

	tests := []struct {
		name          string
		tokenBaseDir  string
		tokenFileName string
		expectedErr   bool
	}{
		{"Good:token file path exists", testTokenBaseDir, testTokenFileName, false},
		{"Bad:token file path not exists ", testTokenBaseDir, "not-exists", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			defer func() {
				_ = os.RemoveAll(test.tokenBaseDir)
			}()

			tokenCleaner, err := NewTokenCleaner("http://localhost:18500", test.tokenBaseDir, test.tokenFileName, lc, httpCaller)
			require.NoError(t, err)
			require.NotNil(t, tokenCleaner)
			testFilePath := filepath.Join(serviceDir, test.tokenFileName)
			actualToken, err := tokenCleaner.readTokenFromFile(testFilePath)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, "token-x", actualToken)
			}
		})
	}
}

func TestRetrieveTokenSelf(t *testing.T) {
	lc := logger.MockLogger{}
	testServiceKey := "test-service-key"
	testTokenBaseDir := "test-base-dir"
	testTokenFileName := "test-token"

	// prepare test
	serviceDir := filepath.Join(testTokenBaseDir, testServiceKey)
	err := helper.CreateDirectoryIfNotExists(serviceDir)
	require.NoError(t, err)
	testToken := "token-x"
	// write a token into service directory
	tokenFilePath := filepath.Join(serviceDir, testTokenFileName)
	err = ioutil.WriteFile(tokenFilePath, []byte(testToken), 0600)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(testTokenBaseDir)
	}()

	tests := []struct {
		name         string
		token        string
		statusCode   int
		callerClient internal.HttpCaller
		expectedErr  bool
	}{
		{"Good:read self token resp ok", testToken, 200, &mockAuthHttpCaller{
			authTokenHeader: share.ConsulTokenHeader,
			authToken:       testToken,
			getStatusCode:   200,
			returnError:     false,
			getResponse:     `{"AccessorID":"xxxx", "SecretID":"` + testToken + `"}`,
		}, false},
		{"Good:read self token with empty token", share.EmptyToken, 404, &mockAuthHttpCaller{}, false},
		{"Good:read self token with token not found", testToken, 403, &mockAuthHttpCaller{
			authTokenHeader: share.ConsulTokenHeader,
			authToken:       testToken,
			getStatusCode:   403,
			returnError:     false,
			getResponse:     "ACL not found",
		}, false},
		{"Bad:read self token with internal server error", testToken, 500, &mockAuthHttpCaller{
			authTokenHeader: share.ConsulTokenHeader,
			authToken:       testToken,
			getStatusCode:   500,
			returnError:     false,
			getResponse:     "internal server error",
		}, true},
		{"Bad:read self token req not ok", testToken, 500, &mockAuthHttpCaller{
			returnError: true,
			getResponse: "EOF",
		}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCleaner, err := NewTokenCleaner("http://localhost:18500", testTokenBaseDir, testTokenFileName, lc, test.callerClient)
			require.NoError(t, err)
			require.NotNil(t, tokenCleaner)
			tokenInfo, err := tokenCleaner.retrieveSelfToken(tokenFilePath, test.token)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if test.token == share.EmptyToken {
					return
				}

				if test.statusCode == http.StatusForbidden {
					// means no self token found
					require.Nil(t, tokenInfo)
				} else {
					require.NotNil(t, tokenInfo)
				}
			}
		})
	}
}

func TestDeleteToken(t *testing.T) {
	lc := logger.MockLogger{}
	testServiceKey := "test-service-key"
	testTokenBaseDir := "test-base-dir"
	testTokenFileName := "test-token"
	bootstrapToken := "bootstrap-token"
	accessorID := "xxxxxx"

	// prepare test
	serviceDir := filepath.Join(testTokenBaseDir, testServiceKey)
	err := helper.CreateDirectoryIfNotExists(serviceDir)
	require.NoError(t, err)
	testToken := "token-x"
	// write a token into service directory
	tokenFilePath := filepath.Join(serviceDir, testTokenFileName)
	err = ioutil.WriteFile(tokenFilePath, []byte(testToken), 0600)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(testTokenBaseDir)
	}()

	tests := []struct {
		name           string
		bootstrapToken string
		accessorID     string
		callerClient   internal.HttpCaller
		expectedErr    bool
	}{
		{"Good:delete token resp ok", testToken, accessorID, &mockAuthHttpCaller{
			authTokenHeader:  share.ConsulTokenHeader,
			authToken:        bootstrapToken,
			deleteStatusCode: 200,
			returnError:      false,
			deleteResponse:   "true",
		}, false},
		{"Good:delete token accessorID is empty", testToken, "", &mockAuthHttpCaller{}, false},
		{"Bad:delete token resp status not ok", testToken, accessorID, &mockAuthHttpCaller{
			authTokenHeader:  share.ConsulTokenHeader,
			authToken:        testToken,
			deleteStatusCode: 403,
			returnError:      false,
			deleteResponse:   "permission denied",
		}, true},
		{"Bad:delete token resp status ok but returned not true", testToken, accessorID, &mockAuthHttpCaller{
			authTokenHeader:  share.ConsulTokenHeader,
			authToken:        testToken,
			deleteStatusCode: 200,
			returnError:      false,
			deleteResponse:   "false",
		}, true},
		{"Bad:delete token req not ok", testToken, accessorID, &mockAuthHttpCaller{
			returnError:    true,
			deleteResponse: "EOF",
		}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenCleaner, err := NewTokenCleaner("http://localhost:18500", testTokenBaseDir, testTokenFileName, lc, test.callerClient)
			require.NoError(t, err)
			require.NotNil(t, tokenCleaner)
			err = tokenCleaner.deleteToken(tokenFilePath, test.accessorID, test.bootstrapToken)

			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type mockAuthHttpCaller struct {
	authTokenHeader  string
	authToken        string
	getStatusCode    int
	returnError      bool
	getResponse      string
	deleteStatusCode int
	deleteResponse   string
}

func (smhc *mockAuthHttpCaller) Do(req *http.Request) (*http.Response, error) {
	switch req.Method {
	case http.MethodGet:
		if smhc.returnError {
			return &http.Response{
				StatusCode: smhc.getStatusCode,
			}, errors.New("http request Method Get error")
		}

		if req.Header.Get(smhc.authTokenHeader) != smhc.authToken {
			return nil, fmt.Errorf("auth header %s is expected but not present in request", smhc.authTokenHeader)
		}

		return &http.Response{
			StatusCode: smhc.getStatusCode,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(smhc.getResponse))),
		}, nil

	case http.MethodDelete:
		if smhc.returnError {
			return &http.Response{
				StatusCode: 500,
			}, errors.New("http request Method DELETE error")
		}

		return &http.Response{
			StatusCode: smhc.deleteStatusCode,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(smhc.deleteResponse))),
		}, nil

	default:
		return nil, errors.New("unsupported HTTP method")
	}

}
