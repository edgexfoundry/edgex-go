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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
)

type badReader struct{}

func (b badReader) Read(p []byte) (int, error) {
	return 0, errors.New("Error")
}

func TestDoRequestBadReader(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	realvc := (vc).(*vaultClient)

	// Act
	code, err := realvc.doRequest(commonRequestArgs{
		AuthToken:            "",
		Method:               "method",
		Path:                 "somepath",
		JSONObject:           nil,
		BodyReader:           badReader{},
		OperationDescription: "opname",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       nil,
	})

	// Assert
	assert.Error(err)
	assert.Equal(0, code)
}

func TestDoRequestUnexpectedStatus(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	realvc := (vc).(*vaultClient)

	// Act
	code, err := realvc.doRequest(commonRequestArgs{
		AuthToken:            "",
		Method:               "method",
		Path:                 "somepath",
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "opname",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       nil,
	})

	// Assert
	assert.Error(err)
	assert.Equal(http.StatusBadRequest, code)
}

func TestDoRequestBadJSONObject(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("badly{formatted"))
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	realvc := (vc).(*vaultClient)

	// Act
	code, err := realvc.doRequest(commonRequestArgs{
		AuthToken:            "",
		Method:               "method",
		Path:                 "somepath",
		JSONObject:           func() {},
		BodyReader:           nil,
		OperationDescription: "opname",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       nil,
	})

	// Assert
	assert.Error(err)
	assert.Equal(0, code)
}

func TestDoRequestBadBody(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("badly{formatted"))
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)
	realvc := (vc).(*vaultClient)

	var responseObject interface{}

	// Act
	code, err := realvc.doRequest(commonRequestArgs{
		AuthToken:            "",
		Method:               "method",
		Path:                 "somepath",
		JSONObject:           nil,
		BodyReader:           nil,
		OperationDescription: "opname",
		ExpectedStatusCode:   http.StatusOK,
		ResponseObject:       &responseObject,
	})

	// Assert
	assert.Error(err)
	assert.Equal(http.StatusOK, code)
}
