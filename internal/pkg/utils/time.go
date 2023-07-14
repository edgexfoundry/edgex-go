//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

// CheckMinInterval parses the ISO 8601 time duration string to Duration type
// and evaluates if the duration value is smaller than the suggested minimum duration
func CheckMinInterval(value string, minDuration time.Duration, lc logger.LoggingClient) {
	valueDuration, err := time.ParseDuration(value)
	if err != nil {
		lc.Errorf("failed to parse the interval duration string %s to a duration time value: %v", value, err)
		return
	}

	if valueDuration < minDuration {
		// the duration value is smaller than the min
		lc.Warnf("the interval value '%s' is smaller than the suggested value '%s', which might cause abnormal CPU increase", value, minDuration)
	}
}
