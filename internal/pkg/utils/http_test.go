//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGetAllObjectsRequestQueryString(t *testing.T) {
	testLabel := "testLabel"
	tests := []struct {
		name              string
		offset            string
		limit             string
		labels            string
		maxLimit          int
		expectedOffset    int
		expectedLimit     int
		expectedLabels    []string
		expectedErrorKind errors.ErrKind
	}{
		{"valid", "0", "2", testLabel, 10, 0, 2, []string{testLabel}, ""},
		{"valid - labels is empty", "0", "2", "", 10, 0, 2, nil, ""},
		{"valid - offset and limit is empty", "", "", testLabel, 50, common.DefaultOffset, common.DefaultLimit, []string{testLabel}, ""},
		{"valid - limit is empty and default value is greater then the maximum", "", "", testLabel, 5, common.DefaultOffset, 5, []string{testLabel}, ""},
		{"invalid - offset exceeds the minimum ", "-1", "2", testLabel, 50, 0, 2, []string{testLabel}, errors.KindContractInvalid},
		{"invalid - offset exceeds the maximum ", strconv.Itoa(math.MaxUint32), "2", testLabel, 10, 0, 2, []string{testLabel}, errors.KindContractInvalid},
		{"invalid - limit exceeds the minimum ", "0", "-2", testLabel, 50, 0, 2, []string{testLabel}, errors.KindContractInvalid},
		{"invalid - limit exceeds the maximum ", "0", "100", testLabel, 50, 0, 2, []string{testLabel}, errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			e := echo.New()
			req, err := http.NewRequest(http.MethodGet, common.ApiAllEventRoute, http.NoBody)
			require.NoError(t, err)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(common.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(common.Limit, testCase.limit)
			}
			if testCase.labels != "" {
				query.Add(common.Labels, testCase.labels)
			}
			req.URL.RawQuery = query.Encode()

			c := e.NewContext(req, nil)
			offset, limit, labels, err := ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, testCase.maxLimit)
			if testCase.expectedErrorKind != "" {
				assert.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedOffset, offset)
			assert.Equal(t, testCase.expectedLimit, limit)
			assert.Equal(t, testCase.expectedLabels, labels)
		})
	}

}

func TestParseQueryStringToInt64(t *testing.T) {
	startPathParam := "1720241589000000010"
	expectedStart := int64(1720241589000000010)

	tests := []struct {
		name              string
		start             string
		defaultValue      int64
		minStart          int64
		maxStart          int64
		expectedStart     int64
		expectedErrorKind errors.ErrKind
	}{
		{"valid", startPathParam, 0, 0, math.MaxInt64, expectedStart, ""},
		{"valid - defaultValue exceeds the maximum ", startPathParam, 1720241589000000025, 172024158900000005, 1720241589000000015, expectedStart, ""},
		{"invalid - minimum exceeds the maximum ", startPathParam, 0, 1720241589000000025, 1720241589000000015, expectedStart, errors.KindContractInvalid},
		{"invalid - path param not integer", "invalidStart", 0, 0, math.MaxInt64, expectedStart, errors.KindContractInvalid},
		{"invalid - parsed result exceeds the maximum", "1753601008000000000", 0, 172024158900000005, 1720241589000000015, expectedStart, errors.KindContractInvalid},
		{"invalid - parsed result less than the minimum", "171024158900000005", 0, 172024158900000005, 1720241589000000015, expectedStart, errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Run(testCase.name, func(t *testing.T) {
				e := echo.New()
				req, err := http.NewRequest(http.MethodGet, common.ApiAllEventRoute, http.NoBody)
				require.NoError(t, err)
				query := req.URL.Query()
				if testCase.start != "" {
					query.Add(common.Start, testCase.start)
				}
				req.URL.RawQuery = query.Encode()

				c := e.NewContext(req, nil)
				start, err := ParseQueryStringToInt64(c, common.Start, testCase.defaultValue, testCase.minStart, testCase.maxStart)
				if testCase.expectedErrorKind != "" {
					assert.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
					return
				}
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedStart, start)
			})
		})
	}
}

func TestParsePathParamToInt64(t *testing.T) {
	startPathParam := "1720241589000000010"
	expectedStart := int64(1720241589000000010)

	tests := []struct {
		name              string
		start             string
		minStart          int64
		maxStart          int64
		expectedStart     int64
		expectedErrorKind errors.ErrKind
	}{
		{"valid", startPathParam, 0, math.MaxInt64, expectedStart, ""},
		{"invalid", "", 0, math.MaxInt64, expectedStart, errors.KindContractInvalid},
		{"invalid - minimum exceeds the maximum ", startPathParam, 1720241589000000025, 1720241589000000015, expectedStart, errors.KindContractInvalid},
		{"invalid - path param not integer", "invalidStart", 0, math.MaxInt64, expectedStart, errors.KindContractInvalid},
		{"invalid - parsed result exceeds the maximum", "1753601008000000000", 172024158900000005, 1720241589000000015, expectedStart, errors.KindContractInvalid},
		{"invalid - parsed result less than the minimum", "171024158900000005", 172024158900000005, 1720241589000000015, expectedStart, errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Run(testCase.name, func(t *testing.T) {
				e := echo.New()
				req, err := http.NewRequest(http.MethodGet, common.ApiReadingByTimeRangeRoute, http.NoBody)
				require.NoError(t, err)

				// Act
				recorder := httptest.NewRecorder()
				c := e.NewContext(req, recorder)
				c.SetParamNames(common.Start)
				c.SetParamValues(testCase.start)
				start, err := ParsePathParamToInt64(c, common.Start, testCase.minStart, testCase.maxStart)
				if testCase.expectedErrorKind != "" {
					assert.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
					return
				}
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedStart, start)
			})
		})
	}
}

func TestParseAggregateFuncQueryString(t *testing.T) {
	tests := []struct {
		name              string
		aggFunc           string
		expectedResult    string
		expectedErrorKind errors.ErrKind
	}{
		{"valid", "min", "MIN", ""},
		{"invalid", "unknown", "", errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Run(testCase.name, func(t *testing.T) {
				e := echo.New()
				req, err := http.NewRequest(http.MethodGet, common.ApiAllReadingRoute, http.NoBody)
				require.NoError(t, err)

				query := req.URL.Query()
				query.Add(common.AggregateFunc, testCase.aggFunc)
				req.URL.RawQuery = query.Encode()

				// Act
				recorder := httptest.NewRecorder()
				c := e.NewContext(req, recorder)
				aggFunc := c.QueryParam(common.AggregateFunc)
				start, err := ParseAggregateFuncQueryString(aggFunc)
				if testCase.expectedErrorKind != "" {
					assert.Equal(t, testCase.expectedErrorKind, errors.Kind(err))
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResult, start)
			})
		})
	}
}
