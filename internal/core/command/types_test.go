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

package command

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// FailingMockHttpCaller a HttpCaller which always returns an error when executing a request.
type FailingMockHttpCaller struct{}

// Do always returns an empty http.Response and an error.
func (FailingMockHttpCaller) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{}, errors.New("testing error handling")
}

// ReadFailMockHttpCaller an HttpCaller which returns a http.Response that will always fail when attempting to read the Body.
type ReadFailMockHttpCaller struct{}

// Do returns a http.Response which contains a Body that will return an error when attempting to read.
func (ReadFailMockHttpCaller) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Body: MockBody{},
	}, nil
}

// MockBody a Body which will return 0 and an error when attempting to read.
type MockBody struct{}

// Read returns 0 and an error
func (MockBody) Read(p []byte) (n int, err error) {
	return 0, errors.New("testing read error")
}

// Close implementation not required
func (MockBody) Close() error {
	panic("implement me")
}

func TestExecute(t *testing.T) {
	expectedResponseBody := "Sample Response Body"
	expectedResponseCode := http.StatusOK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expectedResponseBody)
		w.WriteHeader(expectedResponseCode)
	}))
	defer ts.Close()

	request, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	sc := newServiceCommand(contract.Device{AdminState: contract.Unlocked}, &http.Client{}, request, logger.NewMockClient())

	body, responseCode, err := sc.Execute()
	if err != nil {
		t.Errorf("No error should be present for happy path")
	}

	if body != expectedResponseBody {
		t.Errorf("The response body was not properly propegated to the caller")
	}

	if responseCode != expectedResponseCode {
		t.Errorf("The response code was not properly propegated to the caller")
	}
}

func TestExecuteHttpRequestError(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	sc := newServiceCommand(contract.Device{AdminState: contract.Unlocked}, &FailingMockHttpCaller{}, request, logger.NewMockClient())

	_, _, err := sc.Execute()
	if err == nil {
		t.Errorf("No error should be present for happy path")
	}
}

func TestExecuteHttpReadResponseError(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	sc := newServiceCommand(contract.Device{AdminState: contract.Unlocked}, &ReadFailMockHttpCaller{}, request, logger.NewMockClient())

	_, responseCode, err := sc.Execute()
	if err == nil {
		t.Errorf("The error was not properly propegated to the caller")
	}

	if responseCode != DefaultErrorCode {
		t.Errorf("The response code should be the default value for failing requests")
	}
}
