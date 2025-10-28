//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCheckCountRange(t *testing.T) {
	count := int64(1)
	tests := []struct {
		name         string
		totalCount   int64
		offset       int
		limit        int
		continueExec bool
		expectErr    bool
	}{
		{"valid - total count is zero ", int64(0), 0, 0, false, false},
		{"valid - limit is zero ", count, 0, 0, false, false},
		{"valid - valid range", count, 0, 1, true, false},
		{"invalid - offset out of range", count, 2, 1, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cont, err := CheckCountRange(tt.totalCount, tt.offset, tt.limit)
			require.Equal(t, tt.continueExec, cont)
			if tt.continueExec {
				require.NoError(t, err)
			}
			if tt.expectErr {
				require.Error(t, err)
			}
		})
	}
}
