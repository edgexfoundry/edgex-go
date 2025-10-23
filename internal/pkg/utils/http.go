//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/labstack/echo/v4"
)

func WriteHttpHeader(w http.ResponseWriter, ctx context.Context, statusCode int) {
	w.Header().Set(common.CorrelationHeader, correlation.FromContext(ctx))
	w.Header().Set(common.ContentType, common.ContentTypeJSON)
	// when the request destination  server is shut down or unreachable
	// the the statusCode in the response header  would be  zero .
	// http.ResponseWriter.WriteHeader will check statusCode,if less than 100 or bigger than 900,
	// when this check not pass would raise a panic, response to the caller can not be completed
	// to avoid panic see http.checkWriteHeaderCode
	if statusCode < 100 || statusCode > 900 {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(statusCode)
	}
}

// WriteErrorResponse writes Http header, encode error response with JSON format and writes to the HTTP response.
func WriteErrorResponse(w *echo.Response, ctx context.Context, lc logger.LoggingClient, err errors.EdgeX, requestId string) error {
	correlationId := correlation.FromContext(ctx)
	if errors.Kind(err) != errors.KindEntityDoesNotExist {
		lc.Error(err.Error(), common.CorrelationHeader, correlationId)
	}
	lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationId)
	errResponses := commonDTO.NewBaseResponse(requestId, err.Message(), err.Code())
	WriteHttpHeader(w, ctx, err.Code())
	return pkg.EncodeAndWriteResponse(errResponses, w, lc)
}

// ParseGetAllObjectsRequestQueryString parses offset, limit and labels from the query parameters. And use maximum and minimum to check whether the offset and limit are valid.
func ParseGetAllObjectsRequestQueryString(c echo.Context, minOffset int, maxOffset int, minLimit int, maxLimit int) (offset int, limit int, labels []string, err errors.EdgeX) {
	offset, err = ParseQueryStringToInt(c, common.Offset, common.DefaultOffset, minOffset, maxOffset)
	if err != nil {
		return offset, limit, labels, err
	}

	limit, err = ParseQueryStringToInt(c, common.Limit, common.DefaultLimit, minLimit, maxLimit)
	if err != nil {
		return offset, limit, labels, err
	}

	// Use maxLimit to specify the supported maximum size.
	if limit == -1 {
		limit = maxLimit
	}

	labels = ParseQueryStringToStrings(c, common.Labels, common.CommaSeparator)
	return offset, limit, labels, err
}

// Parse the specified query string key to an integer.  If specified query string key is found more than once in the
// http request, only the first specified query string will be parsed and converted to an integer.  If no specified
// query string key could be found in the http request, specified default value will be returned.  EdgeX error will be
// returned if any parsing error occurs.
func ParseQueryStringToInt(c echo.Context, queryStringKey string, defaultValue int, min int, max int) (int, errors.EdgeX) {
	// first check if specified min is bigger than max, throw error for such case
	if min > max {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("specified min %v is bigger than specified max %v", min, max), nil)
	}
	// defaultValue should not greater than maximum
	if defaultValue > max {
		defaultValue = max
	}
	var result = defaultValue
	var parsingErr error
	value := c.QueryParam(queryStringKey)
	if value != "" {
		result, parsingErr = strconv.Atoi(strings.TrimSpace(value))
		if parsingErr != nil {
			return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("failed to parse querystring %s's value %s into integer. Error:%s", queryStringKey, value, parsingErr.Error()), nil)
		}
		if result < min || result > max {
			return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("querystring %s's value %v is out of min %v ~ max %v range.", queryStringKey, result, min, max), nil)
		}
	}
	return result, nil
}

// ParseQueryStringToInt64 parses the specified query string key to a 64-bit integer.  If specified query string key is found more than once in the
// http request, only the first specified query string will be parsed and converted to an integer.  If no specified
// query string key could be found in the http request, specified default value will be returned.  EdgeX error will be
// returned if any parsing error occurs.
func ParseQueryStringToInt64(c echo.Context, queryStringKey string, defaultValue int64, min int64, max int64) (int64, errors.EdgeX) {
	// first check if specified min is bigger than max, throw error for such case
	if min > max {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("specified min %v is bigger than specified max %v", min, max), nil)
	}
	// defaultValue should not greater than maximum
	if defaultValue > max {
		defaultValue = max
	}
	var result = defaultValue
	var parsingErr error
	value := c.QueryParam(queryStringKey)
	if value != "" {
		result, parsingErr = strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if parsingErr != nil {
			return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("failed to parse querystring %s's value %s into integer. Error:%s", queryStringKey, value, parsingErr.Error()), nil)
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
func ParseQueryStringToStrings(c echo.Context, queryStringKey string, separator string) (stringArray []string) {
	if len(separator) == 0 {
		separator = common.CommaSeparator
	}
	value := c.QueryParam(queryStringKey)
	if value != "" {
		stringArray = strings.Split(strings.TrimSpace(value), separator)
	}
	return stringArray
}

// Parse the specified query string key to a string.  If specified query string key is found more than once in
// the http request, only the first specified query string will be parsed and converted to a string.  If no specified
// query string could be found, defaultValue will be returned.
func ParseQueryStringToString(r *http.Request, queryStringKey string, defaultValue string) string {
	value, ok := r.URL.Query()[queryStringKey]
	if !ok {
		return defaultValue
	}
	return value[0]
}

func ParseTimeRangeOffsetLimit(c echo.Context, minOffset int, maxOffset int, minLimit int, maxLimit int) (start int64, end int64, offset int, limit int, edgexErr errors.EdgeX) {
	start, edgexErr = ParsePathParamToInt64(c, common.Start, 0, math.MaxInt64)
	if edgexErr != nil {
		return start, end, offset, limit, edgexErr
	}
	end, edgexErr = ParsePathParamToInt64(c, common.End, 0, math.MaxInt64)
	if edgexErr != nil {
		return start, end, offset, limit, edgexErr
	}
	if end < start {
		return start, end, offset, limit, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("end's value %v is not allowed to be greater than start's value %v", end, start), nil)
	}
	offset, limit, _, edgexErr = ParseGetAllObjectsRequestQueryString(c, minOffset, maxOffset, minLimit, maxLimit)
	if edgexErr != nil {
		return start, end, offset, limit, edgexErr
	}

	return start, end, offset, limit, nil
}

func ParseQueryStringTimeRangeOffsetLimit(c echo.Context, minOffset int, maxOffset int, minLimit int, maxLimit int) (start int64, end int64, offset int, limit int, edgexErr errors.EdgeX) {
	start, edgexErr = ParseQueryStringToInt64(c, common.Start, 0, 0, math.MaxInt64)
	if edgexErr != nil {
		return start, end, offset, limit, edgexErr
	}
	end, edgexErr = ParseQueryStringToInt64(c, common.End, time.Now().UnixMilli(), 0, math.MaxInt64)
	if edgexErr != nil {
		return start, end, offset, limit, edgexErr
	}
	if end < start {
		return start, end, offset, limit, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("end's value %v is not allowed to be greater than start's value %v", end, start), nil)
	}
	offset, limit, _, edgexErr = ParseGetAllObjectsRequestQueryString(c, minOffset, maxOffset, minLimit, maxLimit)
	if edgexErr != nil {
		return start, end, offset, limit, edgexErr
	}

	return start, end, offset, limit, nil
}

// ParsePathParamToInt64 parses the specified path parameter to a 64-bit integer.  EdgeX error will be returned if any parsing error occurs or
// specified path parameter is empty.
func ParsePathParamToInt64(c echo.Context, pathKey string, min, max int64) (int64, errors.EdgeX) {
	val := c.Param(pathKey)
	if val == "" {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("empty path param %s is not allowed", pathKey), nil)
	}

	// check if specified min is bigger than max, throw error for such case
	if min > max {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("specified min %v is bigger than specified max %v", min, max), nil)
	}

	result, parsingErr := strconv.ParseInt(val, 10, 64)
	if parsingErr != nil {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("failed to parse path param %s's value %s into integer. Error:%s", pathKey, val, parsingErr.Error()), nil)
	}

	if result < min || result > max {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("path param %s's value %v is out of min %v ~ max %v range.", pathKey, result, min, max), nil)
	}
	return result, nil
}

// ParseBodyToMap parses the body of http request to a map[string]interfaces{}.  EdgeX error will be returned if any parsing error occurs.
func ParseBodyToMap(r *http.Request) (map[string]interface{}, errors.EdgeX) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// cannot use logging client here, just print to console
			fmt.Printf("error occured while closing the request body: %s", err.Error())
		}
	}(r.Body)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to read request body", err)
	}

	var result map[string]interface{}
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to parse request body", err)
	}

	return result, nil
}

// ParseAggregateFuncQueryString checks whether the aggregateFunc query param is valid, and returns the value in all capital letters
func ParseAggregateFuncQueryString(aggFuncParam string) (string, errors.EdgeX) {
	aggFunc := strings.ToUpper(aggFuncParam)
	if slices.Contains([]string{common.MinFunc, common.MaxFunc, common.AvgFunc, common.CountFunc, common.SumFunc}, aggFunc) {
		return aggFunc, nil
	}
	return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid aggregate function specified", nil)
}
