//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"math"
	"net/http"
	"strconv"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"

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
		{"valid - offset and limit is empty", "", "", testLabel, 50, v2.DefaultOffset, v2.DefaultLimit, []string{testLabel}, ""},
		{"valid - limit is empty and default value is greater then the maximum", "", "", testLabel, 5, v2.DefaultOffset, 5, []string{testLabel}, ""},
		{"invalid - offset exceeds the minimum ", "-1", "2", testLabel, 50, 0, 2, []string{testLabel}, errors.KindContractInvalid},
		{"invalid - offset exceeds the maximum ", strconv.Itoa(math.MaxUint32), "2", testLabel, 10, 0, 2, []string{testLabel}, errors.KindContractInvalid},
		{"invalid - limit exceeds the minimum ", "0", "-2", testLabel, 50, 0, 2, []string{testLabel}, errors.KindContractInvalid},
		{"invalid - limit exceeds the maximum ", "0", "100", testLabel, 50, 0, 2, []string{testLabel}, errors.KindContractInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, v2.ApiAllEventRoute, http.NoBody)
			require.NoError(t, err)
			query := req.URL.Query()
			if testCase.offset != "" {
				query.Add(v2.Offset, testCase.offset)
			}
			if testCase.limit != "" {
				query.Add(v2.Limit, testCase.limit)
			}
			if testCase.labels != "" {
				query.Add(v2.Labels, testCase.labels)
			}
			req.URL.RawQuery = query.Encode()

			offset, limit, labels, err := ParseGetAllObjectsRequestQueryString(req, 0, math.MaxInt32, -1, testCase.maxLimit)
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
