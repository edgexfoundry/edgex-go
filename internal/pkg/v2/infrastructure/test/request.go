/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/correlationid"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// InvalidJSON returns content that is not valid JSON.
func InvalidJSON() []byte {
	return []byte("This Is Not Valid JSON")
}

// Marshal is a factory method that returns json-marshalled content.
func Marshal(t *testing.T, content interface{}) []byte {
	body, e := json.Marshal(content)
	if e != nil {
		assert.FailNow(t, "Unexpected marshal failure:", e.Error())
	}
	return body
}

// SendRequest is common implementation to create recorder, send a request, and return recorder for evaluation.
func SendRequest(
	t *testing.T,
	router *mux.Router,
	method string,
	url string,
	correlationID string,
	body []byte) (*httptest.ResponseRecorder, string) {

	w := httptest.NewRecorder()

	r, e := http.NewRequest(method, url, bytes.NewReader(body))
	if e != nil {
		assert.FailNow(t, "Unexpected http.NewRequest failure:", e.Error())
		return nil, ""
	}

	r.Header.Set(correlationid.HTTPHeader, correlationID)

	router.ServeHTTP(w, r)
	return w, correlationID
}

// SendRequestWithBody is common implementation to create recorder, send a request, and return recorder for evaluation.
func SendRequestWithBody(
	t *testing.T,
	router *mux.Router,
	method string,
	url string,
	body []byte) (*httptest.ResponseRecorder, string) {

	return SendRequest(t, router, method, url, FactoryRandomString(), body)
}

// SendRequestWithoutBody is common implementation to create recorder, send a request that has no body, and return
// recorder for evaluation.
func SendRequestWithoutBody(t *testing.T, router *mux.Router, method, url string) (*httptest.ResponseRecorder, string) {
	return SendRequest(t, router, method, url, FactoryRandomString(), []byte{})
}
