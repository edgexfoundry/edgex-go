//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	// TODO: Import the config package from the support/cronscheduler directory if available
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/config"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

// ConfigurationName contains the name of scheduler's config.ConfigurationStruct implementation in the DIC.
var ConfigurationName = di.TypeInstanceToName(config.ConfigurationStruct{})

// ConfigurationFrom helper function queries the DIC and returns scheduler's config.ConfigurationStruct implementation.
func ConfigurationFrom(get di.Get) *config.ConfigurationStruct {
	return get(ConfigurationName).(*config.ConfigurationStruct)
}
