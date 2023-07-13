//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestCheckMinInterval(t *testing.T) {
	lc := logger.NewMockClient()

	tests := []struct {
		name     string
		interval string
		min      time.Duration
		result   bool
	}{
		{"valid - interval is bigger than the minimum value", "1s", 10 * time.Millisecond, true},
		{"invalid - interval is smaller than the minimum value", "100us", 1 * time.Millisecond, false},
		{"invalid - parsing duration string failed", "INVALID", 1 * time.Millisecond, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckMinInterval(tt.interval, tt.min, lc)
			assert.Equal(t, tt.result, result)
		})
	}
}
