//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

var methods = map[string]struct{}{
	http.MethodGet: {}, http.MethodHead: {}, http.MethodPost: {}, http.MethodPut: {}, http.MethodPatch: {},
	http.MethodDelete: {}, http.MethodTrace: {}, http.MethodConnect: {},
}

// SendRequestWithRESTAddress sends request with REST address
func SendRequestWithRESTAddress(lc logger.LoggingClient, content string, contentType string,
	address models.RESTAddress, jwtSecretProvider interfaces.AuthenticationInjector) (res string, err errors.EdgeX) {

	executingUrl := getUrlStr(address)

	req, err := getHttpRequest(address.HTTPMethod, executingUrl, content, contentType)
	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "fail to create http request", err)
	}

	if jwtSecretProvider != nil {
		if err2 := jwtSecretProvider.AddAuthenticationData(req); err2 != nil {
			return "", errors.NewCommonEdgeXWrapper(err2)
		}
	}

	client := &http.Client{}
	res, err = sendRequestAndGetResponse(client, req)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}
	lc.Debugf("success to send rest request with address %v", address.BaseAddress)
	return res, nil
}

func getUrlStr(address models.RESTAddress) string {
	return fmt.Sprintf("http://%s:%d%s", address.Host, address.Port, address.Path)
}

func validMethod(method string) bool {
	_, contains := methods[strings.ToUpper(method)]
	return contains
}

func getHttpRequest(
	httpMethod string,
	executingUrl string,
	content string, contentType string) (*http.Request, errors.EdgeX) {
	if !validMethod(httpMethod) {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("net/http: invalid method %q", httpMethod), nil)
	}

	var body []byte
	params := strings.TrimSpace(content)

	if len(params) > 0 {
		body = []byte(params)
	} else {
		body = nil
	}

	req, err := http.NewRequest(httpMethod, executingUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "create new request occurs error", err)
	}

	if contentType == "" {
		contentType = common.ContentTypeJSON
	}
	req.Header.Set(common.ContentType, contentType)

	if len(params) > 0 {
		req.Header.Set(common.ContentLength, strconv.Itoa(len(params)))
	}

	return req, nil
}

func sendRequestAndGetResponse(client *http.Client, req *http.Request) (res string, edgeXerr errors.EdgeX) {
	resp, err := client.Do(req)

	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "fail to send the HTTP request", err)
	}

	defer resp.Body.Close()
	resp.Close = true

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindIOError, "fail to read the response body", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", errors.NewCommonEdgeX(errors.KindMapping(resp.StatusCode), fmt.Sprintf("request failed, status code: %d, err: %s", resp.StatusCode, string(bodyBytes)), nil)
	}
	return string(bodyBytes), nil
}
