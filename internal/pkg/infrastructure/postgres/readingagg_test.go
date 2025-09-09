//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestParseAggNumericReading(t *testing.T) {
	createNumeric := func(val string, isInt64 bool) *pgtype.Numeric {
		n := &pgtype.Numeric{}
		if isInt64 {
			int64Num, _ := strconv.ParseInt(val, 10, 64)
			_ = n.ScanInt64(pgtype.Int8{Int64: int64Num, Valid: true})
		} else {
			_ = n.Scan(val)
		}

		return n
	}

	tests := []struct {
		name          string
		valueType     string
		aggregateFunc string
		numericValue  *pgtype.Numeric
		expectedValue any
		expectError   bool
	}{
		{
			name:          "CountFunc - valid input",
			valueType:     common.ValueTypeInt64,
			aggregateFunc: common.CountFunc,
			numericValue:  createNumeric("100", true),
			expectedValue: uint64(100),
			expectError:   false,
		},
		{
			name:          "AvgFunc - float value",
			valueType:     common.ValueTypeFloat64,
			aggregateFunc: common.AvgFunc,
			numericValue:  createNumeric("123.456", false),
			expectedValue: float64(123.456),
			expectError:   false,
		},
		{
			name:          "AvgFunc - integer value (should convert to float)",
			valueType:     common.ValueTypeInt64,
			aggregateFunc: common.AvgFunc,
			numericValue:  createNumeric("50", false),
			expectedValue: float64(50),
			expectError:   false,
		},
		{
			name:          "SumFunc - Float64",
			valueType:     common.ValueTypeFloat64,
			aggregateFunc: common.SumFunc,
			numericValue:  createNumeric("789.012", false),
			expectedValue: float64(789.012),
			expectError:   false,
		},
		{
			name:          "MinFunc - Float32",
			valueType:     common.ValueTypeFloat32,
			aggregateFunc: common.MinFunc,
			numericValue:  createNumeric("1.23", false),
			expectedValue: float64(1.23),
			expectError:   false,
		},
		{
			name:          "MaxFunc - Int64",
			valueType:     common.ValueTypeInt64,
			aggregateFunc: common.MaxFunc,
			numericValue:  createNumeric("9999999999", false),
			expectedValue: int64(9999999999),
			expectError:   false,
		},
		{
			name:          "SumFunc - Int16",
			valueType:     common.ValueTypeInt16,
			aggregateFunc: common.SumFunc,
			numericValue:  createNumeric("1500", true),
			expectedValue: int64(1500),
			expectError:   false,
		},
		{
			name:          "MinFunc - Uint64",
			valueType:     common.ValueTypeUint64,
			aggregateFunc: common.MinFunc,
			numericValue:  createNumeric("18446744073709551615", false),
			expectedValue: uint64(18446744073709551615),
			expectError:   false,
		},
		{
			name:          "MaxFunc - Uint8",
			valueType:     common.ValueTypeUint8,
			aggregateFunc: common.MaxFunc,
			numericValue:  createNumeric("250", true),
			expectedValue: uint64(250),
			expectError:   false,
		},
		{
			name:          "SumFunc - unsupported valueType",
			valueType:     "invalidType",
			aggregateFunc: common.SumFunc,
			numericValue:  createNumeric("10", false),
			expectedValue: nil,
			expectError:   true,
		},
		{
			name:          "Default case - unexpected aggregateFunc",
			valueType:     common.ValueTypeInt64,
			aggregateFunc: "InvalidFunc",
			numericValue:  createNumeric("1", false),
			expectedValue: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAggNumericReading(tt.valueType, tt.aggregateFunc, tt.numericValue)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one: %v", err)
				// Use reflect.TypeOf to compare the type of the returned value
				assert.Equal(t, reflect.TypeOf(tt.expectedValue), reflect.TypeOf(result), "Returned value has unexpected type")
				assert.Equal(t, tt.expectedValue, result, "Returned value is not as expected")
			}
		})
	}
}
