//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common_config

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/environment"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// getOverwriteConfig returns a boolean value based on whether the -o flag or the EDGEX_OVERWRITE_CONFIG environment variable is set
func getOverwriteConfig(f *flags.Default, lc logger.LoggingClient) bool {
	overwrite := f.OverwriteConfig()

	if b, ok := environment.OverwriteConfig(lc); ok {
		overwrite = b
	}

	return overwrite
}
