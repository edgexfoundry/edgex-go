//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"math/big"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNumericReadingVal(t *testing.T) {
	testInt64Power10 := pgtype.Numeric{
		Int:   big.NewInt(1),
		Exp:   18,
		Valid: true,
	}
	testInt64 := pgtype.Numeric{
		Int:   big.NewInt(9223372036854775807),
		Exp:   0,
		Valid: true,
	}
	testUint64Power := pgtype.Numeric{
		Int:   big.NewInt(1),
		Exp:   19,
		Valid: true,
	}
	testUint64 := pgtype.Numeric{
		Int:   new(big.Int).SetUint64(^uint64(0)), //  ^uint64(0) represents the max uint 64 value 18446744073709551615
		Exp:   0,
		Valid: true,
	}

	testInt32 := pgtype.Numeric{
		Int:   big.NewInt(-2147483648),
		Exp:   0,
		Valid: true,
	}
	testUint32 := pgtype.Numeric{
		Int:   big.NewInt(4294967295),
		Exp:   0,
		Valid: true,
	}
	testInt16 := pgtype.Numeric{
		Int:   big.NewInt(-32768),
		Exp:   0,
		Valid: true,
	}
	testUint16 := pgtype.Numeric{
		Int:   big.NewInt(65535),
		Exp:   0,
		Valid: true,
	}
	testInt8 := pgtype.Numeric{
		Int:   big.NewInt(-128),
		Exp:   0,
		Valid: true,
	}
	testUint8 := pgtype.Numeric{
		Int:   big.NewInt(255),
		Exp:   0,
		Valid: true,
	}
	testFloat64 := pgtype.Numeric{
		Int:   big.NewInt(1), // approximately equals to 2^53
		Exp:   -16383,
		Valid: true,
	}
	testFloat32 := pgtype.Numeric{
		Int:   big.NewInt(14), // approximately equals to 2^53
		Exp:   -46,
		Valid: true,
	}

	tests := []struct {
		name          string
		valueType     string
		numericValue  *pgtype.Numeric
		expectedValue any
		expectError   bool
	}{
		{
			name:          "Valid - Test number is power of 10 to int64",
			valueType:     common.ValueTypeInt64,
			numericValue:  &testInt64Power10,
			expectedValue: int64(1e18),
			expectError:   false,
		},
		{
			name:          "Valid - Test number to int64",
			valueType:     common.ValueTypeInt64,
			numericValue:  &testInt64,
			expectedValue: int64(9223372036854775807),
			expectError:   false,
		},
		{
			name:          "Valid - Test number is power of 10 to Uint64",
			valueType:     common.ValueTypeUint64,
			numericValue:  &testUint64Power,
			expectedValue: uint64(1e19),
			expectError:   false,
		},
		{
			name:          "Valid - Test number to Uint64",
			valueType:     common.ValueTypeUint64,
			numericValue:  &testUint64,
			expectedValue: ^uint64(0),
			expectError:   false,
		},

		{
			name:          "Valid - Test number to Int32",
			valueType:     common.ValueTypeInt32,
			numericValue:  &testInt32,
			expectedValue: int64(-2147483648),
			expectError:   false,
		},
		{
			name:          "Valid - Test number to Uint32",
			valueType:     common.ValueTypeUint32,
			numericValue:  &testUint32,
			expectedValue: uint64(4294967295),
			expectError:   false,
		},
		{
			name:          "Valid - Test number to Int16",
			valueType:     common.ValueTypeInt16,
			numericValue:  &testInt16,
			expectedValue: int64(-32768),
			expectError:   false,
		},
		{
			name:          "Valid - Test number to Uint16",
			valueType:     common.ValueTypeUint16,
			numericValue:  &testUint16,
			expectedValue: uint64(65535),
			expectError:   false,
		},
		{
			name:          "Valid - Test number to Int8",
			valueType:     common.ValueTypeInt8,
			numericValue:  &testInt8,
			expectedValue: int64(-128),
			expectError:   false,
		},
		{
			name:          "Valid - Test number to Uint8",
			valueType:     common.ValueTypeUint8,
			numericValue:  &testUint8,
			expectedValue: uint64(255),
			expectError:   false,
		},
		{
			name:          "Valid - Test number to Float64",
			valueType:     common.ValueTypeFloat64,
			numericValue:  &testFloat64,
			expectedValue: 1e-16383,
			expectError:   false,
		},
		{
			name:          "Valid - Test number to Float32",
			valueType:     common.ValueTypeFloat32,
			numericValue:  &testFloat32,
			expectedValue: 1.4e-45,
			expectError:   false,
		},
		{
			name:          "Invalid - Unknown value type",
			valueType:     "UNKNOWN",
			numericValue:  &testFloat32,
			expectedValue: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := numericReadingVal(tt.valueType, tt.numericValue)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result, "Returned value is not as expected")
			}
		})
	}
}
