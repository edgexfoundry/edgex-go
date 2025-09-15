//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"math/big"
	"reflect"
	"strconv"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	contractModels "github.com/edgexfoundry/go-mod-core-contracts/v4/models"

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
		name              string
		valueType         string
		aggregateFunc     string
		numericValue      *pgtype.Numeric
		expectedValue     any
		expectedValueType string
		expectError       bool
	}{
		{
			name:              "CountFunc - valid input",
			valueType:         common.ValueTypeInt64,
			aggregateFunc:     common.CountFunc,
			numericValue:      createNumeric("100", true),
			expectedValue:     uint64(100),
			expectedValueType: common.ValueTypeUint64,
			expectError:       false,
		},
		{
			name:              "AvgFunc - float value",
			valueType:         common.ValueTypeFloat64,
			aggregateFunc:     common.AvgFunc,
			numericValue:      createNumeric("123.456", false),
			expectedValue:     float64(123.456),
			expectedValueType: common.ValueTypeFloat64,
			expectError:       false,
		},
		{
			name:              "AvgFunc - integer value (should convert to float)",
			valueType:         common.ValueTypeInt64,
			aggregateFunc:     common.AvgFunc,
			numericValue:      createNumeric("50", false),
			expectedValue:     float64(50),
			expectedValueType: common.ValueTypeFloat64,
			expectError:       false,
		},
		{
			name:              "SumFunc - Float64",
			valueType:         common.ValueTypeFloat64,
			aggregateFunc:     common.SumFunc,
			numericValue:      createNumeric("789.012", false),
			expectedValue:     float64(789.012),
			expectedValueType: common.ValueTypeFloat64,
			expectError:       false,
		},
		{
			name:              "MinFunc - Float32",
			valueType:         common.ValueTypeFloat32,
			aggregateFunc:     common.MinFunc,
			numericValue:      createNumeric("1.23", false),
			expectedValue:     float64(1.23),
			expectedValueType: common.ValueTypeFloat64,
			expectError:       false,
		},
		{
			name:              "MaxFunc - Int64",
			valueType:         common.ValueTypeInt64,
			aggregateFunc:     common.MaxFunc,
			numericValue:      createNumeric("9999999999", false),
			expectedValue:     int64(9999999999),
			expectedValueType: common.ValueTypeInt64,
			expectError:       false,
		},
		{
			name:              "SumFunc - Int16",
			valueType:         common.ValueTypeInt16,
			aggregateFunc:     common.SumFunc,
			numericValue:      createNumeric("1500", true),
			expectedValue:     int64(1500),
			expectedValueType: common.ValueTypeInt64,
			expectError:       false,
		},
		{
			name:              "MinFunc - Uint64",
			valueType:         common.ValueTypeUint64,
			aggregateFunc:     common.MinFunc,
			numericValue:      createNumeric("18446744073709551615", false),
			expectedValue:     uint64(18446744073709551615),
			expectedValueType: common.ValueTypeUint64,
			expectError:       false,
		},
		{
			name:              "MaxFunc - Uint8",
			valueType:         common.ValueTypeUint8,
			aggregateFunc:     common.MaxFunc,
			numericValue:      createNumeric("250", true),
			expectedValue:     uint64(250),
			expectedValueType: common.ValueTypeUint64,
			expectError:       false,
		},
		{
			name:              "SumFunc - unsupported valueType",
			valueType:         "invalidType",
			aggregateFunc:     common.SumFunc,
			numericValue:      createNumeric("10", false),
			expectedValue:     nil,
			expectedValueType: "",
			expectError:       true,
		},
		{
			name:              "Default case - unexpected aggregateFunc",
			valueType:         common.ValueTypeInt64,
			aggregateFunc:     "InvalidFunc",
			numericValue:      createNumeric("1", false),
			expectedValue:     nil,
			expectedValueType: "",
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := contractModels.NumericReading{
				BaseReading: contractModels.BaseReading{ValueType: tt.valueType},
			}
			err := parseAggNumericReading(&reading, tt.aggregateFunc, tt.numericValue)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one: %v", err)
				// Use reflect.TypeOf to compare the type of the returned value
				assert.Equal(t, reflect.TypeOf(tt.expectedValue), reflect.TypeOf(reading.NumericValue), "Returned value has unexpected type")
				assert.Equal(t, tt.expectedValue, reading.NumericValue, "Returned value is not as expected")
				assert.Equal(t, tt.expectedValueType, reading.ValueType, "Returned value type is not as expected")
			}
		})
	}
}

func TestNumericToUint64(t *testing.T) {
	tests := []struct {
		name           string
		numeric        *pgtype.Numeric
		expectedResult uint64
		expectedErr    bool
	}{
		{
			name: "Valid - simple integer",
			numeric: &pgtype.Numeric{
				Int:   big.NewInt(123),
				Exp:   0,
				Valid: true,
			},
			expectedResult: 123,
			expectedErr:    false,
		},
		{
			name: "Valid - positive exponent (scale up)",
			numeric: &pgtype.Numeric{
				Int:   big.NewInt(123),
				Exp:   3, // 123 * 10^3 = 123000
				Valid: true,
			},
			expectedResult: 123000,
			expectedErr:    false,
		},
		{
			name: "Valid - negative exponent with exact division (no fractional part)",
			numeric: &pgtype.Numeric{
				Int:   big.NewInt(1234500),
				Exp:   -2, // 1234500 / 10^2 = 12345 (exact)
				Valid: true,
			},
			expectedResult: 12345,
			expectedErr:    false,
		},
		{
			name: "Invalid - negative exponent with fractional part",
			numeric: &pgtype.Numeric{
				Int:   big.NewInt(12345),
				Exp:   -2, // 12345 / 100 leaves remainder
				Valid: true,
			},
			expectedErr: true,
		},
		{
			name: "Invalid - not a finite valid value",
			numeric: &pgtype.Numeric{
				Int:   big.NewInt(0),
				Exp:   0,
				Valid: false,
			},
			expectedErr: true,
		},
		{
			name: "Invalid - NaN",
			numeric: &pgtype.Numeric{
				Int:   big.NewInt(0),
				Exp:   0,
				Valid: true,
				NaN:   true,
			},
			expectedErr: true,
		},
		{
			name: "Invalid - Infinity",
			numeric: &pgtype.Numeric{
				Int:              big.NewInt(0),
				Exp:              0,
				Valid:            true,
				InfinityModifier: pgtype.Infinity,
			},
			expectedErr: true,
		},
		{
			name: "Valid - uint64 max boundary",
			numeric: &pgtype.Numeric{
				Int:   new(big.Int).SetUint64(^uint64(0)), // math.MaxUint64
				Exp:   0,
				Valid: true,
			},
			expectedResult: ^uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := numericToUint64(tt.numeric)
			if tt.expectedErr {
				assert.Error(t, err, "Expected an error but result none")
			} else {
				assert.Equal(t, tt.expectedResult, result, "Returned value is not as expected")
			}
		})
	}
}
