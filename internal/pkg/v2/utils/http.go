//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
)

func WriteHttpHeader(w http.ResponseWriter, ctx context.Context, statusCode int) {
	w.Header().Set(clients.CorrelationHeader, correlation.FromContext(ctx))
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(statusCode)
}

func ParseGetAllObjectsRequestQueryString(r *http.Request, minOffset int, maxOffset int, minLimit int, maxLimit int) (offset int, limit int, labels []string, err errors.EdgeX) {
	offset, err = ParseQueryStringToInt(r, contractsV2.Offset, contractsV2.DefaultOffset, minOffset, maxOffset)
	if err != nil {
		return offset, limit, labels, err
	}

	limit, err = ParseQueryStringToInt(r, contractsV2.Limit, contractsV2.DefaultLimit, minLimit, maxLimit)
	if err != nil {
		return offset, limit, labels, err
	}

	labels = ParseQueryStringToStrings(r, contractsV2.Labels, contractsV2.CommaSeparator)
	return offset, limit, labels, err
}

// Parse the specified query string key to an integer.  If specified query string key is found more than once in the
// http request, only the first specified query string will be parsed and converted to an integer.  If no specified
// query string key could be found in the http request, specified default value will be returned.  EdgeX error will be
// returned if any parsing error occurs.
func ParseQueryStringToInt(r *http.Request, queryStringKey string, defaultValue int, min int, max int) (int, errors.EdgeX) {
	// first check if specified min is bigger than max, throw error for such case
	if min > max {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("specified min %v is bigger than specified max %v", min, max), nil)
	}
	var result = defaultValue
	var parsingErr error
	values, ok := r.URL.Query()[queryStringKey]
	if ok && len(values) > 0 {
		result, parsingErr = strconv.Atoi(strings.TrimSpace(values[0]))
		if parsingErr != nil {
			return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("failed to parse querystring %s's value %s into integer. Error:%s", queryStringKey, values[0], parsingErr.Error()), nil)
		}
		if result < min || result > max {
			return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("querystring %s's value %v is out of min %v ~ max %v range.", queryStringKey, result, min, max), nil)
		}
	}
	return result, nil
}

// Parse the specified query string key to an array of string.  If specified query string key is found more than once in
// the http request, only the first specified query string will be parsed and converted to an array of string.  The
// value of query string will be split into an array of string by the passing separator.  If separator is passed in as
// an empty string, comma separator will be used.
func ParseQueryStringToStrings(r *http.Request, queryStringKey string, separator string) (stringArray []string) {
	if len(separator) == 0 {
		separator = contractsV2.CommaSeparator
	}
	values, ok := r.URL.Query()[queryStringKey]
	if ok && len(values) >= 1 {
		stringArray = strings.Split(strings.TrimSpace(values[0]), separator)
	}
	return stringArray
}
