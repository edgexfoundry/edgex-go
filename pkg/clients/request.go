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

package clients

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
)

const (
	ContentType = "Content-Type"
	ContentJson = "application/json"
	ContentYaml = "application/x-yaml"
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
func GetRequest(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, types.ErrResponseNil{}
	}
	defer resp.Body.Close()

	bodyBytes, err := getBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, types.ErrNotFound{}
	} else if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusAccepted) {
		return nil, types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return bodyBytes, nil
}

// Helper method to make the count request
func CountRequest(url string) (int, error) {
	data, err := GetRequest(url)
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Helper method to make the post JSON request and return the body
func PostJsonRequest(url string, data interface{}) (string, error) {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return PostRequest(url, jsonStr, ContentJson)
}

// Helper method to make the post request and return the body
func PostRequest(url string, data []byte, content string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set(ContentType, content)

	resp, err := makeRequest(req)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", types.ErrResponseNil{}
	}
	defer resp.Body.Close()

	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusAccepted) {
		return "", types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	bodyString := string(bodyBytes)
	return bodyString, nil
}

// Helper method to make a post request in order to upload a file and return the request body
func UploadFileRequest(url string, filePath string) (string, error) {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Create multipart/form-data request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	formFileWriter, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	_, err = io.Copy(formFileWriter, bytes.NewReader(fileContents))
	if err != nil {
		return "", err
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", err
	}
	req.Header.Add(ContentType, writer.FormDataContentType())

	resp, err := makeRequest(req)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", types.ErrResponseNil{}
	}
	defer resp.Body.Close()

	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusAccepted) {
		return "", types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	bodyString := string(bodyBytes)
	return bodyString, nil
}

// Helper method to make the update request
func UpdateRequest(url string, data interface{}) error {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = PutRequest(url, jsonStr)
	return err
}

// Helper method to make the put request
func PutRequest(url string, body []byte) (string, error) {
	var err error
	var req *http.Request

	if body != nil {
		req, err = http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(http.MethodPut, url, nil)
	}
	if err != nil {
		return "", err
	}

	if body != nil {
		req.Header.Set(ContentType, ContentJson)
	}

	resp, err := makeRequest(req)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", types.ErrResponseNil{}
	}
	defer resp.Body.Close()

	bodyBytes, err := getBody(resp)
	if err != nil {
		return "", err
	}

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusAccepted) {
		return "", types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	bodyString := string(bodyBytes)
	return bodyString, nil
}

// Helper method to make the delete request
func DeleteRequest(url string) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := makeRequest(req)
	if err != nil {
		return err
	}
	if resp == nil {
		return types.ErrResponseNil{}
	}
	defer resp.Body.Close()

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusAccepted) {
		bodyBytes, err := getBody(resp)
		if err != nil {
			return err
		}

		return types.NewErrServiceClient(resp.StatusCode, bodyBytes)
	}

	return nil
}
