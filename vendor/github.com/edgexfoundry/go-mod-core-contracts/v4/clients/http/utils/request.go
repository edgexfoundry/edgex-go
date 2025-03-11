//
// Copyright (C) 2020-2024 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// GetRequest makes the get request and return the body
func GetRequest(ctx context.Context, returnValuePointer interface{}, baseUrl string, requestPath string, requestParams url.Values, authInjector interfaces.AuthenticationInjector) errors.EdgeX {
	req, err := CreateRequest(ctx, http.MethodGet, baseUrl, requestPath, requestParams)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// GetRequestAndReturnBinaryRes makes the get request and return the binary response and content type(i.e., application/json, application/cbor, ... )
func GetRequestAndReturnBinaryRes(ctx context.Context, baseUrl string, requestPath string, requestParams url.Values, authInjector interfaces.AuthenticationInjector) (res []byte, contentType string, edgeXerr errors.EdgeX) {
	req, edgeXerr := CreateRequest(ctx, http.MethodGet, baseUrl, requestPath, requestParams)
	if edgeXerr != nil {
		return nil, "", errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	resp, edgeXerr := makeRequest(req, authInjector)
	if edgeXerr != nil {
		return nil, "", errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	defer resp.Body.Close()

	res, edgeXerr = getBody(resp)
	if edgeXerr != nil {
		return nil, "", errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	if resp.StatusCode <= http.StatusMultiStatus {
		return res, resp.Header.Get(common.ContentType), nil
	}

	// Handle error response
	return nil,
		"",
		errors.NewCommonEdgeX(
			errors.KindMapping(resp.StatusCode),
			fmt.Sprintf("request failed, status code: %d, err: %s", resp.StatusCode, string(res)), nil)
}

// GetRequestWithBodyRawData makes the GET request with JSON raw data as request body and return the response
func GetRequestWithBodyRawData(ctx context.Context, returnValuePointer interface{}, baseUrl string, requestPath string, requestParams url.Values, data interface{}, authInjector interfaces.AuthenticationInjector) errors.EdgeX {
	req, err := CreateRequestWithRawDataAndParams(ctx, http.MethodGet, baseUrl, requestPath, requestParams, data)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// PostRequest makes the post request with encoded data and return the body
func PostRequest(
	ctx context.Context,
	returnValuePointer interface{},
	baseUrl string, requestPath string,
	data []byte,
	encoding string, authInjector interfaces.AuthenticationInjector) errors.EdgeX {

	req, err := CreateRequestWithEncodedData(ctx, http.MethodPost, baseUrl, requestPath, data, encoding)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// PostRequestWithRawData makes the post JSON request with raw data and return the body
func PostRequestWithRawData(
	ctx context.Context,
	returnValuePointer interface{},
	baseUrl string, requestPath string,
	requestParams url.Values,
	data interface{}, authInjector interfaces.AuthenticationInjector) errors.EdgeX {

	req, err := CreateRequestWithRawData(ctx, http.MethodPost, baseUrl, requestPath, requestParams, data)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// PostRequestWithRawDataAndHeaders makes the post JSON request with raw data and request headers, and returns the body
func PostRequestWithRawDataAndHeaders(
	ctx context.Context,
	returnValuePointer interface{},
	baseUrl string, requestPath string,
	requestParams url.Values,
	data interface{}, authInjector interfaces.AuthenticationInjector,
	headers map[string]string) errors.EdgeX {

	req, err := CreateRequestWithRawDataAndHeaders(ctx, http.MethodPost, baseUrl, requestPath, requestParams, data, headers)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// PutRequest makes the put JSON request and return the body
func PutRequest(
	ctx context.Context,
	returnValuePointer interface{},
	baseUrl string, requestPath string,
	requestParams url.Values,
	data interface{}, authInjector interfaces.AuthenticationInjector) errors.EdgeX {

	req, err := CreateRequestWithRawData(ctx, http.MethodPut, baseUrl, requestPath, requestParams, data)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// PatchRequest makes a PATCH request and unmarshals the response to the returnValuePointer
func PatchRequest(
	ctx context.Context,
	returnValuePointer interface{},
	baseUrl string, requestPath string,
	requestParams url.Values,
	data interface{}, authInjector interfaces.AuthenticationInjector) errors.EdgeX {

	req, err := CreateRequestWithRawData(ctx, http.MethodPatch, baseUrl, requestPath, requestParams, data)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// PostByFileRequest makes the post file request and return the body
func PostByFileRequest(
	ctx context.Context,
	returnValuePointer interface{},
	baseUrl string, requestPath string,
	filePath string, authInjector interfaces.AuthenticationInjector) errors.EdgeX {

	req, err := CreateRequestFromFilePath(ctx, http.MethodPost, baseUrl, requestPath, filePath)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// PutByFileRequest makes the put file request and return the body
func PutByFileRequest(
	ctx context.Context,
	returnValuePointer interface{},
	baseUrl string, requestPath string,
	filePath string, authInjector interfaces.AuthenticationInjector) errors.EdgeX {

	req, err := CreateRequestFromFilePath(ctx, http.MethodPut, baseUrl, requestPath, filePath)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// DeleteRequest makes the delete request and return the body
func DeleteRequest(ctx context.Context, returnValuePointer interface{}, baseUrl string, requestPath string, authInjector interfaces.AuthenticationInjector) errors.EdgeX {
	req, err := CreateRequest(ctx, http.MethodDelete, baseUrl, requestPath, nil)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// DeleteRequestWithParams makes the delete request with URL query params and return the body
func DeleteRequestWithParams(ctx context.Context, returnValuePointer interface{}, baseUrl string, requestPath string, requestParams url.Values, authInjector interfaces.AuthenticationInjector) errors.EdgeX {
	req, err := CreateRequest(ctx, http.MethodDelete, baseUrl, requestPath, requestParams)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return processRequest(ctx, returnValuePointer, req, authInjector)
}

// processRequest is a helper function to process the request and get the return value
func processRequest(ctx context.Context,
	returnValuePointer any, req *http.Request, authInjector interfaces.AuthenticationInjector) errors.EdgeX {
	resp, err := SendRequest(ctx, req, authInjector)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	// Check the response content length to avoid json unmarshal error
	if len(resp) == 0 {
		return nil
	}
	if err := json.Unmarshal(resp, returnValuePointer); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to parse the response body", err)
	}
	return nil
}
