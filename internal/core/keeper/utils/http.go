//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/constants"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// ParseGetKeyRequestQueryString parses keyOnly and plaintext from the query parameters.
func ParseGetKeyRequestQueryString(r *http.Request) (keysOnly bool, isRaw bool, err errors.EdgeX) {
	keysOnly, err = ParseQueryStringToBool(r, constants.KeyOnly)
	if err != nil {
		return keysOnly, isRaw, errors.NewCommonEdgeXWrapper(err)
	}
	isRaw, err = ParseQueryStringToBool(r, constants.Plaintext)
	if err != nil {
		return keysOnly, isRaw, errors.NewCommonEdgeXWrapper(err)
	}
	return keysOnly, isRaw, nil
}

// ParseAddKeyRequestQueryString parses flatten from the query parameters.
func ParseAddKeyRequestQueryString(r *http.Request) (isFlatten bool, err errors.EdgeX) {
	isFlatten, err = ParseQueryStringToBool(r, constants.Flatten)
	if err != nil {
		return isFlatten, errors.NewCommonEdgeXWrapper(err)
	}
	return isFlatten, nil
}

// ParseDeleteKeyRequestQueryString parses prefixMatch from the query parameters.
func ParseDeleteKeyRequestQueryString(r *http.Request) (prefixMatch bool, err errors.EdgeX) {
	prefixMatch, err = ParseQueryStringToBool(r, constants.PrefixMatch)
	if err != nil {
		return prefixMatch, errors.NewCommonEdgeXWrapper(err)
	}
	return prefixMatch, nil
}

// ParseQueryStringToBool parses the specified query string key to a bool.  If specified query string key is found more than once in the
// http request, only the first specified query string will be parsed and converted to a bool.  If no specified
// query string key could be found in the http request, specified default value will be returned.  EdgeX error will be
// returned if any parsing error occurs.
func ParseQueryStringToBool(r *http.Request, queryStringKey string) (bool, errors.EdgeX) {
	var result bool
	var parsingErr error
	param := r.URL.Query().Get(queryStringKey)

	if param != "" {
		result, parsingErr = strconv.ParseBool(strings.TrimSpace(param))
		if parsingErr != nil {
			return false, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("failed to parse querystring %s into bool. Error:%s", queryStringKey, parsingErr.Error()), nil)
		}
	}
	return result, nil
}
