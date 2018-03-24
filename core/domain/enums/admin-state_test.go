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

func TestStringAdminState(t *testing.T) {
	tests := []struct {
		name string
		as   AdminStateType
	}{
		{"locked", LOCKED},
		{"unlocked", UNLOCKED},
		{"invalid", UNLOCKED + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.as.String() == "" {
				t.Errorf("String should not be empty")
			}
		})
	}
}
