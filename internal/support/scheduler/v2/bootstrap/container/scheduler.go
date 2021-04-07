//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/v2/infrastructure/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// SchedulerClientInterfaceName contains the name of the interfaces.SchedulerClient implementation in the DIC.
var SchedulerClientInterfaceName = di.TypeInstanceToName((*interfaces.SchedulerClient)(nil))

// SchedulerClientFrom helper function queries the DIC and returns the interfaces.SchedulerClient implementation.
func SchedulerClientFrom(get di.Get) interfaces.SchedulerClient {
	return get(SchedulerClientInterfaceName).(interfaces.SchedulerClient)
}
