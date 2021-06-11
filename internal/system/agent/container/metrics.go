//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
)

// V2MetricsInterfaceName contains the name of the interfaces.Metrics implementation in the DIC.
var V2MetricsInterfaceName = di.TypeInstanceToName((*interfaces.Metrics)(nil))

// V2MetricsFrom helper function queries the DIC and returns the interfaces.Metrics implementation.
func V2MetricsFrom(get di.Get) interfaces.Metrics {
	return get(V2MetricsInterfaceName).(interfaces.Metrics)
}
