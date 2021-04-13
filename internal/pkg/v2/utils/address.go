//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

const (
	ContentTypeJsonValue = "application/json; charset=utf-8"
	ContentLengthKey     = "Content-Length"
)

// SendRequestWithAddress sends request or message according to Address type(REST, MQTT, EMAIL, ...)
func SendRequestWithAddress(lc logger.LoggingClient, address models.Address) error {
	switch address.GetBaseAddress().Type {
	case v2.REST:
		restAddress := address.(models.RESTAddress)
		err := sendRESTRequest(lc, restAddress)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	default:
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Unsupported address type", nil)
	}
	return nil
}

func sendRESTRequest(lc logger.LoggingClient, address models.RESTAddress) errors.EdgeX {
	executingUrl := getUrlStr(address)

	req, err := getHttpRequest(address.HTTPMethod, executingUrl, address)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "fail to create http request", err)
	}

	client := &http.Client{}
	responseBytes, err := sendRequestAndGetResponse(client, req)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "fail to send http request", err)
	}
	responseStr := string(responseBytes)
	lc.Debugf("execution returns response content : %s", responseStr)
	return nil
}

func getUrlStr(address models.RESTAddress) string {
	return fmt.Sprintf("http://%s:%d%s", address.Host, address.Port, address.Path)
}

func validMethod(method string) bool {
	var methods = map[string]struct{}{
		http.MethodGet: {}, http.MethodHead: {}, http.MethodPost: {}, http.MethodPut: {}, http.MethodPatch: {},
		http.MethodDelete: {}, http.MethodTrace: {}, http.MethodConnect: {},
	}
	_, contains := methods[strings.ToUpper(method)]
	return contains
}

func getHttpRequest(
	httpMethod string,
	executingUrl string,
	address models.RESTAddress) (*http.Request, errors.EdgeX) {
	if !validMethod(httpMethod) {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("net/http: invalid method %q", httpMethod), nil)
	}

	var body []byte
	params := strings.TrimSpace(address.RequestBody)

	if len(params) > 0 {
		body = []byte(params)
	} else {
		body = nil
	}

	req, err := http.NewRequest(httpMethod, executingUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "create new request occurs error", err)
	}

	req.Header.Set(clients.ContentType, ContentTypeJsonValue)

	if len(params) > 0 {
		req.Header.Set(ContentLengthKey, strconv.Itoa(len(params)))
	}

	return req, nil
}

func sendRequestAndGetResponse(client *http.Client, req *http.Request) ([]byte, errors.EdgeX) {
	resp, err := client.Do(req)

	if err != nil {
		return []byte{}, errors.NewCommonEdgeX(errors.KindServerError, "fail to send the HTTP request", err)
	}

	defer resp.Body.Close()
	resp.Close = true

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, errors.NewCommonEdgeX(errors.KindServerError, "fail to read the response body", err)
	}

	return bodyBytes, nil
}
