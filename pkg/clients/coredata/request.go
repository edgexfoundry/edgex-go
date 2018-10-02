/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright 2018 Joan Duran
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

package coredata

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

var (
	// ErrResponseNil is the error in case of empty response
	ErrResponseNil = errors.New("Response was nil")
)

// Helper method to get the body from the response after making the request
func getBody(resp *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// Helper method to make the request and return the response
func makeRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)

	return resp, err
}

// Helper method to make the get request and return the body
func getRequest(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, ErrResponseNil
	}
	defer resp.Body.Close()

	bodyBytes, err := getBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return bodyBytes, nil
}

// Helper method to make the count request
func countRequest(url string) (int, error) {
	data, err := getRequest(url)
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Helper method to make the post request and return the body
func postRequest(url string, data interface{}) (string, error) {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonStr))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", ErrResponseNil
	}
	defer resp.Body.Close()

	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	bodyString := string(bodyBytes)
	return bodyString, nil
}

// Helper method to make the put request
func putRequest(url string, data interface{}) error {
	var err error
	var req *http.Request

	if data != nil {
		var data []byte
		data, err = json.Marshal(data)
		if err != nil {
			return err
		}

		req, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	} else {
		req, err = http.NewRequest(http.MethodPut, url, nil)
	}

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}

		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}

// Helper method to make the delete request
func deleteRequest(url string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrResponseNil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}

		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}
