//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package support_domain

import (
	"testing"
)

func TestIsValidLogLevel(t *testing.T) {
	var tests = []struct {
		level string
		res   bool
	}{
		{TRACE, true},
		{DEBUG, true},
		{INFO, true},
		{WARN, true},
		{ERROR, true},
		{"EERROR", false},
		{"ERRORR", false},
		{"INF", false},
	}
	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			r := IsValidLogLevel(tt.level)
			if r != tt.res {
				t.Errorf("Level %s labeled as %v and should be %v",
					tt.level, r, tt.res)
			}
		})
	}
}
