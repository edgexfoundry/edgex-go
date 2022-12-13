//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

// DBClientInterfaceName contains the name of the interfaces.DBClient implementation in the DIC.
var DBClientInterfaceName = di.TypeInstanceToName((*interfaces.DBClient)(nil))

// DBClientFrom helper function queries the DIC and returns the interfaces.DBClient implementation.
func DBClientFrom(get di.Get) interfaces.DBClient {
	return get(DBClientInterfaceName).(interfaces.DBClient)
}
