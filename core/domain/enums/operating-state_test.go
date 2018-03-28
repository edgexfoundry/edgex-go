//
// Copyright (c) 2018
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//
package enums

import (
	"testing"
)

func TestStringOperatingState(t *testing.T) {
	tests := []struct {
		name string
		os   OperatingStateType
	}{
		{"enabled", ENABLED},
		{"disabled", DISABLED},
		{"invalid", DISABLED + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.os.String() == "" {
				t.Errorf("String should not be empty")
			}
		})
	}
}
