//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/infrastructure/interfaces"
)

// UnitsOfMeasureInterfaceName contains the name of the interfaces.UnitsOfMeasure implementation in the DIC.
var UnitsOfMeasureInterfaceName = di.TypeInstanceToName((*interfaces.UnitsOfMeasure)(nil))

// UnitsOfMeasureFrom helper function queries the DIC and returns the interfaces.UnitsOfMeasure implementation.
func UnitsOfMeasureFrom(get di.Get) interfaces.UnitsOfMeasure {
	return get(UnitsOfMeasureInterfaceName).(interfaces.UnitsOfMeasure)
}
