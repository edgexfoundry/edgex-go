//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger/mocks"
)

func TestCheckMinInterval(t *testing.T) {
	validInterval := "1s"
	lessThanMinInterval := "100us"
	invalidInterval := "INVALID"
	minInterval1 := 10 * time.Millisecond
	minInterval2 := 1 * time.Millisecond

	warnMsg := "the interval value '%s' is smaller than the suggested value '%s', which might cause abnormal CPU increase"
	errMsg := "failed to parse the interval duration string %s to a duration time value: %v"
	expectedErr := errors.New("time: invalid duration \"" + invalidInterval + "\"")

	lcMock := &mocks.LoggingClient{}
	lcMock.On("Warnf", warnMsg, validInterval, minInterval1)
	lcMock.On("Warnf", warnMsg, lessThanMinInterval, minInterval2)
	lcMock.On("Errorf", errMsg, invalidInterval, expectedErr)

	tests := []struct {
		name        string
		interval    string
		min         time.Duration
		logExpected bool
		logLevel    string
		logMsg      string
		err         error
	}{
		{"valid - interval is bigger than the minimum value", validInterval, minInterval1, false, "", "", nil},
		{"invalid - interval is smaller than the minimum value", lessThanMinInterval, minInterval2, true, "Warnf", warnMsg, nil},
		{"invalid - parsing duration string failed", invalidInterval, minInterval2, true, "Errorf", errMsg, expectedErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CheckMinInterval(tt.interval, tt.min, lcMock)
			if tt.logExpected {
				if tt.logLevel == "Warnf" {
					lcMock.AssertCalled(t, tt.logLevel, tt.logMsg, tt.interval, tt.min)
				} else {
					lcMock.AssertCalled(t, tt.logLevel, tt.logMsg, tt.interval, tt.err)
				}
			} else {
				lcMock.AssertNotCalled(t, tt.logLevel, tt.logMsg, tt.interval, tt.min)
			}
		})
	}
}
