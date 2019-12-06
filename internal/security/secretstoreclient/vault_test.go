/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

 *******************************************************************************/

package secretstoreclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/fileioperformer"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != VaultHealthAPI {
			t.Errorf("expected request to /%s, got %s instead", VaultHealthAPI, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	code, err := vc.HealthCheck()

	assert.NoError(t, err)
	if code != http.StatusOK {
		t.Errorf("incorrect vault health check status.")
	}
}

// TestHealthCheckUninit tests that a no error object is returned
// for uninitialized vault (with 5xx error)
func TestHealthCheckUninit(t *testing.T) {
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	code, err := vc.HealthCheck()

	assert.NoError(t, err)
	if code != http.StatusNotImplemented {
		t.Errorf("incorrect vault health check status.")
	}
}

// TestHealthCheckSealed tests that a no error object is returned
// for sealed vault (with 5xx error)
func TestHealthCheckSealed(t *testing.T) {
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	code, err := vc.HealthCheck()

	assert.NoError(t, err)
	if code != http.StatusServiceUnavailable {
		t.Errorf("incorrect vault health check status.")
	}
}

func TestInit(t *testing.T) {
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"keys": [
			  "test-keys"
			],
			"keys_base64": [
			  "test-keys-base64"
			],
			"root_token": "test-root-token"
		}
		`))
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != VaultInitAPI {
			t.Errorf("expected request to /%s, got %s instead", VaultInitAPI, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()

	config := SecretServiceInfo{
		TokenFolderPath: "testdata",
		TokenFile:       "test-resp-init.json",
	}

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	tokenFile, err := os.OpenFile(filepath.Join(config.TokenFolderPath, config.TokenFile), os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		t.Errorf("failed to open token file %s", err.Error())
	}
	defer tokenFile.Close()
	code, _ := vc.Init(config, tokenFile)
	if code != http.StatusOK {
		t.Errorf("incorrect vault init status. The returned code is %d", code)
	}
}

func TestUnseal(t *testing.T) {
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sealed": false, "t": 1, "n": 1, "progress": 100}`))
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s instead", r.Method)
		}

		if r.URL.EscapedPath() != VaultUnsealAPI {
			t.Errorf("expected request to /%s, got %s instead", VaultUnsealAPI, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()

	config := SecretServiceInfo{
		TokenFolderPath:   "testdata",
		TokenFile:         "test-resp-init.json",
		VaultSecretShares: 1,
	}

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	tokenFile, err := fileioperformer.NewDefaultFileIoPerformer().OpenFileReader(filepath.Join(config.TokenFolderPath, config.TokenFile), os.O_RDONLY, 0400)
	if err != nil {
		t.Errorf("failed to open token file %s", err.Error())
	}
	code, err := vc.Unseal(config, tokenFile)
	if code != http.StatusOK {
		t.Errorf("incorrect vault unseal status. The returned code is %d, %s", code, err.Error())
	}
}

func TestInstallPolicy(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("PUT", r.Method)
		assert.Equal("/v1/sys/policies/acl/policy-name", r.URL.EscapedPath())
		assert.Equal("fake-token", r.Header.Get("X-Vault-Token"))

		// Make sure the policy doc was base64 encoded in the json response object
		body := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.Nil(err)
		assert.Equal("policydoc", body["policy"].(string))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)

	// Act
	policyDoc := "policydoc"
	code, err := vc.InstallPolicy("fake-token", "policy-name", policyDoc)

	// Assert
	assert.Nil(err)
	assert.Equal(http.StatusNoContent, code)
}

func TestCreateToken(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method)
		assert.Equal(CreateTokenAPI, r.URL.EscapedPath())
		assert.Equal("fake-token", r.Header.Get("X-Vault-Token"))

		body := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.Nil(err)

		assert.Equal("sample-value", body["sample_parameter"])

		w.WriteHeader(http.StatusOK)

		response := struct {
			RequestID string `json:"request_id"`
		}{
			RequestID: "f00341c1-fad5-f6e6-13fd-235617f858a1",
		}
		err = json.NewEncoder(w).Encode(response)
		assert.Nil(err)

	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)

	// Act
	parameters := make(map[string]interface{})
	parameters["sample_parameter"] = "sample-value"
	response := make(map[string]interface{})
	code, err := vc.CreateToken("fake-token", parameters, &response)

	// Assert
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal("f00341c1-fad5-f6e6-13fd-235617f858a1", response["request_id"].(string))
}

func TestListAccessors(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("LIST", r.Method)
		assert.Equal(ListAccessorsAPI, r.URL.EscapedPath())
		assert.Equal("fake-token", r.Header.Get("X-Vault-Token"))

		// No body for this request

		w.WriteHeader(http.StatusOK)

		response := struct {
			Data interface{} `json:"data"`
		}{
			Data: struct {
				Keys []string `json:"keys"`
			}{
				Keys: []string{"accessor1", "accessor2"},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		assert.Nil(err)

	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)

	// Act
	var response []string
	code, err := vc.ListAccessors("fake-token", &response)

	// Assert
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal("accessor1", response[0])
	assert.Equal("accessor2", response[1])
}

func TestRevokeAccessor(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method)
		assert.Equal(RevokeAccessorAPI, r.URL.EscapedPath())
		assert.Equal("fake-token", r.Header.Get("X-Vault-Token"))

		body := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.Nil(err)

		assert.Equal("accessor1", body["accessor"])

		w.WriteHeader(http.StatusNoContent)

		// no response body
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)

	// Act
	code, err := vc.RevokeAccessor("fake-token", "accessor1")

	// Assert
	assert.Nil(err)
	assert.Equal(http.StatusNoContent, code)
}

func TestLookupAccessor(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method)
		assert.Equal(LookupAccessorAPI, r.URL.EscapedPath())
		assert.Equal("fake-token", r.Header.Get("X-Vault-Token"))

		body := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.Nil(err)

		assert.Equal("8609694a-cdbc-db9b-d345-e782dbb562ed", body["accessor"])

		w.WriteHeader(http.StatusOK)

		response := struct {
			Data interface{} `json:"data"`
		}{
			Data: struct {
				Accessor string `json:"accessor"`
			}{
				Accessor: "accessor-value",
			},
		}
		err = json.NewEncoder(w).Encode(response)
		assert.Nil(err)

	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)

	// Act
	var md TokenMetadata
	code, err := vc.LookupAccessor("fake-token", "8609694a-cdbc-db9b-d345-e782dbb562ed", &md)

	// Assert
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal("accessor-value", md.Accessor)
}

func TestLookupSelf(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("GET", r.Method)
		assert.Equal(LookupSelfAPI, r.URL.EscapedPath())
		assert.Equal("fake-token", r.Header.Get("X-Vault-Token"))

		// No body for this request

		w.WriteHeader(http.StatusOK)

		response := struct {
			Data interface{} `json:"data"`
		}{
			Data: struct {
				Accessor string `json:"accessor"`
			}{
				Accessor: "accessor-value",
			},
		}
		err := json.NewEncoder(w).Encode(response)
		assert.Nil(err)

	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)

	// Act
	var md TokenMetadata
	code, err := vc.LookupSelf("fake-token", &md)

	// Assert
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal("accessor-value", md.Accessor)
}

func TestRevokeSelf(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal("POST", r.Method)
		assert.Equal(RevokeSelfAPI, r.URL.EscapedPath())
		assert.Equal("fake-token", r.Header.Get("X-Vault-Token"))

		// No body, no response body for this request

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)

	// Act
	code, err := vc.RevokeSelf("fake-token")

	// Assert
	assert.Nil(err)
	assert.Equal(http.StatusNoContent, code)
}
