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

package secretstoreclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestRegenRootToken(t *testing.T) {
	// Arrange
	assert := assert.New(t)
	mockLogger := logger.MockLogger{}

	requestNumber := 0

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestNumber++
		theMethod := r.Method
		thePath := r.URL.EscapedPath()
		switch requestNumber {
		case 1:
			assert.Equal("DELETE", theMethod)
			assert.Equal(RootTokenControlAPI, thePath)
			w.WriteHeader(http.StatusNoContent)
		case 2:
			assert.Equal("PUT", theMethod)
			assert.Equal(RootTokenControlAPI, thePath)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(RootTokenControlResponse{
				Complete: false,
				Otp:      "jzEHVfxe6w0Q0yz5jQuvlQG557",
				Nonce:    "2dbd10f1-8528-6246-09e7-82b25b8aba63",
			})
		case 3:
			assert.Equal("PUT", theMethod)
			assert.Equal(RootTokenRetrievalAPI, thePath)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(RootTokenRetrievalResponse{
				Complete:     true,
				EncodedToken: "GVQfeQ5eIQ5+IlczQy0JBw80ITI6FHFme3w",
			})
		}
	}))
	defer ts.Close()

	host := strings.Replace(ts.URL, "https://", "", -1)
	vc := NewSecretStoreClient(mockLogger, NewRequestor(mockLogger).Insecure(), "https", host)

	// Act
	initResp := InitResponse{
		// Keys is unused
		KeysBase64: []string{"dGVzdC1rZXktMQ==", "dGVzdC1rZXktMgo="},
	}
	var rootToken string
	err := vc.RegenRootToken(&initResp, &rootToken)

	// Assert
	assert.Nil(err)
	assert.Equal("s.Z1X8YkHUgbsTs2eeTDVE6SNK", string(rootToken))
}
